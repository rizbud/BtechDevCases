package transaction

import (
	"context"
	"sync"
	"testing"
)

func newTestTransactionService() *TransactionService {
	return &TransactionService{Repository: &TransactionRepository{DB: testPool}}
}

func TestTransactionService_TopUp(t *testing.T) {
	s := newTestTransactionService()
	ctx := context.Background()
	userID := createTestUser(t, uniqueEmail("svc-topup"))

	tx, err := s.topUp(ctx, userID, 75)
	if err != nil {
		t.Fatalf("topUp() error = %v", err)
	}
	if tx.Amount != 75 {
		t.Errorf("Amount = %v, want 75", tx.Amount)
	}

	balance, err := s.getUserBalance(ctx, userID)
	if err != nil {
		t.Fatalf("getUserBalance() error = %v", err)
	}
	if balance != 75 {
		t.Errorf("balance = %v, want 75", balance)
	}
}

func TestTransactionService_Transfer(t *testing.T) {
	s := newTestTransactionService()
	ctx := context.Background()
	fromID := createTestUser(t, uniqueEmail("svc-transfer-from"))
	toID := createTestUser(t, uniqueEmail("svc-transfer-to"))

	if _, err := s.topUp(ctx, fromID, 100); err != nil {
		t.Fatalf("topUp() error = %v", err)
	}

	if _, err := s.transfer(ctx, fromID, toID, 40); err != nil {
		t.Fatalf("transfer() error = %v", err)
	}

	fromBalance, _ := s.getUserBalance(ctx, fromID)
	toBalance, _ := s.getUserBalance(ctx, toID)
	if fromBalance != 60 {
		t.Errorf("sender balance = %v, want 60", fromBalance)
	}
	if toBalance != 40 {
		t.Errorf("recipient balance = %v, want 40", toBalance)
	}
}

func TestTransactionService_Transfer_InsufficientBalance(t *testing.T) {
	s := newTestTransactionService()
	ctx := context.Background()
	fromID := createTestUser(t, uniqueEmail("svc-insufficient-from"))
	toID := createTestUser(t, uniqueEmail("svc-insufficient-to"))

	if _, err := s.topUp(ctx, fromID, 10); err != nil {
		t.Fatalf("topUp() error = %v", err)
	}

	_, err := s.transfer(ctx, fromID, toID, 1000)
	if err == nil {
		t.Fatal("expected error for insufficient balance, got nil")
	}
	if err.Error() != "Insufficient balance" {
		t.Errorf("error = %q, want %q", err.Error(), "Insufficient balance")
	}
}

func TestTransactionService_Transfer_ConcurrentOverdraft(t *testing.T) {
	s := newTestTransactionService()
	ctx := context.Background()
	fromID := createTestUser(t, uniqueEmail("svc-race-from"))
	toA := createTestUser(t, uniqueEmail("svc-race-to-a"))
	toB := createTestUser(t, uniqueEmail("svc-race-to-b"))

	if _, err := s.topUp(ctx, fromID, 100); err != nil {
		t.Fatalf("topUp() error = %v", err)
	}

	var wg sync.WaitGroup
	results := make([]error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, results[0] = s.transfer(ctx, fromID, toA, 60)
	}()
	go func() {
		defer wg.Done()
		_, results[1] = s.transfer(ctx, fromID, toB, 60)
	}()
	wg.Wait()

	balance, err := s.getUserBalance(ctx, fromID)
	if err != nil {
		t.Fatalf("getUserBalance() error = %v", err)
	}

	successes := 0
	for _, err := range results {
		if err == nil {
			successes++
		}
	}

	if balance < 0 {
		t.Errorf("sender balance went negative (%v) with %d successful concurrent transfers out of a 100 balance: transfer() does not serialize the balance check against the insert", balance, successes)
	}
}
