package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/raworiginal/go-notes-api/internal/user"
)

type UsersHandler struct {
	service *user.Service
}

func NewUsersHandler(service *user.Service) *UsersHandler {
	return &UsersHandler{service}
}

// Register a new user POST /users/register
func (h *UsersHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	u, err := h.service.Register(req.Username, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidInput):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, user.ErrEmailTaken):
			http.Error(w, "Email already taken", http.StatusConflict)
		default:
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(u); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
}
