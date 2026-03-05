// Package handler manages the http handlers
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/raworiginal/go-notes-api/internal/auth"
	"github.com/raworiginal/go-notes-api/internal/user"
)

type AuthHandler struct {
	userService *user.Service
	jwtSecret   string
}

func NewAuthHandler(userService *user.Service, jwtSecret string) *AuthHandler {
	return &AuthHandler{userService, jwtSecret}
}

// Login Method POST /login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON"})
		return
	}

	u, err := h.userService.Authenticate(req.Email, req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, user.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid credentials"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		return
	}

	token, err := auth.GenerateToken(u.ID, u.Email, u.Username, h.jwtSecret)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Failed to generate token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
	}
}
