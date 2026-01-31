package handlers

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

func SendError(w http.ResponseWriter, r *http.Request, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
		Code:    status,
	}

	if err != nil && status >= 500 {
		resp.Message = "An internal error occurred"
	}

	json.NewEncoder(w).Encode(resp)
}

func SendJSON(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

var (
	ErrTableNotFound    = "Table not found"
	ErrRecordNotFound   = "Record not found"
	ErrInvalidRequest   = "Invalid request"
	ErrInvalidFilter    = "Invalid filter"
	ErrUnauthorized     = "Unauthorized"
	ErrForbidden        = "Forbidden"
	ErrInvalidInput     = "Invalid input"
	ErrDatabaseError    = "Database error"
	ErrConflict         = "Conflict"
)
