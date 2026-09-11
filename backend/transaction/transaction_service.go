package transaction

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"
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

func (s *TransactionService) validateTransactionsRequest(
	startDateStr,
	endDateStr,
	pageStr,
	pageSizeStr string,
) (map[string]string, int, int) {
	validationErrors := map[string]string{}
	if startDateStr == "" {
		validationErrors["start_date"] = "Start date is required"
	}
	if endDateStr == "" {
		validationErrors["end_date"] = "End date is required"
	}

	loc := time.Now().Location() // use local timezone for date parsing
	startDate, err := time.ParseInLocation("2006-01-02", startDateStr, loc)
	if err != nil {
		validationErrors["start_date"] = "Invalid start date format. Use YYYY-MM-DD"
	}

	endDate, err := time.ParseInLocation("2006-01-02", endDateStr, loc)
	if err != nil {
		validationErrors["end_date"] = "Invalid end date format. Use YYYY-MM-DD"
	}
	if startDate.After(endDate) {
		validationErrors["date_range"] = "Start date cannot be after end date"
	}

	// max date range of 1 year
	if endDate.Sub(startDate).Hours() > 24*365 {
		validationErrors["date_range"] = "Date range cannot exceed 1 year"
	}

	// end date cannot be in the future
	if endDate.After(time.Now()) {
		validationErrors["end_date"] = "End date cannot be in the future"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		validationErrors["page"] = "Invalid page number"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		validationErrors["page_size"] = "Invalid page size"
	}

	if page <= 0 {
		validationErrors["page"] = "Page must be greater than zero"
	}
	if pageSize <= 0 {
		validationErrors["page_size"] = "Page size must be greater than zero"
	}

	return validationErrors, page, pageSize
}

func (s *TransactionService) getTransactionsByUserID(
	ctx context.Context,
	userID,
	startDate,
	endDate string,
	page,
	pageSize int,
) ([]Transaction, int, error) {
	transactions, totalRecords, err := s.Repository.GetTransactionsByUserID(
		ctx,
		userID,
		startDate,
		endDate,
		page,
		pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	return transactions, totalRecords, nil
}

func (s *TransactionService) getTransactionByID(ctx context.Context, userID, transactionID string) (Transaction, error) {
	transaction, err := s.Repository.GetTransactionByID(ctx, userID, transactionID)
	if err != nil {
		return Transaction{}, err
	}
	return transaction, nil
}
