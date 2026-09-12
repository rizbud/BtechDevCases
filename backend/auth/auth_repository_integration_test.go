package auth

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"btech-wallet/testutil"
	"btech-wallet/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	pool, err := testutil.Connect()
	if err != nil {
		fmt.Println("skipping integration tests: database unavailable:", err)
		os.Exit(0)
	}
	testPool = pool
	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%s@example.com", prefix, uuid.NewString())
}

func createTestUser(t *testing.T, email string) string {
	t.Helper()
	ctx := context.Background()
	userRepo := &user.UserRepository{DB: testPool}
	if err := userRepo.CreateUser(ctx, email, "hashed-password"); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	t.Cleanup(func() { testutil.CleanupUserByEmail(testPool, email) })

	u, err := userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("failed to fetch created test user: %v", err)
	}
	return u.ID
}

func TestAuthRepository_CreateAndGetRefreshToken(t *testing.T) {
	repo := &AuthRepository{DB: testPool}
	ctx := context.Background()
	email := uniqueEmail("refresh")
	userID := createTestUser(t, email)

	token, err := repo.CreateRefreshToken(ctx, testPool, userID, uuid.NewString(), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateRefreshToken() error = %v", err)
	}

	gotUserID, gotEmail, _, err := repo.GetRefreshToken(ctx, testPool, *token)
	if err != nil {
		t.Fatalf("GetRefreshToken() error = %v", err)
	}
	if *gotUserID != userID {
		t.Errorf("userID = %q, want %q", *gotUserID, userID)
	}
	if *gotEmail != email {
		t.Errorf("email = %q, want %q", *gotEmail, email)
	}
}

func TestAuthRepository_GetRefreshToken_UnknownToken(t *testing.T) {
	repo := &AuthRepository{DB: testPool}

	_, _, _, err := repo.GetRefreshToken(context.Background(), testPool, uuid.NewString())
	if err == nil {
		t.Fatal("expected error for unknown refresh token, got nil")
	}
	if err.Error() != "Refresh token not found or expired" {
		t.Errorf("error = %q, want %q", err.Error(), "Refresh token not found or expired")
	}
}

func TestAuthRepository_GetRefreshToken_Expired(t *testing.T) {
	repo := &AuthRepository{DB: testPool}
	ctx := context.Background()
	userID := createTestUser(t, uniqueEmail("expired"))

	token, err := repo.CreateRefreshToken(ctx, testPool, userID, uuid.NewString(), time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("CreateRefreshToken() error = %v", err)
	}

	_, _, _, err = repo.GetRefreshToken(ctx, testPool, *token)
	if err == nil {
		t.Fatal("expected error for expired refresh token, got nil")
	}
}

func TestAuthRepository_RevokeRefreshToken(t *testing.T) {
	repo := &AuthRepository{DB: testPool}
	ctx := context.Background()
	userID := createTestUser(t, uniqueEmail("revoke"))

	token, err := repo.CreateRefreshToken(ctx, testPool, userID, uuid.NewString(), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateRefreshToken() error = %v", err)
	}

	if err := repo.RevokeRefreshToken(ctx, testPool, *token); err != nil {
		t.Fatalf("RevokeRefreshToken() error = %v", err)
	}

	if _, _, _, err := repo.GetRefreshToken(ctx, testPool, *token); err == nil {
		t.Fatal("expected error after token was revoked, got nil")
	}
}
