package user

import (
	"context"
	"regexp"
)

type UserService struct {
	Repository *UserRepository
}

func ValidateEmail(email string) bool {
	// Only allow emails with alphanumeric characters, dots, underscores, and hyphens before the @ symbol
	// and a valid domain name after the @ symbol.
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(pattern).MatchString(email)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.Repository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	data := &User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	return data, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID string) (*User, error) {
	user, err := s.Repository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	data := &User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	return data, nil
}
