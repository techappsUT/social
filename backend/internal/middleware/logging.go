// ============================================================================
// FILE: backend/internal/middleware/logging.go
// PURPOSE: Structured HTTP request/response logging middleware
// ============================================================================

package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/techappsUT/social-queue/internal/application/common"
)

// Logger wraps a response writer to capture status code and size
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.size += size
	return size, err
}

// RequestLogger creates a middleware that logs HTTP requests
func RequestLogger(logger common.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID from context (set by chi middleware.RequestID)
			requestID := middleware.GetReqID(r.Context())

			// Wrap response writer
			wrapped := newLoggingResponseWriter(w)

			// Process request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Extract user info if available
			userID, _ := GetUserID(r.Context())

			// Log request details
			logger.Info(formatLogMessage(
				requestID,
				r.Method,
				r.URL.Path,
				r.URL.RawQuery,
				wrapped.statusCode,
				duration,
				wrapped.size,
				userID.String(),
				extractIP(r),
				r.UserAgent(),
			))
		})
	}
}

// formatLogMessage formats a structured log message
func formatLogMessage(
	requestID string,
	method string,
	path string,
	query string,
	status int,
	duration time.Duration,
	size int,
	userID string,
	ip string,
	userAgent string,
) string {
	msg := ""

	// Request info
	msg += "method=" + method + " "
	msg += "path=" + path + " "

	if query != "" {
		msg += "query=" + query + " "
	}

	// Response info
	msg += "status=" + http.StatusText(status) + " "
	msg += "status_code=" + string(rune(status+'0')) + " "
	msg += "duration=" + duration.String() + " "
	msg += "size=" + string(rune(size)) + "bytes "

	// Request metadata
	if requestID != "" {
		msg += "request_id=" + requestID + " "
	}

	if userID != "" && userID != "00000000-0000-0000-0000-000000000000" {
		msg += "user_id=" + userID + " "
	}

	msg += "ip=" + ip + " "

	// Add status indicator
	if status >= 500 {
		msg = "❌ " + msg
	} else if status >= 400 {
		msg = "⚠️  " + msg
	} else {
		msg = "✅ " + msg
	}

	return msg
}

// StructuredLogger creates a structured JSON logger middleware
func StructuredLogger(logger common.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := middleware.GetReqID(r.Context())
			wrapped := newLoggingResponseWriter(w)

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			userID, _ := GetUserID(r.Context())

			// For JSON structured logging (future enhancement)
			logData := map[string]interface{}{
				"request_id":  requestID,
				"method":      r.Method,
				"path":        r.URL.Path,
				"query":       r.URL.RawQuery,
				"status":      wrapped.statusCode,
				"duration_ms": duration.Milliseconds(),
				"size":        wrapped.size,
				"ip":          extractIP(r),
				"user_agent":  r.UserAgent(),
				"timestamp":   start.UTC().Format(time.RFC3339),
			}

			if userID.String() != "00000000-0000-0000-0000-000000000000" {
				logData["user_id"] = userID.String()
			}

			// Log based on status code
			if wrapped.statusCode >= 500 {
				logger.Error(formatStructuredLog(logData))
			} else if wrapped.statusCode >= 400 {
				logger.Warn(formatStructuredLog(logData))
			} else {
				logger.Debug(formatStructuredLog(logData))
			}
		})
	}
}

// formatStructuredLog formats log data as a string
func formatStructuredLog(data map[string]interface{}) string {
	msg := ""
	for k, v := range data {
		msg += k + "=" + toString(v) + " "
	}
	return msg
}

// toString converts interface to string
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return string(rune(val + '0'))
	case int64:
		return string(rune(val + '0'))
	default:
		return ""
	}
}

// RecoveryLogger logs panic recoveries
func RecoveryLogger(logger common.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := middleware.GetReqID(r.Context())
					logger.Error("PANIC RECOVERED: request_id=" + requestID + " error=" + toString(err))

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"Internal server error","message":"An unexpected error occurred"}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
