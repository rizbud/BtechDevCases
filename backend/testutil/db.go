// Package testutil provides shared helpers for integration tests that need
// a real Postgres connection (see docker-compose.yaml).
package testutil

import (
	"context"
	"time"

	"btech-wallet/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a pool against DATABASE_URL and runs migrations. Callers
// should skip their test suite (via TestMain) when it returns an error.
func Connect() (*pgxpool.Pool, error) {
	dbURL := config.LoadEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/btech_wallet_db?sslmode=disable")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := config.InitDB(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return pool, nil
}

// CleanupUserByEmail deletes a test user and everything that references it
// (refresh tokens, transactions) so tests don't leave rows behind.
func CleanupUserByEmail(pool *pgxpool.Pool, email string) {
	ctx := context.Background()

	var id string
	if err := pool.QueryRow(ctx, "SELECT id FROM users WHERE email=$1", email).Scan(&id); err != nil {
		return
	}
	pool.Exec(ctx, "DELETE FROM refresh_tokens WHERE user_id=$1", id)
	pool.Exec(ctx, "DELETE FROM transactions WHERE from_user_id=$1 OR to_user_id=$1", id)
	pool.Exec(ctx, "DELETE FROM users WHERE id=$1", id)
}
