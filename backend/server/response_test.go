package server

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, 201, MessageResponse{Message: "ok"})

	if w.Code != 201 {
		t.Errorf("status = %d, want 201", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body MessageResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Message != "ok" {
		t.Errorf("Message = %q, want %q", body.Message, "ok")
	}
}

func TestErrorResponseJSON(t *testing.T) {
	t.Run("with explicit message", func(t *testing.T) {
		w := httptest.NewRecorder()
		ErrorResponseJSON(w, 400, "bad input", map[string]string{"field": "required"})

		var body ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &body)
		if body.Message != "bad input" {
			t.Errorf("Message = %q, want %q", body.Message, "bad input")
		}
	})

	t.Run("falls back to status text when message empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		ErrorResponseJSON(w, 404, "", nil)

		var body ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &body)
		if body.Message != "Not Found" {
			t.Errorf("Message = %q, want %q", body.Message, "Not Found")
		}
	})
}

func TestPaginationResponseJSON(t *testing.T) {
	w := httptest.NewRecorder()
	PaginationResponseJSON(w, 200, "hi", []string{"a", "b"}, 2, 1, 1, 10)

	var body PaginationResponse[string]
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if len(body.Data) != 2 || body.TotalRecords != 2 || body.TotalPages != 1 {
		t.Errorf("unexpected body: %+v", body)
	}
}

func TestPaginationResponseJSONNilData(t *testing.T) {
	w := httptest.NewRecorder()
	PaginationResponseJSON[string](w, 200, "", nil, 0, 0, 1, 10)

	var body PaginationResponse[string]
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Data == nil {
		t.Error("expected Data to be an empty slice, not nil, when input is nil")
	}
}
