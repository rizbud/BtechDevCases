package auth

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

type AuthRepository struct {
	DB *pgxpool.Pool
}

type dbtx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *AuthRepository) CreateRefreshToken(
	ctx context.Context,
	db dbtx,
	userID string,
	token string,
	expiresAt time.Time,
) (*string, error) {
	var refreshToken string
	err := db.QueryRow(
		ctx,
		`INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3) RETURNING token`,
		userID,
		token,
		expiresAt,
	).Scan(&refreshToken)

	if err != nil {
		log.Printf("[AuthRepository.CreateRefreshToken] Failed to create refresh token: %v", err)
		return nil, fmt.Errorf("Failed to create refresh token")
	}

	return &refreshToken, err
}

func (r *AuthRepository) GetRefreshToken(
	ctx context.Context,
	db dbtx,
	token string,
) (*string, *string, time.Time, error) {
	var userID string
	var userEmail string
	var expiresAt time.Time

	err := db.QueryRow(
		ctx,
		`SELECT u.id, u.email, rt.expires_at
		FROM refresh_tokens rt
		JOIN users u ON rt.user_id = u.id
		WHERE rt.token = $1 AND rt.expires_at > NOW() AND rt.revoked_at IS NULL
		FOR UPDATE OF rt`,
		token,
	).Scan(&userID, &userEmail, &expiresAt)
	if err != nil {
		// if the token is not found or expired, return a specific error
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, time.Time{}, fmt.Errorf("Refresh token not found or expired")
		}
		log.Printf("[AuthRepository.GetRefreshToken] Failed to get refresh token: %v", err)
	}

	return &userID, &userEmail, expiresAt, err
}

func (r *AuthRepository) RevokeRefreshToken(
	ctx context.Context,
	db dbtx,
	token string,
) error {
	_, err := db.Exec(
		ctx,
		`UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token = $1 AND expires_at > NOW() AND revoked_at IS NULL`,
		token,
	)
	if err != nil {
		log.Printf("[AuthRepository.RevokeRefreshToken] Failed to revoke refresh token: %v", err)
	}
	return err
}
