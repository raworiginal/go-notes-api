// Package middleware handles all middleware for api requests and responses
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "request-id"

// Middleware that injects request ID
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Retrieve the request ID from context
func RequestIDFromContext(ctx context.Context) string {
	// TODO: Extract from context
	id, ok := ctx.Value(requestIDKey).(string)
	if !ok || id == "" {
		return ""
	}
	return id
}
