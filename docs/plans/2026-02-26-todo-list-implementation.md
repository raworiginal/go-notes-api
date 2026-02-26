# Todo List Feature Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add support for checklist-style notes alongside text notes, with todos stored as JSON and managed via the existing note endpoints.

**Architecture:** Extend the Note domain model with a `Type` field (enum: text/list) and `Todos` array. Store todos as JSON in SQLite. No new endpoints needed — use existing CRUD endpoints for both note types.

**Tech Stack:** Go, GORM (JSON column support), SQLite, net/http

---

## Task 1: Define Todo and NoteType in Domain Model

**Files:**
- Modify: `internal/note/model.go`
- Test: `internal/note/model_test.go` (create if needed)

**Step 1: Write failing test for Todo and NoteType**

Create `internal/note/model_test.go`:

```go
package note

import (
	"encoding/json"
	"testing"
)

func TestNoteTypeConstants(t *testing.T) {
	if NoteTypeText != "text" {
		t.Errorf("NoteTypeText = %s, want 'text'", NoteTypeText)
	}
	if NoteTypeList != "list" {
		t.Errorf("NoteTypeList = %s, want 'list'", NoteTypeList)
	}
}

func TestTodoJSONMarshal(t *testing.T) {
	todo := Todo{Text: "Buy milk", Completed: false}
	data, err := json.Marshal(todo)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var result Todo
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if result.Text != "Buy milk" || result.Completed != false {
		t.Errorf("Unmarshal mismatch: got %+v", result)
	}
}

func TestNoteWithTodos(t *testing.T) {
	note := &Note{
		ID:    1,
		Type:  NoteTypeList,
		Title: "Groceries",
		Todos: []Todo{
			{Text: "Milk", Completed: false},
			{Text: "Bread", Completed: true},
		},
	}

	if note.Type != NoteTypeList {
		t.Errorf("Type = %s, want %s", note.Type, NoteTypeList)
	}
	if len(note.Todos) != 2 {
		t.Errorf("Todos length = %d, want 2", len(note.Todos))
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd /home/raworiginal/Code/projects/go-notes-api
go test ./internal/note -v
```

Expected output: FAIL (undefined NoteType, Todo types)

**Step 3: Update model.go with NoteType and Todo**

Modify `internal/note/model.go` — add after existing imports:

```go
import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// NoteType represents the type of note (text or checklist)
type NoteType string

const (
	NoteTypeText NoteType = "text"
	NoteTypeList NoteType = "list"
)

// Todo represents a single todo item in a checklist
type Todo struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

// Update existing Note struct:
type Note struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Type      NoteType  `gorm:"type:text;default:'text'" json:"type"`
	Title     string    `json:"title"`
	Body      *string   `gorm:"type:text" json:"body,omitempty"`
	Todos     []Todo    `gorm:"type:json;serializer:json" json:"todos,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/note -v
```

Expected output: PASS

**Step 5: Commit**

```bash
git add internal/note/model.go internal/note/model_test.go
git commit -m "feat: add NoteType enum and Todo struct to domain model"
```

---

## Task 2: Update SQLiteNoteStore for New Columns

**Files:**
- Modify: `internal/store/note_sqlite.go`
- Test: `internal/store/sqlite_test.go`

**Step 1: Write failing test for storing list notes**

Add to `internal/store/sqlite_test.go`:

```go
func TestCreateListNote(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteNoteStore(db)

	userID := uint(1)

	newNote := &note.Note{
		UserID: userID,
		Type:   note.NoteTypeList,
		Title:  "Groceries",
		Todos: []note.Todo{
			{Text: "Milk", Completed: false},
			{Text: "Eggs", Completed: true},
		},
	}

	created, err := store.Create(newNote)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.Type != note.NoteTypeList {
		t.Errorf("Type = %s, want %s", created.Type, note.NoteTypeList)
	}
	if len(created.Todos) != 2 {
		t.Errorf("Todos count = %d, want 2", len(created.Todos))
	}
	if created.Todos[0].Text != "Milk" {
		t.Errorf("First todo text = %s, want 'Milk'", created.Todos[0].Text)
	}
}

