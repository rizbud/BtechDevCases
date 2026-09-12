package transaction

import (
	"maps"
	"net/http"
	"time"

	"btech-wallet/middleware"
	"btech-wallet/server"
	"btech-wallet/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionHandler struct {
	TrxService      *TransactionService
	UserService     *user.UserService
	TransferLimiter *middleware.RateLimiter
	TopUpLimiter    *middleware.RateLimiter
}

type TransferRequest struct {
	ToUserEmail string  `json:"recipient"`
	Amount      float64 `json:"amount"`
	Notes       string  `json:"notes"`
}

type TopUpRequest struct {
	Amount float64 `json:"amount"`
	Notes  string  `json:"notes"`
}

type BalanceResponse struct {
	Message string  `json:"message"`
	Balance float64 `json:"balance"`
}

type TransactionResponse struct {
	Transaction
	Message string `json:"message"`
}

func NewTransactionHandler(pool *pgxpool.Pool) *TransactionHandler {
	trxRepo := &TransactionRepository{
		DB: pool,
	}
	userRepo := &user.UserRepository{
		DB: pool,
	}
	trxService := &TransactionService{
		Repository: trxRepo,
	}
	userService := &user.UserService{
		Repository: userRepo,
	}
	return &TransactionHandler{
		TrxService:      trxService,
		UserService:     userService,
		TransferLimiter: middleware.NewRateLimiter(10, time.Minute),
		TopUpLimiter:    middleware.NewRateLimiter(10, time.Minute),
	}
}

// handleGetUserBalance godoc
// @Summary Get wallet balance
// @Tags wallet
// @Produce json
// @Security BearerAuth
// @Success 200 {object} BalanceResponse
// @Failure 401 {object} server.ErrorResponse
// @Router /wallet/balance [get]
func (h *TransactionHandler) handleGetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	email := r.Context().Value("email").(string)

	balance, err := h.TrxService.getUserBalance(r.Context(), userID)
	if err != nil {
		server.ErrorResponseJSON(w, http.StatusInternalServerError, "Failed to get user balance", err)
		return
	}

	server.JSON(w, http.StatusOK, BalanceResponse{Message: server.WelcomeMessage(email), Balance: balance})
}

// handleTransfer godoc
// @Summary Transfer funds
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body TransferRequest true "Transfer details"
// @Success 200 {object} TransactionResponse
// @Failure 400 {object} server.ErrorResponse
// @Failure 401 {object} server.ErrorResponse
// @Failure 429 {object} server.ErrorResponse
// @Router /wallet/transfer [post]
func (h *TransactionHandler) handleTransfer(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	email := r.Context().Value("email").(string)

	var req TransferRequest
	if err := server.ValidateBodyRequest(r, &req); err != nil {
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	if !user.ValidateEmail(req.ToUserEmail) {
		validationErrors := map[string]string{"recipient": "To user email is not a valid email address"}
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Validation errors", validationErrors)
		return
	}

	destinationUser, err := h.UserService.GetUserByEmail(r.Context(), req.ToUserEmail)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Failed to get destination user"
		if err.Error() == "User with email "+req.ToUserEmail+" not found" {
			statusCode = http.StatusBadRequest
			message = err.Error()
		}
		server.ErrorResponseJSON(w, statusCode, message, nil)
		return
	}

	validationErrors := h.TrxService.validateTransferRequest(userID, destinationUser.ID, req.Amount)
	if len(validationErrors) > 0 {
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Validation errors", validationErrors)
		return
	}

	response, err := h.TrxService.transfer(r.Context(), userID, destinationUser.ID, req.Amount, req.Notes)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Failed to transfer funds"
		if err.Error() == "Insufficient balance" {
			statusCode = http.StatusBadRequest
			message = err.Error()
		} else if err.Error() == "User ID does not exist" {
			statusCode = http.StatusBadRequest
			message = "Target user not found"
		}
		server.ErrorResponseJSON(w, statusCode, message, nil)
		return
	}

	server.JSON(w, http.StatusOK, TransactionResponse{Transaction: response, Message: server.WelcomeMessage(email)})
}

