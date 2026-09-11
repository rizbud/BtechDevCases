package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func ValidateBodyRequest(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("Request body is empty")
		}
		return err
	}

	return nil
}
