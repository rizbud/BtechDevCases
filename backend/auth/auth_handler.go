package auth

import (
	"net/http"

	"btech-wallet/server"
	"btech-wallet/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type LoginResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	AuthToken    string `json:"auth_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	Token string `json:"token"`
}

type RefreshTokenResponse struct {
	AuthToken    string `json:"auth_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(pool *pgxpool.Pool, jwtManager *JWTManager) *AuthHandler {
	authRepo := &AuthRepository{DB: pool}
	userRepo := &user.UserRepository{DB: pool}
	service := &AuthService{
		AuthRepository: authRepo,
		UserRepository: userRepo,
		JWTManager:     jwtManager,
	}

	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := server.ValidateBodyRequest(r, &req); err != nil {
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	validationErrors := h.service.validateLoginRequest(req.Email, req.Password)
	if len(validationErrors) > 0 {
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Validation errors", validationErrors)
		return
	}

	response, err := h.service.login(r.Context(), req.Email, req.Password)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "Email or password is incorrect" {
			statusCode = http.StatusUnauthorized
		}
		server.ErrorResponseJSON(w, statusCode, err.Error(), nil)
		return
	}

	server.JSON(w, http.StatusOK, response)
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := server.ValidateBodyRequest(r, &req); err != nil {
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	validationErrors := h.service.validateRegisterRequest(req.Email, req.Password, req.ConfirmPassword)
	if len(validationErrors) > 0 {
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Validation errors", validationErrors)
		return
	}

	if err := h.service.register(r.Context(), req.Email, req.Password); err != nil {
		statusCode := http.StatusInternalServerError
		message := "Failed to register user"
		if err.Error() == "User with email "+req.Email+" already exists" {
			statusCode = http.StatusConflict
			message = err.Error()
		}
		server.ErrorResponseJSON(w, statusCode, message, nil)
		return
	}

	response := map[string]string{
		"message": "User registered successfully",
	}

	server.JSON(w, http.StatusOK, response)
}

func (h *AuthHandler) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := server.ValidateBodyRequest(r, &req); err != nil {
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}
	if req.Token == "" {
		validationErrors := map[string]string{"token": "Refresh token is required"}
		server.ErrorResponseJSON(w, http.StatusBadRequest, "Validation errors", validationErrors)
		return
	}

	response, err := h.service.refreshToken(r.Context(), req.Token)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Failed to refresh token"
		if err.Error() == "Refresh token not found or expired" {
			statusCode = http.StatusUnauthorized
			message = err.Error()
		}
		server.ErrorResponseJSON(w, statusCode, message, nil)
		return
	}

	server.JSON(w, http.StatusOK, response)
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/login", h.handleLogin)
	mux.HandleFunc("POST /auth/register", h.handleRegister)
	mux.HandleFunc("POST /auth/refresh-token", h.handleRefreshToken)
}
