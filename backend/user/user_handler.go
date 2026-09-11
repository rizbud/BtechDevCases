package user

import (
	"btech-wallet/middleware"
	"btech-wallet/server"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	Service *UserService
}

func NewUserHandler(pool *pgxpool.Pool) *UserHandler {
	repository := &UserRepository{
		DB: pool,
	}
	service := &UserService{
		Repository: repository,
	}
	return &UserHandler{
		Service: service,
	}
}

func (h *UserHandler) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	response, err := h.Service.GetUserByID(r.Context(), userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "User with ID "+userID+" not found" {
			statusCode = http.StatusUnauthorized
		}
		server.ErrorResponseJSON(w, statusCode, "Failed to get user profile", nil)
		return
	}

	server.JSON(w, http.StatusOK, response)
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /profile", middleware.AuthMiddleware(http.HandlerFunc(h.handleGetProfile)))
}
