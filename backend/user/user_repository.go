package user

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

type UserRepository struct {
	DB *pgxpool.Pool
}

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *UserRepository) CreateUser(ctx context.Context, email string, password string) error {
	_, err := r.DB.Exec(
		ctx,
		"INSERT INTO users (email, password) VALUES ($1, $2)",
		email,
		password,
	)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("User with email %s already exists", email)
		}

		log.Printf("[UsersRepository.CreateUser] Failed to create user: %v", err)
		return fmt.Errorf("Failed to create user")
	}

	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	err := r.DB.QueryRow(
		ctx,
		"SELECT id, email, password FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("User with email %s not found", email)
		}
		log.Printf("[UsersRepository.GetUserByEmail] Failed to get user by email: %v", err)
		return nil, fmt.Errorf("Failed to get user by email %s", email)
	}

	return &user, nil
}
