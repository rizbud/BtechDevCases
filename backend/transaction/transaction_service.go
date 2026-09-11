package transaction

import (
	"context"
)

type TransactionService struct {
	Repository *TransactionRepository
}

func (s *TransactionService) getUserBalance(ctx context.Context, userID string) (float64, error) {
	balance, err := s.Repository.GetUserBalance(ctx, userID)
	if err != nil {
		return 0, err
	}
	return balance, nil
}
