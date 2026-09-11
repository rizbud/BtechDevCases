package user

import (
	"context"
)

type UserService struct {
	Repository *UserRepository
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
