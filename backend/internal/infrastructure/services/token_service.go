package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/techappsUT/social-queue/internal/application/common"
)

// JWTTokenService implements common.TokenService using JWT
type JWTTokenService struct {
	accessSecret  string
	refreshSecret string
}

// NewJWTTokenService creates a new JWT token service
func NewJWTTokenService(accessSecret, refreshSecret string) common.TokenService {
	return &JWTTokenService{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
	}
}

// Keep your existing GenerateAccessToken method AS IS
func (s *JWTTokenService) GenerateAccessToken(userID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.accessSecret))
}

// Keep your existing GenerateRefreshToken method AS IS
func (s *JWTTokenService) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.refreshSecret))
}

// Keep your existing ValidateAccessToken method AS IS
func (s *JWTTokenService) ValidateAccessToken(tokenString string) (*common.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.accessSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check token type
		if tokenType, ok := claims["type"].(string); !ok || tokenType != "access" {
			return nil, fmt.Errorf("invalid token type")
		}

		// Return claims without EmailVerified (not in TokenClaims struct)
		return &common.TokenClaims{
			UserID: claims["user_id"].(string),
			Email:  claims["email"].(string),
			Role:   claims["role"].(string),
			// TeamID is optional - check if exists
			TeamID: func() string {
				if teamID, ok := claims["team_id"].(string); ok {
					return teamID
				}
				return ""
			}(),
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateRefreshToken validates a refresh token and returns claims
func (s *JWTTokenService) ValidateRefreshToken(tokenString string) (*common.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.refreshSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check token type
		if tokenType, ok := claims["type"].(string); !ok || tokenType != "refresh" {
			return nil, fmt.Errorf("invalid token type")
		}

		return &common.TokenClaims{
			UserID: claims["user_id"].(string),
			Email:  "", // Refresh tokens don't include email
			Role:   "", // Refresh tokens don't include role
			TeamID: "", // Refresh tokens don't include teamID
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RevokeRefreshToken revokes a refresh token (placeholder implementation)
func (s *JWTTokenService) RevokeRefreshToken(ctx context.Context, token string) error {
	// TODO: Implement token blacklist in Redis/database when available
	log.Printf("🔐 Revoking refresh token: %s...", token[:min(20, len(token))])
	return nil
}

// Helper function for min (if not available in your Go version)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
