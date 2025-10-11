package services

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/techappsUT/social-queue/internal/application/common"
)

type JWTTokenService struct {
	accessSecret  string
	refreshSecret string
}

func NewJWTTokenService(accessSecret, refreshSecret string) common.TokenService {
	return &JWTTokenService{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
	}
}

func (s *JWTTokenService) GenerateAccessToken(userID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.accessSecret))
}

func (s *JWTTokenService) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.refreshSecret))
}

func (s *JWTTokenService) ValidateAccessToken(tokenString string) (*common.TokenClaims, error) {
	// TODO: Implement
	return nil, nil
}

func (s *JWTTokenService) RevokeRefreshToken(ctx context.Context, token string) error {
	// TODO: Implement
	return nil
}
