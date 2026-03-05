package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// contextKey is an unexported type for context keys, preventing collisions.
type contextKey string

const userIDKey contextKey = "userID"

// Middleware validates the JWT token from the Authorization header
// and injects the userID into the request context.
func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization: Bearer <token>
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"message": "Missing authorization header"})
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"message": "Invalid authorization format"})
				return
			}

			tokenString := parts[1]
			claims, err := ValidateToken(tokenString, secret)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"message": "Invalid or expired token"})
				return
			}

			// Inject userID into context
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the userID from the request context.
// Returns 0 if not found (which is never a valid user ID).
func UserIDFromContext(ctx context.Context) int {
	userID, ok := ctx.Value(userIDKey).(int)
	if !ok {
		return 0
	}
	return userID
}
