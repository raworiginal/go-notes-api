package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raworiginal/go-notes-api/internal/note"
)

// mockRepo is a test double for note.Repository.
// Each method delegates to a function field so individual test cases
// can inject exactly the behavior they need without a full mock struct per case.
type mockRepo struct {
	getAllFn   func(userID int) ([]*note.Note, error)
	getByIDFn func(userID, id int) (*note.Note, error)
	createFn  func(n *note.Note) error
	updateFn  func(n *note.Note) error
	deleteFn  func(userID, id int) error
}

func (m *mockRepo) GetAll(userID int) ([]*note.Note, error) {
	if m.getAllFn != nil {
		return m.getAllFn(userID)
	}
	return nil, nil
}

func (m *mockRepo) GetByID(userID, id int) (*note.Note, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(userID, id)
	}
	return nil, note.ErrNotFound
}

func (m *mockRepo) Create(n *note.Note) error {
	if m.createFn != nil {
		return m.createFn(n)
	}
	return nil
}

func (m *mockRepo) Update(n *note.Note) error {
	if m.updateFn != nil {
		return m.updateFn(n)
	}
	return nil
}

func (m *mockRepo) Delete(userID, id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(userID, id)
	}
	return nil
}

func newTestHandler(repo note.Repository) *NotesHandler {
	return NewNotesHandler(note.NewService(repo))
}

