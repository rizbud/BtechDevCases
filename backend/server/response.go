package server

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message" default:"Hello World!" example:"Hello World!"`
	Error   any    `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func ErrorResponseJSON(w http.ResponseWriter, statusCode int, message string, err any) {
	JSON(w, statusCode, ErrorResponse{
		Message: message,
		Error:   err,
	})
}
