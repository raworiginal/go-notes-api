package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims is the JWT payload structure.
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT token for the given userID and email.
// TODO(human): implement token signing using the secret key and jwt.NewWithClaims
func GenerateToken(userID int, email string, secret string) (string, error) {
	// TODO: implement
	// Hints:
	// - Create a Claims struct with userID, email, and expiry (e.g., 24 hours from now)
	// - Use jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// - Sign with the secret key: token.SignedString([]byte(secret))
	return "", nil
}

// ValidateToken parses and validates a JWT token, returning the claims if valid.
func ValidateToken(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// Verify the signing method is HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}
