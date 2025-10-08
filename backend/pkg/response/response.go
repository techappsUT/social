// path: backend/pkg/response/response.go
package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// Error writes an error JSON response
func Error(w http.ResponseWriter, status int, message string, err error) {
	errorMsg := message
	if err != nil {
		log.Printf("API Error: %s - %v", message, err)
		errorMsg = err.Error()
	}

	JSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: errorMsg,
		Code:    status,
	})
}

// Success writes a success JSON response
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}