func TestGetAll(t *testing.T) {
	tests := []struct {
		name       string
		repo       *mockRepo
		wantStatus int
		wantCount  int
	}{
		{
			name:       "empty list",
			repo:       &mockRepo{getAllFn: func(userID int) ([]*note.Note, error) { return []*note.Note{}, nil }},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name: "multiple notes",
			repo: &mockRepo{getAllFn: func(userID int) ([]*note.Note, error) {
				return []*note.Note{{ID: 1, UserID: userID, Title: "first"}, {ID: 2, UserID: userID, Title: "second"}}, nil
			}},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(tc.repo)
			req := httptest.NewRequest(http.MethodGet, "/notes", nil)
			w := httptest.NewRecorder()
			h.GetAll(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}

			var got []*note.Note
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if len(got) != tc.wantCount {
				t.Errorf("note count = %d, want %d", len(got), tc.wantCount)
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	tests := []struct {
		name       string
		pathID     string
		repo       *mockRepo
		wantStatus int
	}{
		{
			name:   "found",
			pathID: "1",
			repo: &mockRepo{getByIDFn: func(userID, id int) (*note.Note, error) {
				return &note.Note{ID: 1, UserID: userID, Title: "hello"}, nil
			}},
			wantStatus: http.StatusOK,
		},
		{
			name:   "not found",
			pathID: "99",
			repo: &mockRepo{getByIDFn: func(userID, id int) (*note.Note, error) {
				return nil, note.ErrNotFound
			}},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id",
			pathID:     "abc",
			repo:       &mockRepo{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(tc.repo)
			req := httptest.NewRequest(http.MethodGet, "/notes/"+tc.pathID, nil)
			req.SetPathValue("id", tc.pathID)
			w := httptest.NewRecorder()
			h.GetByID(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "success",
			body:       `{"title":"my note","body":"some content"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "empty title",
			body:       `{"title":"","body":"content"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			body:       `not json`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(&mockRepo{})
			req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.Create(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name       string
		pathID     string
		body       string
		repo       *mockRepo
		wantStatus int
	}{
		{
			name:   "success",
			pathID: "1",
			body:   `{"title":"updated title"}`,
			repo: &mockRepo{
				getByIDFn: func(userID, id int) (*note.Note, error) {
					return &note.Note{ID: 1, UserID: userID, Title: "original"}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "not found",
			pathID: "99",
			body:   `{"title":"updated"}`,
			repo: &mockRepo{
				getByIDFn: func(userID, id int) (*note.Note, error) {
					return nil, note.ErrNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id",
			pathID:     "xyz",
			body:       `{"title":"x"}`,
			repo:       &mockRepo{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(tc.repo)
			req := httptest.NewRequest(http.MethodPut, "/notes/"+tc.pathID, bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("id", tc.pathID)
			w := httptest.NewRecorder()
			h.Update(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name       string
		pathID     string
		repo       *mockRepo
		wantStatus int
	}{
		{
			name:       "success",
			pathID:     "1",
			repo:       &mockRepo{},
			wantStatus: http.StatusOK,
		},
		{
			name:   "not found",
			pathID: "99",
			repo: &mockRepo{deleteFn: func(userID, id int) error {
				return note.ErrNotFound
			}},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id",
			pathID:     "abc",
			repo:       &mockRepo{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(tc.repo)
			req := httptest.NewRequest(http.MethodDelete, "/notes/"+tc.pathID, nil)
			req.SetPathValue("id", tc.pathID)
			w := httptest.NewRecorder()
			h.Delete(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

func TestCreateListNoteHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantType   note.NoteType
		wantTodos  int
	}{
		{
			name:       "create list note with todos",
			body:       `{"type":"list","title":"shopping","body":"items to buy","todos":[{"text":"milk","completed":false},{"text":"bread","completed":true}]}`,
			wantStatus: http.StatusCreated,
			wantType:   note.NoteTypeList,
			wantTodos:  2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(&mockRepo{})
			req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.Create(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}

			var got note.Note
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Type != tc.wantType {
				t.Errorf("type = %q, want %q", got.Type, tc.wantType)
			}
			if len(got.Todos) != tc.wantTodos {
				t.Errorf("todos count = %d, want %d", len(got.Todos), tc.wantTodos)
			}
		})
	}
}

func TestGetListNoteHandler(t *testing.T) {
	tests := []struct {
		name       string
		pathID     string
		repo       *mockRepo
		wantStatus int
		wantType   note.NoteType
	}{
		{
			name:   "get list note returns todos",
			pathID: "1",
			repo: &mockRepo{getByIDFn: func(userID, id int) (*note.Note, error) {
				return &note.Note{
					ID:    1,
					UserID: userID,
					Title: "shopping",
					Type:  note.NoteTypeList,
					Todos: []note.Todo{
						{Text: "milk", Completed: false},
						{Text: "bread", Completed: true},
					},
				}, nil
			}},
			wantStatus: http.StatusOK,
			wantType:   note.NoteTypeList,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(tc.repo)
			req := httptest.NewRequest(http.MethodGet, "/notes/"+tc.pathID, nil)
			req.SetPathValue("id", tc.pathID)
			w := httptest.NewRecorder()
			h.GetByID(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}

			var got note.Note
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Type != tc.wantType {
				t.Errorf("type = %q, want %q", got.Type, tc.wantType)
			}
			if len(got.Todos) != 2 {
				t.Errorf("todos count = %d, want %d", len(got.Todos), 2)
			}
		})
	}
}

func TestUpdateListNoteReplaceTodos(t *testing.T) {
	tests := []struct {
		name       string
		pathID     string
		body       string
		repo       *mockRepo
		wantStatus int
		wantTodos  int
	}{
		{
			name:   "update list note replaces todos",
			pathID: "1",
			body:   `{"type":"list","title":"updated shopping","todos":[{"text":"eggs","completed":false}]}`,
			repo: &mockRepo{
				getByIDFn: func(userID, id int) (*note.Note, error) {
					return &note.Note{
						ID:    1,
						UserID: userID,
						Title: "shopping",
						Type:  note.NoteTypeList,
						Todos: []note.Todo{
							{Text: "milk", Completed: false},
							{Text: "bread", Completed: true},
						},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantTodos:  1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(tc.repo)
			req := httptest.NewRequest(http.MethodPut, "/notes/"+tc.pathID, bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("id", tc.pathID)
			w := httptest.NewRecorder()
			h.Update(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}

			var got note.Note
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if len(got.Todos) != tc.wantTodos {
				t.Errorf("todos count = %d, want %d", len(got.Todos), tc.wantTodos)
			}
		})
	}
}

func TestFullFlowTextAndListNotes(t *testing.T) {
	// Track created notes to simulate database state
	notes := make(map[int]*note.Note)
	nextID := 1

	// Create repo with full CRUD operations
	repo := &mockRepo{
		createFn: func(n *note.Note) error {
			n.ID = nextID
			notes[nextID] = n
			nextID++
			return nil
		},
		getByIDFn: func(userID, id int) (*note.Note, error) {
			n, ok := notes[id]
			if !ok {
				return nil, note.ErrNotFound
			}
			return n, nil
		},
		getAllFn: func(userID int) ([]*note.Note, error) {
			var result []*note.Note
			for _, n := range notes {
				if n.UserID == userID {
					result = append(result, n)
				}
			}
			return result, nil
		},
	}

	// Step 1: Create a text note
	h := newTestHandler(repo)
	textBody := `{"type":"text","title":"My Note","body":"This is a text note"}`
	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(textBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("create text note status = %d, want %d", w.Code, http.StatusCreated)
	}

	var textNote note.Note
	if err := json.NewDecoder(w.Body).Decode(&textNote); err != nil {
		t.Fatalf("decode text note response: %v", err)
	}

	if textNote.Type != note.NoteTypeText {
		t.Errorf("text note type = %q, want %q", textNote.Type, note.NoteTypeText)
	}
	if textNote.Title != "My Note" {
		t.Errorf("text note title = %q, want %q", textNote.Title, "My Note")
	}

	// Step 2: Create a list note
	listBody := `{"type":"list","title":"Shopping List","body":"Items to buy","todos":[{"text":"apples","completed":false},{"text":"milk","completed":false}]}`
	req = httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(listBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("create list note status = %d, want %d", w.Code, http.StatusCreated)
	}

	var listNote note.Note
	if err := json.NewDecoder(w.Body).Decode(&listNote); err != nil {
		t.Fatalf("decode list note response: %v", err)
	}

	// Step 3: Verify list note properties
	if listNote.Type != note.NoteTypeList {
		t.Errorf("list note type = %q, want %q", listNote.Type, note.NoteTypeList)
	}
	if listNote.Title != "Shopping List" {
		t.Errorf("list note title = %q, want %q", listNote.Title, "Shopping List")
	}
	if len(listNote.Todos) != 2 {
		t.Errorf("list note todos count = %d, want %d", len(listNote.Todos), 2)
	}
	if listNote.Todos[0].Text != "apples" {
		t.Errorf("first todo text = %q, want %q", listNote.Todos[0].Text, "apples")
	}
	if listNote.Todos[1].Text != "milk" {
		t.Errorf("second todo text = %q, want %q", listNote.Todos[1].Text, "milk")
	}

	// Step 4: Verify both notes coexist by retrieving them
	req = httptest.NewRequest(http.MethodGet, "/notes", nil)
	w = httptest.NewRecorder()
	h.GetAll(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("get all notes status = %d, want %d", w.Code, http.StatusOK)
	}

	var allNotes []*note.Note
	if err := json.NewDecoder(w.Body).Decode(&allNotes); err != nil {
		t.Fatalf("decode all notes response: %v", err)
	}

	if len(allNotes) != 2 {
		t.Errorf("total notes count = %d, want %d", len(allNotes), 2)
	}
}
