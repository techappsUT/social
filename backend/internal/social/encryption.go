// path: backend/internal/social/encryption.go
package social

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// TokenEncryption handles encryption/decryption of sensitive token data
type TokenEncryption struct {
	key []byte // 32 bytes for AES-256
}

// NewTokenEncryption creates a new encryption service
// In production, retrieve key from KMS (AWS KMS, Google Cloud KMS, HashiCorp Vault)
func NewTokenEncryption(encryptionKey string) (*TokenEncryption, error) {
	key := []byte(encryptionKey)
	if len(key) != 32 {
		return nil, errors.New("encryption key must be 32 bytes for AES-256")
	}

	return &TokenEncryption{
		key: key,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (te *TokenEncryption) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(te.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (te *TokenEncryption) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(te.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptToken encrypts sensitive token fields
func (te *TokenEncryption) EncryptToken(token *PlatformToken) error {
	if token.AccessToken != "" {
		encrypted, err := te.Encrypt(token.AccessToken)
		if err != nil {
			return err
		}
		token.AccessToken = encrypted
	}

	if token.RefreshToken != "" {
		encrypted, err := te.Encrypt(token.RefreshToken)
		if err != nil {
			return err
		}
		token.RefreshToken = encrypted
	}

	return nil
}

// DecryptToken decrypts sensitive token fields
func (te *TokenEncryption) DecryptToken(token *PlatformToken) error {
	if token.AccessToken != "" {
		decrypted, err := te.Decrypt(token.AccessToken)
		if err != nil {
			return err
		}
		token.AccessToken = decrypted
	}

	if token.RefreshToken != "" {
		decrypted, err := te.Decrypt(token.RefreshToken)
		if err != nil {
			return err
		}
		token.RefreshToken = decrypted
	}

	return nil
}
