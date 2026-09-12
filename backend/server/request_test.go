package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateBodyRequest(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	t.Run("valid body", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"alice"}`))
		var p payload
		if err := ValidateBodyRequest(r, &p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "alice" {
			t.Errorf("Name = %q, want %q", p.Name, "alice")
		}
	})

	t.Run("empty body", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		var p payload
		err := ValidateBodyRequest(r, &p)
		if err == nil {
			t.Fatal("expected error for empty body, got nil")
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{invalid`))
		var p payload
		if err := ValidateBodyRequest(r, &p); err == nil {
			t.Fatal("expected error for malformed json, got nil")
		}
	})
}

func TestValidatePaginationRequest(t *testing.T) {
	cases := []struct {
		name       string
		query      string
		wantPage   int
		wantSize   int
		wantFields []string
	}{
		{"defaults", "", 1, 10, nil},
		{"custom values", "page=2&page_size=20", 2, 20, nil},
		{"invalid page", "page=abc&page_size=10", 0, 10, []string{"page"}},
		{"zero page", "page=0&page_size=10", 0, 10, []string{"page"}},
		{"negative page size", "page=1&page_size=-5", 1, -5, []string{"page_size"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/?"+c.query, nil)
			page, pageSize, errs := ValidatePaginationRequest(r)
			if page != c.wantPage {
				t.Errorf("page = %d, want %d", page, c.wantPage)
			}
			if pageSize != c.wantSize {
				t.Errorf("pageSize = %d, want %d", pageSize, c.wantSize)
			}
			if len(errs) != len(c.wantFields) {
				t.Fatalf("errors = %v, want fields %v", errs, c.wantFields)
			}
			for _, f := range c.wantFields {
				if _, ok := errs[f]; !ok {
					t.Errorf("expected validation error for field %q, got %v", f, errs)
				}
			}
		})
	}
}
