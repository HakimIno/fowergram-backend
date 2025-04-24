package security

import (
	"fmt"
	"time"

	"fowergram/internal/core/domain"

	"encoding/hex"

	"crypto/rand"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

func GenerateJWT(userID uint, secret string, expiration time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiration).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID uint, secret string, expiration time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiration).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   "refresh",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString string, secret string) (uint, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return 0, err
	}

	return claims.UserID, nil
}

func ValidateRefreshToken(tokenString string, secret string) (uint, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid || claims.Subject != "refresh" {
		return 0, err
	}

	return claims.UserID, nil
}

func GenerateDeviceID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate device ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func ValidateToken(tokenString string, secret string) (*domain.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &domain.User{ID: claims.UserID}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateRefreshTokenAsAccessToken validates a refresh token but treats it as an access token
// This is useful for endpoints that should accept either token type
func ValidateRefreshTokenAsAccessToken(tokenString string, secret string) (*domain.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Don't check the subject field here, so it works with both access and refresh tokens
		return &domain.User{ID: claims.UserID}, nil
	}

	return nil, fmt.Errorf("invalid token")
}
