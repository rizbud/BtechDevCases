package transaction

import (
	"context"
	"fmt"
	"log"
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

func (s *TransactionService) validateTransferRequest(fromUserID, toUserID string, amount float64) map[string]string {
	validationErrors := map[string]string{}

	if fromUserID == "" {
		validationErrors["from_user_id"] = "From user ID is required"
	}
	if toUserID == "" {
		validationErrors["to_user_email"] = "To user Email is required"
	}
	if fromUserID == toUserID {
		validationErrors["to_user_email"] = "Cannot transfer to the same user"
	}
	if amount <= 0 {
		validationErrors["amount"] = "Amount must be greater than zero"
	}
	return validationErrors
}

func (s *TransactionService) transfer(ctx context.Context, fromUserID, toUserID string, amount float64) (Transaction, error) {
	// begin a transaction
	tx, err := s.Repository.DB.Begin(ctx)
	if err != nil {
		log.Printf("[TransactionService.transfer] Failed to begin transaction: %v", err)
		return Transaction{}, fmt.Errorf("Failed to begin transaction: %v", err)
	}

	defer tx.Rollback(ctx) // rollback the transaction in case of an error

	balance, err := s.Repository.GetUserBalance(ctx, fromUserID)
	if err != nil {
		return Transaction{}, err
	}

	if balance < amount {
		return Transaction{}, fmt.Errorf("Insufficient balance")
	}

	// Perform the transfer
	transactionID, err := s.Repository.CreateTransaction(ctx, &fromUserID, toUserID, amount)
	if err != nil {
		return Transaction{}, err
	}

	transaction, err := s.Repository.GetTransactionByID(ctx, fromUserID, transactionID)
	if err != nil {
		return Transaction{}, err
	}

	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("[TransactionService.transfer] Failed to commit transaction: %v", err)
		return Transaction{}, fmt.Errorf("Failed to commit transaction: %v", err)
	}

	return transaction, nil
}

func (s *TransactionService) topUp(ctx context.Context, userID string, amount float64) (Transaction, error) {
	// begin a transaction
	tx, err := s.Repository.DB.Begin(ctx)
	if err != nil {
		log.Printf("[TransactionService.topUp] Failed to begin transaction: %v", err)
		return Transaction{}, fmt.Errorf("Failed to begin transaction: %v", err)
	}

	defer tx.Rollback(ctx) // rollback the transaction in case of an error

	// Perform the top-up
	transactionID, err := s.Repository.CreateTransaction(ctx, nil, userID, amount)
	if err != nil {
		return Transaction{}, err
	}

	transaction, err := s.Repository.GetTransactionByID(ctx, userID, transactionID)
	if err != nil {
		return Transaction{}, err
	}

	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("[TransactionService.topUp] Failed to commit transaction: %v", err)
		return Transaction{}, fmt.Errorf("Failed to commit transaction: %v", err)
	}

	return transaction, nil
}

func (s *TransactionService) getTransactionByID(ctx context.Context, userID, transactionID string) (Transaction, error) {
	transaction, err := s.Repository.GetTransactionByID(ctx, userID, transactionID)
	if err != nil {
		return Transaction{}, err
	}
	return transaction, nil
}
