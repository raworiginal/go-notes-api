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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON"})
		return
	}

	u, err := h.service.Register(req.Username, req.Email, req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, user.ErrInvalidInput):
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		case errors.Is(err, user.ErrEmailTaken):
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"message": "Email already taken"})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(u); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		return
	}
}
