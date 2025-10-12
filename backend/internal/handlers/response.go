// path: backend/internal/handlers/response.go
package handlers

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// respondSuccess sends a success response with data
func respondSuccess(w http.ResponseWriter, data interface{}) {
	respondJSON(w, http.StatusOK, SuccessResponse{
		Data: data,
	})
}

// respondCreated sends a created response
func respondCreated(w http.ResponseWriter, data interface{}) {
	respondJSON(w, http.StatusCreated, SuccessResponse{
		Data: data,
	})
}

// respondNoContent sends a no content response
func respondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
