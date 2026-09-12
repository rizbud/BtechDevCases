package transaction

import (
	"testing"
	"time"
)

func TestValidateTransferRequest(t *testing.T) {
	s := &TransactionService{}

	cases := []struct {
		name       string
		fromUserID string
		toUserID   string
		amount     float64
		wantFields []string
	}{
		{"valid", "user-1", "user-2", 10, nil},
		{"missing from", "", "user-2", 10, []string{"from_user_id"}},
		{"missing to", "user-1", "", 10, []string{"to_user_email"}},
		{"same user", "user-1", "user-1", 10, []string{"to_user_email"}},
		{"zero amount", "user-1", "user-2", 0, []string{"amount"}},
		{"negative amount", "user-1", "user-2", -5, []string{"amount"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := s.validateTransferRequest(c.fromUserID, c.toUserID, c.amount)
			if len(c.wantFields) != len(errs) {
				t.Fatalf("validateTransferRequest() errors = %v, want fields %v", errs, c.wantFields)
			}
			for _, f := range c.wantFields {
				if _, ok := errs[f]; !ok {
					t.Errorf("expected validation error for field %q, got %v", f, errs)
				}
			}
		})
	}
}

func TestValidateTransactionsRequest(t *testing.T) {
	s := &TransactionService{}
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	cases := []struct {
		name      string
		start     string
		end       string
		wantField string
	}{
		{"valid range", yesterday, today, ""},
		{"missing start", "", today, "start_date"},
		{"missing end", yesterday, "", "end_date"},
		{"invalid start format", "01-01-2024", today, "start_date"},
		{"invalid end format", yesterday, "not-a-date", "end_date"},
		{"start after end", today, yesterday, "end_date"},
		{"end in future", yesterday, tomorrow, "end_date"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := s.validateTransactionsRequest(c.start, c.end)
			if c.wantField == "" {
				if len(errs) != 0 {
					t.Errorf("expected no validation errors, got %v", errs)
				}
				return
			}
			if _, ok := errs[c.wantField]; !ok {
				t.Errorf("expected validation error for field %q, got %v", c.wantField, errs)
			}
		})
	}
}
