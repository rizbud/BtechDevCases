package transaction

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"btech-wallet/server"
	"btech-wallet/user"
)

// These tests cover request parsing and validation paths that return before
// any repository/DB call is made (all handlers read "user_id" from the
// request context, which AuthMiddleware would normally set).

func newTestTransactionHandler() *TransactionHandler {
	return &TransactionHandler{
		TrxService:  &TransactionService{},
		UserService: &user.UserService{},
	}
}

func newRequestWithUser(method, target, body string) *http.Request {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	ctx := context.WithValue(r.Context(), "user_id", "user-1")
	ctx = context.WithValue(ctx, "email", "user-1@example.com")
	return r.WithContext(ctx)
}

func decodeErrors(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var resp server.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	errs, ok := resp.Error.(map[string]any)
	if !ok {
		t.Fatalf("expected error field to be a map, got %v (%T)", resp.Error, resp.Error)
	}
	return errs
}

func TestHandleTransfer_InvalidBody(t *testing.T) {
	h := newTestTransactionHandler()
	r := newRequestWithUser(http.MethodPost, "/wallet/transfer", "not-json")
	w := httptest.NewRecorder()

	h.handleTransfer(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleTransfer_InvalidRecipientEmail(t *testing.T) {
	h := newTestTransactionHandler()
	body := `{"recipient":"not-an-email","amount":10}`
	r := newRequestWithUser(http.MethodPost, "/wallet/transfer", body)
	w := httptest.NewRecorder()

	h.handleTransfer(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	errs := decodeErrors(t, w)
	if _, ok := errs["recipient"]; !ok {
		t.Error("expected recipient validation error")
	}
}

func TestHandleTopUp_InvalidBody(t *testing.T) {
	h := newTestTransactionHandler()
	r := newRequestWithUser(http.MethodPost, "/wallet/topup", "")
	w := httptest.NewRecorder()

	h.handleTopUp(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetTransactions_MissingDates(t *testing.T) {
	h := newTestTransactionHandler()
	r := newRequestWithUser(http.MethodGet, "/wallet/transactions", "")
	w := httptest.NewRecorder()

	h.handleGetTransactions(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	errs := decodeErrors(t, w)
	if _, ok := errs["start_date"]; !ok {
		t.Error("expected start_date validation error")
	}
	if _, ok := errs["end_date"]; !ok {
		t.Error("expected end_date validation error")
	}
}

func TestHandleGetTransactions_InvalidPagination(t *testing.T) {
	h := newTestTransactionHandler()
	today := time.Now().Format("2006-01-02")
	target := "/wallet/transactions?start_date=" + today + "&end_date=" + today + "&page=0"
	r := newRequestWithUser(http.MethodGet, target, "")
	w := httptest.NewRecorder()

	h.handleGetTransactions(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	errs := decodeErrors(t, w)
	if _, ok := errs["page"]; !ok {
		t.Error("expected page validation error")
	}
}
