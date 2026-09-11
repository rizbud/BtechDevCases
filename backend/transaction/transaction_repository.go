package transaction

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	DB *pgxpool.Pool
}

type Transaction struct {
	ID         string  `json:"id"`
	FromUserID string  `json:"from_user_id"`
	ToUserID   string  `json:"to_user_id"`
	Amount     float64 `json:"amount"`
	CreatedAt  string  `json:"created_at"`
}

func (r *TransactionRepository) CreateTransaction(
	ctx context.Context,
	fromUserID string,
	toUserID string,
	amount float64,
) (string, error) {
	var transactionID string

	err := r.DB.QueryRow(
		ctx,
		"INSERT INTO transactions (from_user_id, to_user_id, amount) VALUES ($1, $2, $3) RETURNING id",
		fromUserID,
		toUserID,
		amount,
	).Scan(&transactionID)

	if err != nil {
		log.Printf("[TransactionRepository.CreateTransaction] Failed to create transaction: %v", err)
		return "", fmt.Errorf("Failed to create transaction")
	}

	return transactionID, nil
}

func (r *TransactionRepository) GetTransactionsByUserID(
	ctx context.Context,
	userID string,
	startDate string,
	endDate string,
) ([]Transaction, error) {
	rows, err := r.DB.Query(
		ctx,
		`SELECT id, from_user_id, to_user_id, amount, created_at
			FROM transactions
			WHERE (from_user_id = $1 OR to_user_id = $1)
			AND created_at BETWEEN $2 AND $3
			ORDER BY created_at DESC`,
		userID,
		startDate,
		endDate,
	)
	if err != nil {
		log.Printf("[TransactionRepository.GetTransactionsByUserID] Failed to get transactions: %v", err)
		return nil, fmt.Errorf("Failed to get transactions for user %s", userID)
	}

	defer rows.Close()

	var transactions []Transaction

	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.FromUserID,
			&transaction.ToUserID,
			&transaction.Amount,
			&transaction.CreatedAt,
		)
		if err != nil {
			log.Printf("[TransactionRepository.GetTransactionsByUserID] Failed to scan transaction: %v", err)
			return nil, fmt.Errorf("Failed to get transactions for user %s", userID)
		}

		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[TransactionRepository.GetTransactionsByUserID] Rows error: %v", err)
		return nil, fmt.Errorf("Failed to get transactions for user %s", userID)
	}

	return transactions, nil
}

func (r *TransactionRepository) GetUserBalance(ctx context.Context, userID string) (float64, error) {
	var balance float64
	err := r.DB.QueryRow(
		ctx,
		"SELECT balance FROM user_balances WHERE user_id = $1",
		userID,
	).Scan(&balance)

	if err != nil {
		log.Printf("[TransactionRepository.GetUserBalance] Failed to get user balance: %v", err)
		return 0, fmt.Errorf("Failed to get balance for user %s", userID)
	}

	return balance, nil
}
