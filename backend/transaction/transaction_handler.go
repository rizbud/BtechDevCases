package transaction

import (
	"net/http"

	"btech-wallet/middleware"
	"btech-wallet/server"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionHandler struct {
	Service *TransactionService
}

func NewTransactionHandler(pool *pgxpool.Pool) *TransactionHandler {
	repo := &TransactionRepository{
		DB: pool,
	}
	service := &TransactionService{
		Repository: repo,
	}
	return &TransactionHandler{
		Service: service,
	}
}

func (h *TransactionHandler) handleGetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	balance, err := h.Service.getUserBalance(r.Context(), userID)
	if err != nil {
		server.ErrorResponseJSON(w, http.StatusInternalServerError, "Failed to get user balance", err)
		return
	}

	response := map[string]float64{
		"balance": balance,
	}
	server.JSON(w, http.StatusOK, response)
}

func (h *TransactionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /balance", middleware.AuthMiddleware(http.HandlerFunc(h.handleGetUserBalance)))
}
