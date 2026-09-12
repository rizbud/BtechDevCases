package server

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Error   any    `json:"error,omitempty"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type PaginationResponse[T any] struct {
	Data         []T `json:"data"`
	TotalRecords int `json:"total_records"`
	TotalPages   int `json:"total_pages"`
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
}

func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func ErrorResponseJSON(w http.ResponseWriter, statusCode int, message string, err any) {
	msg := message
	if msg == "" {
		msg = http.StatusText(statusCode)
	}

	JSON(w, statusCode, ErrorResponse{
		Message: msg,
		Error:   err,
	})
}

func PaginationResponseJSON[T any](
	w http.ResponseWriter,
	statusCode int,
	data []T,
	totalRecords,
	totalPages,
	currentPage,
	pageSize int,
) {
	if data == nil {
		data = []T{}
	}
	JSON(w, statusCode, PaginationResponse[T]{
		Data:         data,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  currentPage,
		PageSize:     pageSize,
	})
}
