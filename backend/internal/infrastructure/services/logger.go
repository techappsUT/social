// path: backend/internal/infrastructure/services/logger.go
package services

import (
	"log"
	"os"

	"github.com/techappsUT/social-queue/internal/application/common"
)

// Logger implements common.Logger using standard log package
type Logger struct {
	logger *log.Logger
}

// NewLogger creates a new logger
func NewLogger() common.Logger {
	return &Logger{
		logger: log.New(os.Stdout, "[APP] ", log.LstdFlags|log.Lshortfile),
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.logger.Printf("DEBUG: %s %v", msg, fields)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.logger.Printf("INFO: %s %v", msg, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.logger.Printf("WARN: %s %v", msg, fields)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.logger.Printf("ERROR: %s %v", msg, fields)
}
