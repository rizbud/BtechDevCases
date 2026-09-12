package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"time"

	"btech-wallet/user"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	AuthRepository *AuthRepository
	UserRepository *user.UserRepository
	JWTManager     *JWTManager
}

func (s *AuthService) validateEmail(email string) bool {
	// Only allow emails with alphanumeric characters, dots, underscores, and hyphens before the @ symbol
	// and a valid domain name after the @ symbol.
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(pattern).MatchString(email)
}

func (s *AuthService) validatePassword(password string) bool {
	return len(password) >= 8
}

func (s *AuthService) validateConfirmPassword(password, confirmPassword string) bool {
	return password == confirmPassword
}

func (s *AuthService) validateLoginRequest(email, password string) map[string]string {
	validationErrors := map[string]string{}

	if !s.validateEmail(email) {
		validationErrors["email"] = "Invalid email format"
	}

	if !s.validatePassword(password) {
		validationErrors["password"] = "Password must be at least 8 characters long"
	}

	return validationErrors
}

func (s *AuthService) validateRegisterRequest(email, password, confirmPassword string) map[string]string {
	validationErrors := s.validateLoginRequest(email, password)

	if !s.validateConfirmPassword(password, confirmPassword) {
		validationErrors["confirmPassword"] = "Passwords do not match"
	}

	return validationErrors
}

func (s *AuthService) register(ctx context.Context, email string, password string) error {
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[AuthService.register] Failed to hash password: %v", err)
		return err
	}

	return s.UserRepository.CreateUser(ctx, email, string(hashed_password))
}

func generateRandomToken() string {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		log.Printf("[generateRandomToken] Failed to generate random token: %v", err)
		return ""
	}

	return hex.EncodeToString(b)
}

func (s *AuthService) issueRefreshToken(ctx context.Context, userID string) (string, error) {
	refreshToken, err := s.AuthRepository.CreateRefreshToken(
		ctx,
		userID,
		generateRandomToken(),
		time.Now().Add(7*24*time.Hour), // Set refresh token expiration to 7 days
	)
	if err != nil {
		log.Printf("[AuthService.issueRefreshToken] Failed to create refresh token: %v", err)
		return "", err
	}

	return *refreshToken, nil
}

func (s *AuthService) login(ctx context.Context, email string, password string) (LoginResponse, error) {
	user, err := s.UserRepository.GetUserByEmail(ctx, email)
	if err != nil {
		if err.Error() == "User with email "+email+" not found" {
			return LoginResponse{}, fmt.Errorf("Email or password is incorrect")
		}
		return LoginResponse{}, err
	}

	// Compare the provided password with the hashed password stored in the database
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return LoginResponse{}, fmt.Errorf("Email or password is incorrect")
	}

	refreshToken, err := s.issueRefreshToken(ctx, user.ID)
	if err != nil {
		log.Printf("[AuthService.login] Failed to issue refresh token: %v", err)
		return LoginResponse{}, err
	}

	jwt, err := s.JWTManager.Issue(user.ID, user.Email)
	if err != nil {
		log.Printf("[AuthService.login] Failed to issue JWT: %v", err)
		return LoginResponse{}, err
	}

	userData := LoginResponse{
		ID:           user.ID,
		Email:        user.Email,
		AuthToken:    jwt,
		RefreshToken: refreshToken,
	}

	return userData, nil
}

func (s *AuthService) refreshToken(ctx context.Context, token string) (RefreshTokenResponse, error) {
	tx, err := s.AuthRepository.DB.Begin(ctx)
	if err != nil {
		log.Printf("[AuthService.refreshToken] Failed to begin transaction: %v", err)
		return RefreshTokenResponse{}, err
	}

	defer tx.Rollback(ctx)

	userID, userEmail, _, err := s.AuthRepository.GetRefreshToken(ctx, token)
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	newAuthToken, err := s.JWTManager.Issue(fmt.Sprintf("%d", userID), *userEmail)
	if err != nil {
		log.Printf("[AuthService.refreshToken] Failed to issue new JWT: %v", err)
		return RefreshTokenResponse{}, err
	}

	newRefreshToken, err := s.issueRefreshToken(ctx, *userID)
	if err != nil {
		log.Printf("[AuthService.refreshToken] Failed to issue new refresh token: %v", err)
		return RefreshTokenResponse{}, err
	}

	tx.Commit(ctx)

	response := RefreshTokenResponse{
		AuthToken:    newAuthToken,
		RefreshToken: newRefreshToken,
	}

	return response, nil
}
