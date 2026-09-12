package transaction

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	DB *pgxpool.Pool
}

type dbtx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *TransactionRepository) LockUserForUpdate(ctx context.Context, tx pgx.Tx, userID string) error {
	var id string
	err := tx.QueryRow(ctx, `SELECT id FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("User ID does not exist")
		}
		log.Printf("[TransactionRepository.LockUserForUpdate] Failed to lock user %s: %v", userID, err)
		return fmt.Errorf("Failed to lock user %s", userID)
	}
	return nil
}

type Transaction struct {
	ID             string    `json:"id"`
	SenderID       *string   `json:"sender_id,omitempty"`
	SenderEmail    *string   `json:"sender_email,omitempty"`
	RecipientID    string    `json:"recipient_id"`
	RecipientEmail string    `json:"recipient_email"`
	Amount         float64   `json:"amount"`
	CreatedAt      time.Time `json:"created_at"`
}

func (r *TransactionRepository) CreateTransaction(
	ctx context.Context,
	db dbtx,
	fromUserID *string,
	toUserID string,
	amount float64,
) (string, error) {
	var transactionID string

	err := db.QueryRow(
		ctx,
		`INSERT INTO transactions (from_user_id, to_user_id, amount) VALUES ($1, $2, $3) RETURNING id`,
		fromUserID,
		toUserID,
		amount,
	).Scan(&transactionID)

	if err != nil {
		// if fromUserID or toUserID does not exist, return a specific error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return "", fmt.Errorf("User ID does not exist")
		}
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
	page int,
	pageSize int,
) ([]Transaction, int, error) {
	totalRecords := 0
	err := r.DB.QueryRow(
		ctx,
		`SELECT COUNT(*)
			FROM transactions
			WHERE (from_user_id = $1 OR to_user_id = $1) AND created_at >= $2 AND created_at < $3::date + INTERVAL '1 day'`,
		userID,
		startDate,
		endDate,
	).Scan(&totalRecords)
	if err != nil {
		log.Printf("[TransactionRepository.GetTransactionsByUserID] Failed to count transactions: %v", err)
		return nil, 0, fmt.Errorf("Failed to get transactions for user %s", userID)
	}

	rows, err := r.DB.Query(
		ctx,
		`SELECT
			t.id, COALESCE(fu.id::text, NULL), COALESCE(fu.email, NULL), tu.id::text, tu.email, t.amount, t.created_at
			FROM transactions t
			LEFT JOIN users fu ON t.from_user_id = fu.id
			JOIN users tu ON t.to_user_id = tu.id
			WHERE (t.from_user_id = $1 OR t.to_user_id = $1)
			AND t.created_at >= $2 AND t.created_at < $3::date + INTERVAL '1 day'
			ORDER BY t.created_at DESC
			OFFSET $4 LIMIT $5`,
		userID,
		startDate,
		endDate,
		(page-1)*pageSize,
		pageSize,
	)
	if err != nil {
		// if empty result, return empty slice and totalRecords as 0
		if errors.Is(err, pgx.ErrNoRows) {
			return []Transaction{}, totalRecords, nil
		}
		log.Printf("[TransactionRepository.GetTransactionsByUserID] Failed to get transactions: %v", err)
		return nil, 0, fmt.Errorf("Failed to get transactions for user %s", userID)
	}

	defer rows.Close()

	var transactions []Transaction

	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.SenderID,
			&transaction.SenderEmail,
			&transaction.RecipientID,
			&transaction.RecipientEmail,
			&transaction.Amount,
			&transaction.CreatedAt,
		)
		if err != nil {
			log.Printf("[TransactionRepository.GetTransactionsByUserID] Failed to scan transaction: %v", err)
			return nil, 0, fmt.Errorf("Failed to get transactions for user %s", userID)
		}

		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[TransactionRepository.GetTransactionsByUserID] Rows error: %v", err)
		return nil, 0, fmt.Errorf("Failed to get transactions for user %s", userID)
	}

	return transactions, totalRecords, nil
}

func (r *TransactionRepository) GetTransactionByID(
	ctx context.Context,
	db dbtx,
	userID string,
	transactionID string,
) (Transaction, error) {
	var transaction Transaction
	err := db.QueryRow(
		ctx,
		`SELECT t.id, COALESCE(fu.id::text, NULL), COALESCE(fu.email, NULL), tu.id::text, tu.email, t.amount, t.created_at
			FROM transactions t
			LEFT JOIN users fu ON t.from_user_id = fu.id
			JOIN users tu ON t.to_user_id = tu.id
			WHERE t.id = $1 AND (t.from_user_id = $2 OR t.to_user_id = $2)`,
		transactionID,
		userID,
	).Scan(
		&transaction.ID,
		&transaction.SenderID,
		&transaction.SenderEmail,
		&transaction.RecipientID,
		&transaction.RecipientEmail,
		&transaction.Amount,
		&transaction.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Transaction{}, fmt.Errorf("Transaction ID %s not found", transactionID)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "22P02" {
			// if uuid is invalid, return a specific error
			return Transaction{}, fmt.Errorf("Transaction ID %s not found", transactionID)
		}

		log.Printf("[TransactionRepository.GetTransactionByID] Failed to get transaction: %v", err)
		return Transaction{}, fmt.Errorf("Failed to get transaction with ID %s", transactionID)
	}

	return transaction, nil
}

func (r *TransactionRepository) GetUserBalance(ctx context.Context, db dbtx, userID string) (float64, error) {
	var balance float64
	err := db.QueryRow(
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