// handleTopUp godoc
// @Summary Top up wallet funds
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body TopUpRequest true "Top-up details"
// @Success 200 {object} TransactionResponse
// @Failure 400 {object} server.ErrorResponse
// @Failure 401 {object} server.ErrorResponse
// @Failure 429 {object} server.ErrorResponse
// @Router /wallet/topup [post]
func (h *TransactionHandler) handleTopUp(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	email := r.Context().Value("email").(string)

	var req TopUpRequest
	if err := server.ValidateBodyRequest(r, &req); err != nil {
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	response, err := h.TrxService.topUp(r.Context(), userID, req.Amount, req.Notes)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Failed to top up funds"
		if err.Error() == "User ID does not exist" {
			statusCode = http.StatusBadRequest
			message = "Target user not found"
		}
		server.ErrorResponseJSON(w, statusCode, message, nil)
		return
	}

	server.JSON(w, http.StatusOK, TransactionResponse{Transaction: response, Message: server.WelcomeMessage(email)})
}

// handleGetTransactions godoc
// @Summary List wallet transactions
// @Tags wallet
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} server.PaginationResponse[Transaction]
// @Failure 400 {object} server.ErrorResponse
// @Failure 401 {object} server.ErrorResponse
// @Router /wallet/transactions [get]
func (h *TransactionHandler) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	email := r.Context().Value("email").(string)
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	pageSizeStr := r.URL.Query().Get("page_size")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}

	page, pageSize, validationErrors := server.ValidatePaginationRequest(r)

	validateTrxErrors := h.TrxService.validateTransactionsRequest(
		startDate,
		endDate,
	)

	maps.Copy(validationErrors, validateTrxErrors)

	if len(validationErrors) > 0 {
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Validation errors", validationErrors)
		return
	}

	transactions, totalRecords, err := h.TrxService.getTransactionsByUserID(
		r.Context(),
		userID,
		startDate,
		endDate,
		page,
		pageSize,
	)
	if err != nil {
		server.ErrorResponseJSON(w, http.StatusInternalServerError, "Failed to get transactions", err)
		return
	}

	totalPages := (totalRecords + pageSize - 1) / pageSize

	server.PaginationResponseJSON(
		w,
		http.StatusOK,
		server.WelcomeMessage(email),
		transactions,
		totalRecords,
		totalPages,
		page,
		pageSize,
	)
}

// handleGetTransaction godoc
// @Summary Get a wallet transaction
// @Tags wallet
// @Produce json
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Success 200 {object} TransactionResponse
// @Failure 401 {object} server.ErrorResponse
// @Failure 404 {object} server.ErrorResponse
// @Router /wallet/transaction/{transactionId} [get]
func (h *TransactionHandler) handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	email := r.Context().Value("email").(string)
	transactionID := r.PathValue("transactionId")

	transaction, err := h.TrxService.getTransactionByID(r.Context(), userID, transactionID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Failed to get transaction"
		if err.Error() == "Transaction ID "+transactionID+" not found" {
			statusCode = http.StatusNotFound
			message = err.Error()
		}
		server.ErrorResponseJSON(w, statusCode, message, nil)
		return
	}

	server.JSON(w, http.StatusOK, TransactionResponse{Transaction: transaction, Message: server.WelcomeMessage(email)})
}

func (h *TransactionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /wallet/balance", middleware.AuthMiddleware(http.HandlerFunc(h.handleGetUserBalance)))
	mux.Handle(
		"POST /wallet/transfer",
		middleware.AuthMiddleware(h.TransferLimiter.Middleware(http.HandlerFunc(h.handleTransfer))),
	)
	mux.Handle(
		"POST /wallet/topup",
		middleware.AuthMiddleware(h.TopUpLimiter.Middleware(http.HandlerFunc(h.handleTopUp))),
	)
	mux.Handle(
		"GET /wallet/transactions",
		middleware.AuthMiddleware(http.HandlerFunc(h.handleGetTransactions)),
	)
	mux.Handle(
		"GET /wallet/transaction/{transactionId}",
		middleware.AuthMiddleware(http.HandlerFunc(h.handleGetTransaction)),
	)
}
