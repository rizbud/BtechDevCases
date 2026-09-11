package auth

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"btech-wallet/user"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repository *user.UserRepository
	JWTManager *JWTManager
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

	return s.Repository.CreateUser(ctx, email, string(hashed_password))
}

func (s *AuthService) login(ctx context.Context, email string, password string) (LoginResponse, error) {
	user, err := s.Repository.GetUserByEmail(ctx, email)
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

	jwt, err := s.JWTManager.Issue(user.ID, user.Email)
	if err != nil {
		log.Printf("[AuthService.login] Failed to issue JWT: %v", err)
		return LoginResponse{}, err
	}

	userData := LoginResponse{
		ID:        user.ID,
		Email:     user.Email,
		AuthToken: jwt,
	}

	return userData, nil
}