func TestGetListNote(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteNoteStore(db)

	newNote := &note.Note{
		UserID: 1,
		Type:   note.NoteTypeList,
		Title:  "Todos",
		Todos: []note.Todo{
			{Text: "Task 1", Completed: false},
		},
	}

	created, _ := store.Create(newNote)

	fetched, err := store.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if fetched.Type != note.NoteTypeList {
		t.Errorf("Type mismatch: got %s", fetched.Type)
	}
	if len(fetched.Todos) != 1 {
		t.Errorf("Todos mismatch: got %d todos", len(fetched.Todos))
	}
}

func TestUpdateTodosReplace(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteNoteStore(db)

	original := &note.Note{
		UserID: 1,
		Type:   note.NoteTypeList,
		Title:  "List",
		Todos: []note.Todo{
			{Text: "Old 1", Completed: false},
			{Text: "Old 2", Completed: false},
		},
	}

	created, _ := store.Create(original)

	// Replace todos entirely
	created.Todos = []note.Todo{
		{Text: "New 1", Completed: true},
	}

	updated, err := store.Update(created)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if len(updated.Todos) != 1 {
		t.Errorf("Expected 1 todo after update, got %d", len(updated.Todos))
	}
	if updated.Todos[0].Text != "New 1" {
		t.Errorf("Todo text = %s, want 'New 1'", updated.Todos[0].Text)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/store -v -run "TestCreateListNote|TestGetListNote|TestUpdateTodosReplace"
```

Expected: FAIL (columns don't exist yet)

**Step 3: Add database migration**

Create migration in `internal/store/migration.go` (or add to existing migration logic):

```go
func runMigrations(db *gorm.DB) error {
	// Auto-migrate Note model (includes new Type and Todos fields)
	if err := db.AutoMigrate(&note.Note{}); err != nil {
		return err
	}

	// For existing SQLite databases, add columns if they don't exist
	// (AutoMigrate doesn't modify existing columns, but adds new ones)
	type NoteSchema struct {
		Type  string
		Todos string
	}

	if !db.Migrator().HasColumn(&note.Note{}, "type") {
		db.Migrator().AddColumn(&note.Note{}, "type")
		db.Model(&note.Note{}).Update("type", "text") // default existing notes to text
	}
	if !db.Migrator().HasColumn(&note.Note{}, "todos") {
		db.Migrator().AddColumn(&note.Note{}, "todos")
	}

	return nil
}
```

Update `internal/store/note_sqlite.go` initialization to call migrations:

```go
func NewSQLiteNoteStore(db *gorm.DB) *SQLiteNoteStore {
	// Run migrations to ensure schema is up-to-date
	if err := runMigrations(db); err != nil {
		// In production, this would return an error
		// For tests, we'll initialize it in setupTestDB
		panic(fmt.Sprintf("migration failed: %v", err))
	}
	return &SQLiteNoteStore{db: db}
}
```

Or if migrations are called elsewhere in `main.go`, ensure they run there.

**Step 4: Run test to verify it passes**

```bash
go test ./internal/store -v -run "TestCreateListNote|TestGetListNote|TestUpdateTodosReplace"
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/store/note_sqlite.go internal/store/sqlite_test.go
git commit -m "feat: add type and todos columns to note store"
```

---

## Task 3: Add Validation in Service Layer

**Files:**
- Modify: `internal/note/service.go`
- Test: `internal/note/service_test.go` (create if needed)

**Step 1: Write failing test for type validation**

Add to `internal/note/service_test.go`:

```go
package note

import (
	"testing"
)

func TestValidateNoteType(t *testing.T) {
	tests := []struct {
		name      string
		noteType  NoteType
		shouldErr bool
	}{
		{"text type", NoteTypeText, false},
		{"list type", NoteTypeList, false},
		{"invalid type", NoteType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNoteType(tt.noteType)
			if (err != nil) != tt.shouldErr {
				t.Errorf("validateNoteType(%s) err = %v, want err = %v", tt.noteType, err, tt.shouldErr)
			}
		})
	}
}

func TestCreateTextNote(t *testing.T) {
	// Mock repository
	repo := &mockRepository{notes: make(map[uint]*Note)}
	svc := NewService(repo)

	newNote := &Note{
		Type:  NoteTypeText,
		Title: "My note",
		Body:  stringPtr("Some text"),
		Todos: []Todo{}, // empty for text notes
	}

	created, err := svc.Create(newNote)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.Type != NoteTypeText {
		t.Errorf("Type = %s, want text", created.Type)
	}
}

func TestCreateListNoteWithTodos(t *testing.T) {
	repo := &mockRepository{notes: make(map[uint]*Note)}
	svc := NewService(repo)

	newNote := &Note{
		Type:  NoteTypeList,
		Title: "Groceries",
		Todos: []Todo{
			{Text: "Milk", Completed: false},
			{Text: "Bread", Completed: false},
		},
	}

	created, err := svc.Create(newNote)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.Type != NoteTypeList {
		t.Errorf("Type = %s, want list", created.Type)
	}
	if len(created.Todos) != 2 {
		t.Errorf("Todos count = %d, want 2", len(created.Todos))
	}
}

// Helper functions for testing
func stringPtr(s string) *string {
	return &s
}

type mockRepository struct {
	notes map[uint]*Note
	nextID uint
}

func (m *mockRepository) Create(n *Note) (*Note, error) {
	m.nextID++
	n.ID = m.nextID
	m.notes[n.ID] = n
	return n, nil
}

func (m *mockRepository) GetByID(id uint) (*Note, error) {
	if n, ok := m.notes[id]; ok {
		return n, nil
	}
	return nil, ErrNotFound
}

func (m *mockRepository) Update(n *Note) (*Note, error) {
	if _, ok := m.notes[n.ID]; !ok {
		return nil, ErrNotFound
	}
	m.notes[n.ID] = n
	return n, nil
}

func (m *mockRepository) Delete(id uint) error {
	delete(m.notes, id)
	return nil
}

func (m *mockRepository) List(userID uint) ([]*Note, error) {
	var notes []*Note
	for _, n := range m.notes {
		if n.UserID == userID {
			notes = append(notes, n)
		}
	}
	return notes, nil
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/note -v -run "TestValidateNoteType|TestCreateTextNote|TestCreateListNoteWithTodos"
```

Expected: FAIL (functions don't exist)

**Step 3: Implement validation in service.go**

Add to `internal/note/service.go`:

```go
// validateNoteType checks if the note type is valid
func validateNoteType(t NoteType) error {
	switch t {
	case NoteTypeText, NoteTypeList:
		return nil
	default:
		return fmt.Errorf("%w: invalid type '%s'", ErrInvalidInput, t)
	}
}

// Update Service.Create method to validate type
func (s *Service) Create(n *Note) (*Note, error) {
	if strings.TrimSpace(n.Title) == "" {
		return nil, ErrInvalidInput
	}

	// Validate note type
	if err := validateNoteType(n.Type); err != nil {
		return nil, err
	}

	// Default type to text if empty
	if n.Type == "" {
		n.Type = NoteTypeText
	}

	return s.repo.Create(n)
}

// Update Service.Update method similarly
func (s *Service) Update(n *Note) (*Note, error) {
	if strings.TrimSpace(n.Title) == "" {
		return nil, ErrInvalidInput
	}

	// Validate note type
	if err := validateNoteType(n.Type); err != nil {
		return nil, err
	}

	return s.repo.Update(n)
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/note -v -run "TestValidateNoteType|TestCreateTextNote|TestCreateListNoteWithTodos"
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/note/service.go internal/note/service_test.go
git commit -m "feat: add note type validation to service layer"
```

---

## Task 4: Update Handler to Parse and Return Todos

**Files:**
- Modify: `internal/handler/notes.go`
- Modify: `internal/handler/notes_test.go`

**Step 1: Write failing test for handler with todos**

Add to `internal/handler/notes_test.go`:

```go
func TestCreateListNoteHandler(t *testing.T) {
	repo := &mockNoteRepository{notes: make(map[uint]*note.Note)}
	svc := note.NewService(repo)
	handler := NewNotesHandler(svc)

	body := `{
		"type": "list",
		"title": "Shopping",
		"todos": [
			{"text": "Apples", "completed": false},
			{"text": "Oranges", "completed": true}
		]
	}`

	req := httptest.NewRequest("POST", "/notes", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp["type"] != "list" {
		t.Errorf("type = %s, want 'list'", resp["type"])
	}

	todos, ok := resp["todos"].([]interface{})
	if !ok || len(todos) != 2 {
		t.Errorf("todos parsing failed or count mismatch")
	}
}

func TestGetListNoteHandler(t *testing.T) {
	repo := &mockNoteRepository{notes: make(map[uint]*note.Note)}
	svc := note.NewService(repo)
	handler := NewNotesHandler(svc)

	// Create a list note first
	listNote := &note.Note{
		ID:    1,
		Type:  note.NoteTypeList,
		Title: "Todos",
		Todos: []note.Todo{
			{Text: "Task 1", Completed: false},
		},
	}
	repo.notes[1] = listNote

	req := httptest.NewRequest("GET", "/notes/1", nil)
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if resp["type"] != "list" {
		t.Errorf("type mismatch in response")
	}
}

func TestUpdateListNoteReplaceTodos(t *testing.T) {
	repo := &mockNoteRepository{notes: make(map[uint]*note.Note)}
	svc := note.NewService(repo)
	handler := NewNotesHandler(svc)

	// Create initial note with todos
	original := &note.Note{
		ID:    1,
		Type:  note.NoteTypeList,
		Title: "List",
		Todos: []note.Todo{
			{Text: "Old task", Completed: false},
		},
	}
	repo.notes[1] = original

	// Update with new todos
	body := `{
		"type": "list",
		"title": "Updated List",
		"todos": [
			{"text": "New task 1", "completed": true},
			{"text": "New task 2", "completed": false}
		]
	}`

	req := httptest.NewRequest("PUT", "/notes/1", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	todos, _ := resp["todos"].([]interface{})
	if len(todos) != 2 {
		t.Errorf("Expected 2 todos after update, got %d", len(todos))
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/handler -v -run "TestCreateListNoteHandler|TestGetListNoteHandler|TestUpdateListNoteReplaceTodos"
```

Expected: FAIL (type field not being parsed)

**Step 3: Update handler to parse Type and Todos**

Modify `internal/handler/notes.go` Create method:

```go
func (h *NotesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type  note.NoteType `json:"type"`
		Title string        `json:"title"`
		Body  *string       `json:"body,omitempty"`
		Todos []note.Todo   `json:"todos,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	n := &note.Note{
		Type:  req.Type,
		Title: req.Title,
		Body:  req.Body,
		Todos: req.Todos,
	}

	created, err := h.svc.Create(n)
	if err != nil {
		if errors.Is(err, note.ErrInvalidInput) {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}
```

Similarly update `Update` method to parse Type and Todos.

**Step 4: Run test to verify it passes**

```bash
go test ./internal/handler -v -run "TestCreateListNoteHandler|TestGetListNoteHandler|TestUpdateListNoteReplaceTodos"
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/handler/notes.go internal/handler/notes_test.go
git commit -m "feat: add todo parsing and response in notes handler"
```

---

## Task 5: Verify All Integration Tests Pass

**Files:**
- Test: `internal/store/sqlite_test.go`
- Test: `internal/handler/notes_test.go`

**Step 1: Run all existing tests to ensure no regressions**

```bash
go test ./... -v
```

Expected: All tests PASS (including new ones from previous tasks)

**Step 2: Add integration test for full flow**

Add to `internal/handler/notes_test.go`:

```go
func TestFullFlowTextAndListNotes(t *testing.T) {
	repo := &mockNoteRepository{notes: make(map[uint]*note.Note)}
	svc := note.NewService(repo)
	handler := NewNotesHandler(svc)

	// Create a text note
	textBody := `{
		"type": "text",
		"title": "My thoughts",
		"body": "Some reflections"
	}`
	req := httptest.NewRequest("POST", "/notes", strings.NewReader(textBody))
	w := httptest.NewRecorder()
	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Create text note failed with status %d", w.Code)
	}

	// Create a list note
	listBody := `{
		"type": "list",
		"title": "Todos",
		"todos": [
			{"text": "Task 1", "completed": false},
			{"text": "Task 2", "completed": false}
		]
	}`
	req = httptest.NewRequest("POST", "/notes", strings.NewReader(listBody))
	w = httptest.NewRecorder()
	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Create list note failed with status %d", w.Code)
	}

	var listResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listResp)

	if listResp["type"] != "list" {
		t.Error("List note type not preserved")
	}

	todos, ok := listResp["todos"].([]interface{})
	if !ok || len(todos) != 2 {
		t.Error("List todos not returned correctly")
	}
}
```

**Step 3: Run integration test**

```bash
go test ./internal/handler -v -run "TestFullFlowTextAndListNotes"
```

Expected: PASS

**Step 4: Run full test suite**

```bash
go test ./... -v
```

Expected: All tests PASS with no failures or regressions

**Step 5: Commit**

```bash
git add internal/handler/notes_test.go
git commit -m "test: add integration test for text and list notes"
```

---

## Task 6: Manual Testing with Server

**Files:**
- Run: `go run ./cmd/api`

**Step 1: Start the server**

```bash
go run ./cmd/api
```

Expected: Server starts on `http://localhost:8080`

**Step 2: Test creating a text note**

```bash
curl -X POST http://localhost:8080/notes \
  -H "Content-Type: application/json" \
  -d '{
    "type": "text",
    "title": "My Note",
    "body": "Some text here"
  }'
```

Expected: Returns created note with `"type": "text"` and `"body": "Some text here"`

**Step 3: Test creating a list note**

```bash
curl -X POST http://localhost:8080/notes \
  -H "Content-Type: application/json" \
  -d '{
    "type": "list",
    "title": "Groceries",
    "todos": [
      {"text": "Milk", "completed": false},
      {"text": "Eggs", "completed": true}
    ]
  }'
```

Expected: Returns created note with todos array populated

**Step 4: Test getting a list note**

Replace `{id}` with the ID from previous response:

```bash
curl http://localhost:8080/notes/{id}
```

Expected: Returns list note with todos intact

**Step 5: Test updating a list note (replace todos)**

```bash
curl -X PUT http://localhost:8080/notes/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "type": "list",
    "title": "Groceries Updated",
    "todos": [
      {"text": "Cheese", "completed": false}
    ]
  }'
```

Expected: Returns updated note with only 1 todo (old todos replaced)

**Step 6: Test listing all notes**

```bash
curl http://localhost:8080/notes
```

Expected: Returns array with both text and list notes

**Step 7: Verify no errors in server logs**

Check terminal where server is running — should see request logs with no errors

---

## Summary

This plan implements the todo list feature in 6 tasks:

1. **Domain Model** — Add NoteType enum and Todo struct with tests
2. **Storage** — Update SQLiteNoteStore to handle new columns and verify schema
3. **Service** — Add validation for note types
4. **Handler** — Parse Type and Todos from requests and return in responses
5. **Integration Testing** — Verify all tests pass together
6. **Manual Testing** — Test full API flow with curl

Each task:
- Writes failing tests first (TDD)
- Implements minimal code to pass
- Commits frequently (after each step)
- Focuses on single responsibility

No new endpoints needed. Existing CRUD routes handle both text and list notes seamlessly.

---

## Execution Notes

- GORM's `type:json;serializer:json` tag handles JSON serialization automatically
- Replace-all behavior for todos is default — just send new Todos array on update
- Backward compatibility: existing text notes default to `type='text'`
- No breaking changes to API structure — just new optional fields

