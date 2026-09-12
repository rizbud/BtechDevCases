package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

type PaginationRequest struct {
	Page     int `json:"page" default:"1" example:"1"`
	PageSize int `json:"page_size" default:"10" example:"10"`
}

func ValidateBodyRequest(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("Request body is empty")
		}
		return err
	}

	return nil
}

func ValidatePaginationRequest(r *http.Request) (int, int, map[string]string) {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	pageSizeStr := r.URL.Query().Get("page_size")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}

	validationErrors := map[string]string{}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		validationErrors["page"] = "Invalid page number"
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		validationErrors["page_size"] = "Invalid page size"
	}

	if page <= 0 {
		validationErrors["page"] = "Page must be greater than zero"
	}

	if pageSize <= 0 {
		validationErrors["page_size"] = "Page size must be greater than zero"
	}

	return page, pageSize, validationErrors
}
