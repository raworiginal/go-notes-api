package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/raworiginal/go-notes-api/internal/auth"
	"github.com/raworiginal/go-notes-api/internal/note"
)

type NotesHandler struct {
	service *note.Service
}

func NewNotesHandler(service *note.Service) *NotesHandler {
	return &NotesHandler{service}
}

// GetAll -List all notes  GET /notes
func (h *NotesHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	notes, err := h.service.GetAll(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notes); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		return
	}
}

// GetByID GET /notes/{id}
func (h *NotesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	// Extract {id} from path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid ID"})
		return
	}

	n, err := h.service.GetByID(userID, id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, note.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Note not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(n); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		return
	}
}

// Create a note POST /notes
func (h *NotesHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	// Parse JSON body
	var req struct {
		Title string        `json:"title"`
		Body  *string       `json:"body"`
		Type  note.NoteType `json:"type"`
		Todos []note.Todo   `json:"todos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON"})
		return
	}

	// Default type to text if not provided
	if req.Type == "" {
		req.Type = note.NoteTypeText
	}

	n, err := h.service.CreateWithType(userID, req.Title, req.Body, req.Type, req.Todos...)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, note.ErrInvalidInput) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid input"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(n); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		return
	}
}

// Update a note - PUT /notes/{id}
func (h *NotesHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	// Extract {id} from path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid ID"})
		return
	}

	// Parse JSON body
	var req struct {
		Title string        `json:"title"`
		Body  *string       `json:"body"`
		Type  note.NoteType `json:"type"`
		Todos []note.Todo   `json:"todos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON"})
		return
	}
	n, err := h.service.UpdateWithTypeAndTodos(userID, id, req.Title, req.Body, req.Type, req.Todos)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, note.ErrInvalidInput) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid input"})
			return
		}
		if errors.Is(err, note.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Note not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(n); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		return
	}
}

// Delete a note - DELETE /notes/{id}
func (h *NotesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	// Extract {id} from path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid ID"})
		return
	}
	if err := h.service.Delete(userID, id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, note.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Note not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "note deleted"}); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		return
	}
}
