// ===========================================================================
// FILE: backend/internal/middleware/validation.go
// Add this NEW middleware to validate requests before they reach handlers
// ===========================================================================
package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError response structure
// type ValidationErrorResponse struct {
// 	Error   string            `json:"error"`
// 	Message string            `json:"message"`
// 	Fields  map[string]string `json:"fields,omitempty"`
// }

// SuccessResponse for all successful API responses
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse for all error responses
type ErrorResponse struct {
	Success bool              `json:"success"`
	Error   string            `json:"error"`             // Error code: "validation_error", "not_found", etc.
	Message string            `json:"message"`           // Human-readable message
	Details map[string]string `json:"details,omitempty"` // Field-specific errors for validation
}

// ValidateRequest is a middleware that validates request body against struct tags
// Usage: Wrap handlers that decode JSON
func ValidateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only validate POST, PUT, PATCH requests with JSON body
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				// RespondValidationError(w, http.StatusBadRequest, "Content-Type must be application/json", nil)
				RespondValidationError(w, "Content-Type must be application/json", nil)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// ValidateStruct validates a struct using validator tags
// Call this from handlers after decoding JSON
func ValidateStruct(v interface{}) error {
	return validate.Struct(v)
}

// FormatValidationErrors converts validator errors to map
func FormatValidationErrors(err error) map[string]string {
	fields := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()

			switch e.Tag() {
			case "required":
				fields[field] = field + " is required"
			case "email":
				fields[field] = field + " must be a valid email"
			case "min":
				fields[field] = field + " must be at least " + e.Param() + " characters"
			case "max":
				fields[field] = field + " must be at most " + e.Param() + " characters"
			case "uuid":
				fields[field] = field + " must be a valid UUID"
			case "oneof":
				fields[field] = field + " must be one of: " + e.Param()
			default:
				fields[field] = field + " validation failed: " + e.Tag()
			}
		}
	}

	return fields
}

// respondValidationError sends validation error response
// func RespondValidationError(w http.ResponseWriter, status int, message string, fields map[string]string) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(status)

// 	response := ValidationErrorResponse{
// 		Error:   "validation_error",
// 		Message: message,
// 		Fields:  fields,
// 	}

// 	json.NewEncoder(w).Encode(response)
// }

// RespondSuccess sends a successful JSON response
func RespondSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// RespondCreated sends a 201 Created response
func RespondCreated(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// RespondNoContent sends a 204 No Content response
func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// ===========================================================================
// ERROR RESPONSE HELPERS
// ===========================================================================

// RespondError sends a generic error response
func RespondError(w http.ResponseWriter, status int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Error:   errorCode,
		Message: message,
	})
}

// RespondValidationError sends validation error with field details
func RespondValidationError(w http.ResponseWriter, message string, fieldErrors map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Error:   "validation_error",
		Message: message,
		Details: fieldErrors,
	})
}

// RespondNotFound sends 404 Not Found
func RespondNotFound(w http.ResponseWriter, resource string) {
	RespondError(w, http.StatusNotFound, "not_found", resource+" not found")
}

// RespondUnauthorized sends 401 Unauthorized
func RespondUnauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Authentication required"
	}
	RespondError(w, http.StatusUnauthorized, "unauthorized", message)
}

// RespondForbidden sends 403 Forbidden
func RespondForbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "You don't have permission to access this resource"
	}
	RespondError(w, http.StatusForbidden, "forbidden", message)
}

// RespondConflict sends 409 Conflict
func RespondConflict(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusConflict, "conflict", message)
}

// RespondInternalError sends 500 Internal Server Error
func RespondInternalError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "An internal server error occurred"
	}
	RespondError(w, http.StatusInternalServerError, "internal_error", message)
}
