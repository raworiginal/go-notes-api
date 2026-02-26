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
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	u, err := h.userService.Authenticate(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(u.ID, u.Email, u.Username, h.jwtSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}
}
