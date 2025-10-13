// ============================================================================
// FILE: backend/internal/infrastructure/services/encryption.go
// FIXED VERSION - Type conversion bugs resolved
// ============================================================================
package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptionService handles encryption and decryption of sensitive data using AES-256-GCM
type EncryptionService struct {
	key []byte
}

// NewEncryptionService creates a new encryption service
// key must be 32 bytes for AES-256
func NewEncryptionService(key string) (*EncryptionService, error) {
	keyBytes := []byte(key)

	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("encryption key must be exactly 32 bytes for AES-256, got %d bytes", len(keyBytes))
	}

	return &EncryptionService{
		key: keyBytes,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns base64-encoded ciphertext
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("plaintext cannot be empty")
	}

	// Create AES cipher block
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-GCM
func (s *EncryptionService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", fmt.Errorf("ciphertext cannot be empty")
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create AES cipher block
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// GenerateKey generates a random 32-byte key for AES-256 (for initial setup)
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
