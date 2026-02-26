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

// GET /notes - List all notes
func (h *NotesHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	notes, err := h.service.GetAll(userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notes); err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
}

// GET /notes/{id} - Get note by ID
func (h *NotesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	// Extract {id} from path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	n, err := h.service.GetByID(userID, id)
	if err != nil {
		if errors.Is(err, note.ErrNotFound) {
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(n); err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
}

// POST /notes - Create a note
func (h *NotesHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	// Parse JSON body
	var req struct {
		Title string      `json:"title"`
		Body  string      `json:"body"`
		Type  note.NoteType `json:"type"`
		Todos []note.Todo `json:"todos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Default type to text if not provided
	if req.Type == "" {
		req.Type = note.NoteTypeText
	}

	n, err := h.service.CreateWithType(userID, req.Title, req.Body, req.Type, req.Todos...)
	if err != nil {
		if errors.Is(err, note.ErrInvalidInput) {
			http.Error(w, "Invalid Input", http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(n); err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
}

// PUT /notes/{id} - Update a note
func (h *NotesHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	// Extract {id} from path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Parse JSON body
	var req struct {
		Title string      `json:"title"`
		Body  *string     `json:"body"`
		Type  note.NoteType `json:"type"`
		Todos []note.Todo `json:"todos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	n, err := h.service.UpdateWithTypeAndTodos(userID, id, req.Title, req.Body, req.Type, req.Todos)
	if err != nil {
		if errors.Is(err, note.ErrInvalidInput) {
			http.Error(w, "Invalid Input", http.StatusBadRequest)
			return
		}
		if errors.Is(err, note.ErrNotFound) {
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(n); err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
}

// DELETE /notes/{id} - Delete a note
func (h *NotesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	// Extract {id} from path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if err := h.service.Delete(userID, id); err != nil {
		if errors.Is(err, note.ErrNotFound) {
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "note deleted"}); err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
}
