package transaction

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

func TestTransactionRepository_TopUpAndGetByID(t *testing.T) {
	repo := &TransactionRepository{DB: testPool}
	ctx := context.Background()
	userID := createTestUser(t, uniqueEmail("topup"))

	txID, err := repo.CreateTransaction(ctx, testPool, nil, userID, 50)
	if err != nil {
		t.Fatalf("CreateTransaction() error = %v", err)
	}

	tx, err := repo.GetTransactionByID(ctx, testPool, userID, txID)
	if err != nil {
		t.Fatalf("GetTransactionByID() error = %v", err)
	}
	if tx.RecipientID != userID {
		t.Errorf("RecipientID = %q, want %q", tx.RecipientID, userID)
	}
	if tx.SenderID != nil {
		t.Errorf("SenderID = %v, want nil for a top-up", tx.SenderID)
	}
	if tx.Amount != 50 {
		t.Errorf("Amount = %v, want 50", tx.Amount)
	}
}

func TestTransactionRepository_CreateTransaction_UnknownUser(t *testing.T) {
	repo := &TransactionRepository{DB: testPool}

	_, err := repo.CreateTransaction(context.Background(), testPool, nil, uuid.NewString(), 10)
	if err == nil {
		t.Fatal("expected error for a non-existent recipient, got nil")
	}
	if err.Error() != "User ID does not exist" {
		t.Errorf("error = %q, want %q", err.Error(), "User ID does not exist")
	}
}

func TestTransactionRepository_GetTransactionByID_MalformedUUID(t *testing.T) {
	repo := &TransactionRepository{DB: testPool}
	userID := createTestUser(t, uniqueEmail("malformed"))

	_, err := repo.GetTransactionByID(context.Background(), testPool, userID, "not-a-valid-uuid")
	if err == nil {
		t.Fatal("expected error for a malformed transaction ID, got nil")
	}
	want := "Transaction ID not-a-valid-uuid not found"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestTransactionRepository_GetTransactionByID_NotFound(t *testing.T) {
	repo := &TransactionRepository{DB: testPool}
	userID := createTestUser(t, uniqueEmail("notfound"))

	_, err := repo.GetTransactionByID(context.Background(), testPool, userID, uuid.NewString())
	if err == nil {
		t.Fatal("expected error for an unknown transaction ID, got nil")
	}
}

func TestTransactionRepository_GetUserBalance(t *testing.T) {
	repo := &TransactionRepository{DB: testPool}
	ctx := context.Background()
	userID := createTestUser(t, uniqueEmail("balance"))

	if _, err := repo.CreateTransaction(ctx, testPool, nil, userID, 100); err != nil {
		t.Fatalf("CreateTransaction() error = %v", err)
	}
	other := createTestUser(t, uniqueEmail("balance-other"))
	if _, err := repo.CreateTransaction(ctx, testPool, &userID, other, 40); err != nil {
		t.Fatalf("CreateTransaction() error = %v", err)
	}

	balance, err := repo.GetUserBalance(ctx, testPool, userID)
	if err != nil {
		t.Fatalf("GetUserBalance() error = %v", err)
	}
	if balance != 60 {
		t.Errorf("balance = %v, want 60", balance)
	}
}

func TestTransactionRepository_GetTransactionsByUserID_Pagination(t *testing.T) {
	repo := &TransactionRepository{DB: testPool}
	ctx := context.Background()
	userID := createTestUser(t, uniqueEmail("pagination"))

	for i := 0; i < 5; i++ {
		if _, err := repo.CreateTransaction(ctx, testPool, nil, userID, float64(i+1)); err != nil {
			t.Fatalf("CreateTransaction() error = %v", err)
		}
	}

	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	page1, total, err := repo.GetTransactionsByUserID(ctx, userID, today, tomorrow, 1, 2)
	if err != nil {
		t.Fatalf("GetTransactionsByUserID() page 1 error = %v", err)
	}
	if total != 5 {
		t.Fatalf("total = %d, want 5", total)
	}
	if len(page1) != 2 {
		t.Fatalf("page 1 len = %d, want 2", len(page1))
	}

	page2, _, err := repo.GetTransactionsByUserID(ctx, userID, today, tomorrow, 2, 2)
	if err != nil {
		t.Fatalf("GetTransactionsByUserID() page 2 error = %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("page 2 len = %d, want 2", len(page2))
	}
	if page1[0].ID == page2[0].ID {
		t.Error("page 1 and page 2 should not overlap")
	}

	page3, _, err := repo.GetTransactionsByUserID(ctx, userID, today, tomorrow, 3, 2)
	if err != nil {
		t.Fatalf("GetTransactionsByUserID() page 3 error = %v", err)
	}
	if len(page3) != 1 {
		t.Fatalf("page 3 len = %d, want 1 (5 records, page size 2)", len(page3))
	}
}
