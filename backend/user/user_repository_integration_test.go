package user

import (
	"context"
	"fmt"
	"os"
	"testing"

	"btech-wallet/testutil"

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

func TestUserRepository_CreateAndGetByEmail(t *testing.T) {
	repo := &UserRepository{DB: testPool}
	ctx := context.Background()
	email := uniqueEmail("create")
	t.Cleanup(func() { testutil.CleanupUserByEmail(testPool, email) })

	if err := repo.CreateUser(ctx, email, "hashed-password"); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	got, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByEmail() error = %v", err)
	}
	if got.Email != email {
		t.Errorf("Email = %q, want %q", got.Email, email)
	}
	if got.Password != "hashed-password" {
		t.Errorf("Password = %q, want %q", got.Password, "hashed-password")
	}
	if got.ID == "" {
		t.Error("expected a generated ID")
	}
}

func TestUserRepository_CreateDuplicateEmail(t *testing.T) {
	repo := &UserRepository{DB: testPool}
	ctx := context.Background()
	email := uniqueEmail("dup")
	t.Cleanup(func() { testutil.CleanupUserByEmail(testPool, email) })

	if err := repo.CreateUser(ctx, email, "hashed-password"); err != nil {
		t.Fatalf("first CreateUser() error = %v", err)
	}

	err := repo.CreateUser(ctx, email, "hashed-password")
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
	want := fmt.Sprintf("User with email %s already exists", email)
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestUserRepository_GetUserByEmail_NotFound(t *testing.T) {
	repo := &UserRepository{DB: testPool}
	email := uniqueEmail("missing")

	_, err := repo.GetUserByEmail(context.Background(), email)
	if err == nil {
		t.Fatal("expected error for unknown email, got nil")
	}
}

func TestUserRepository_GetUserByID(t *testing.T) {
	repo := &UserRepository{DB: testPool}
	ctx := context.Background()
	email := uniqueEmail("byid")
	t.Cleanup(func() { testutil.CleanupUserByEmail(testPool, email) })

	if err := repo.CreateUser(ctx, email, "hashed-password"); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	created, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByEmail() error = %v", err)
	}

	got, err := repo.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}
	if got.Email != email {
		t.Errorf("Email = %q, want %q", got.Email, email)
	}
}

func TestUserRepository_GetUserByID_NotFound(t *testing.T) {
	repo := &UserRepository{DB: testPool}

	_, err := repo.GetUserByID(context.Background(), uuid.NewString())
	if err == nil {
		t.Fatal("expected error for unknown user ID, got nil")
	}
}
