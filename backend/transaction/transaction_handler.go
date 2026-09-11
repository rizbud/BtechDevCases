package transaction

import (
	"net/http"

	"btech-wallet/middleware"
	"btech-wallet/server"
	"btech-wallet/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionHandler struct {
	TrxService  *TransactionService
	UserService *user.UserService
}

type TransferRequest struct {
	ToUserEmail string  `json:"to_user_email"`
	Amount      float64 `json:"amount"`
}

type TopUpRequest struct {
	Amount float64 `json:"amount"`
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
		TrxService:  trxService,
		UserService: userService,
	}
}

func (h *TransactionHandler) handleGetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	balance, err := h.TrxService.getUserBalance(r.Context(), userID)
	if err != nil {
		server.ErrorResponseJSON(w, http.StatusInternalServerError, "Failed to get user balance", err)
		return
	}

	response := map[string]float64{
		"balance": balance,
	}
	server.JSON(w, http.StatusOK, response)
}

func (h *TransactionHandler) handleTransfer(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req TransferRequest
	if err := server.ValidateBodyRequest(r, &req); err != nil {
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Invalid request payload", err.Error())
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

	response, err := h.TrxService.transfer(r.Context(), userID, destinationUser.ID, req.Amount)
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

	server.JSON(w, http.StatusOK, response)
}

func (h *TransactionHandler) handleTopUp(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req TopUpRequest
	if err := server.ValidateBodyRequest(r, &req); err != nil {
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	response, err := h.TrxService.topUp(r.Context(), userID, req.Amount)
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

	server.JSON(w, http.StatusOK, response)
}

func (h *TransactionHandler) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
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

	validationErrors, page, pageSize := h.TrxService.validateTransactionsRequest(
		startDate,
		endDate,
		pageStr,
		pageSizeStr,
	)
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

	response := server.PaginationResponse[Transaction]{
		Data:         transactions,
		TotalRecords: totalRecords,
		TotalPages:   (totalRecords + pageSize - 1) / pageSize,
		CurrentPage:  page,
		PageSize:     pageSize,
	}
	server.JSON(w, http.StatusOK, response)
}

func (h *TransactionHandler) handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
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

	server.JSON(w, http.StatusOK, transaction)
}

func (h *TransactionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /wallet/balance", middleware.AuthMiddleware(http.HandlerFunc(h.handleGetUserBalance)))
	mux.Handle("POST /wallet/transfer", middleware.AuthMiddleware(http.HandlerFunc(h.handleTransfer)))
	mux.Handle("POST /wallet/topup", middleware.AuthMiddleware(http.HandlerFunc(h.handleTopUp)))
	mux.Handle(
		"GET /wallet/transactions",
		middleware.AuthMiddleware(http.HandlerFunc(h.handleGetTransactions)),
	)
	mux.Handle(
		"GET /wallet/transaction/{transactionId}",
		middleware.AuthMiddleware(http.HandlerFunc(h.handleGetTransaction)),
	)
}
