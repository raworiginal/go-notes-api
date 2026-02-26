package store

import (
	"errors"
	"testing"

	"github.com/raworiginal/go-notes-api/internal/note"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB opens a fresh in-memory SQLite database and migrates the
// Note schema. Each test gets its own DB so they are fully isolated.
func setupTestDB(t *testing.T) *SQLiteNoteStore {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := runMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return &SQLiteNoteStore{db}
}

func TestSQLiteNoteStore_GetAll(t *testing.T) {
	s := setupTestDB(t)
	const userID = 1

	notes, err := s.GetAll(userID)
	if err != nil {
		t.Fatalf("GetAll on empty db: %v", err)
	}
	if len(notes) != 0 {
		t.Errorf("want 0 notes, got %d", len(notes))
	}

	_ = s.Create(&note.Note{UserID: userID, Title: "first"})
	_ = s.Create(&note.Note{UserID: userID, Title: "second"})

	notes, err = s.GetAll(userID)
	if err != nil {
		t.Fatalf("GetAll after inserts: %v", err)
	}
	if len(notes) != 2 {
		t.Errorf("want 2 notes, got %d", len(notes))
	}
}

func TestSQLiteNoteStore_GetByID(t *testing.T) {
	s := setupTestDB(t)
	const userID = 1

	bodyText := "body text"
	n := &note.Note{UserID: userID, Title: "find me", Body: &bodyText}
	if err := s.Create(n); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByID(userID, n.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != n.Title {
		t.Errorf("title = %q, want %q", got.Title, n.Title)
	}
	if (got.Body == nil && n.Body != nil) || (got.Body != nil && n.Body == nil) || (got.Body != nil && n.Body != nil && *got.Body != *n.Body) {
		var gotBody, wantBody string
		if got.Body != nil {
			gotBody = *got.Body
		}
		if n.Body != nil {
			wantBody = *n.Body
		}
		t.Errorf("body = %q, want %q", gotBody, wantBody)
	}

	// Non-existent record should map to ErrNotFound, not a raw GORM error.
	_, err = s.GetByID(userID, 9999)
	if !errors.Is(err, note.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestSQLiteNoteStore_Create(t *testing.T) {
	s := setupTestDB(t)
	const userID = 1

	content := "content"
	n := &note.Note{UserID: userID, Title: "new note", Body: &content}
	if err := s.Create(n); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if n.ID == 0 {
		t.Error("expected GORM to populate a non-zero ID after create")
	}

	// Verify the record was actually persisted.
	got, err := s.GetByID(userID, n.ID)
	if err != nil {
		t.Fatalf("GetByID after create: %v", err)
	}
	gotBody := ""
	if got.Body != nil {
		gotBody = *got.Body
	}
	wantBody := ""
	if n.Body != nil {
		wantBody = *n.Body
	}
	if got.Title != n.Title || gotBody != wantBody {
		t.Errorf("persisted note = {%q, %q}, want {%q, %q}", got.Title, gotBody, n.Title, wantBody)
	}
}

func TestSQLiteNoteStore_Update(t *testing.T) {
	s := setupTestDB(t)
	const userID = 1

	oldBody := "old body"
	n := &note.Note{UserID: userID, Title: "original", Body: &oldBody}
	if err := s.Create(n); err != nil {
		t.Fatalf("Create: %v", err)
	}

	n.Title = "updated"
	newBody := "new body"
	n.Body = &newBody
	if err := s.Update(n); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := s.GetByID(userID, n.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Title != "updated" {
		t.Errorf("title = %q, want %q", got.Title, "updated")
	}
	gotBody := ""
	if got.Body != nil {
		gotBody = *got.Body
	}
	if gotBody != "new body" {
		t.Errorf("body = %q, want %q", gotBody, "new body")
	}
}

func TestSQLiteNoteStore_Delete(t *testing.T) {
	s := setupTestDB(t)
	const userID = 1

	n := &note.Note{UserID: userID, Title: "to delete"}
	if err := s.Create(n); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := s.Delete(userID, n.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := s.GetByID(userID, n.ID)
	if !errors.Is(err, note.ErrNotFound) {
		t.Errorf("after delete, err = %v, want ErrNotFound", err)
	}
}

func TestCreateListNote(t *testing.T) {
	s := setupTestDB(t)
	const userID = 1

	todos := []note.Todo{
		{Text: "Buy milk", Completed: false},
		{Text: "Write tests", Completed: true},
	}

	n := &note.Note{
		UserID: userID,
		Title:  "Shopping list",
		Type:   note.NoteTypeList,
		Todos:  todos,
	}

	if err := s.Create(n); err != nil {
		t.Fatalf("Create list note: %v", err)
	}

	if n.ID == 0 {
		t.Error("expected GORM to populate a non-zero ID after create")
	}

	// Verify the list note was persisted with todos
	got, err := s.GetByID(userID, n.ID)
	if err != nil {
		t.Fatalf("GetByID after create: %v", err)
	}

	if got.Type != note.NoteTypeList {
		t.Errorf("type = %q, want %q", got.Type, note.NoteTypeList)
	}

	if len(got.Todos) != len(todos) {
		t.Errorf("todos count = %d, want %d", len(got.Todos), len(todos))
	}

	for i, todo := range got.Todos {
		if todo.Text != todos[i].Text || todo.Completed != todos[i].Completed {
			t.Errorf("todo[%d] = {%q, %v}, want {%q, %v}", i, todo.Text, todo.Completed, todos[i].Text, todos[i].Completed)
		}
	}
}

func TestGetListNote(t *testing.T) {
	s := setupTestDB(t)
	const userID = 1

	todos := []note.Todo{
		{Text: "First task", Completed: false},
		{Text: "Second task", Completed: false},
		{Text: "Third task", Completed: true},
	}

	n := &note.Note{
		UserID: userID,
		Title:  "Project checklist",
		Type:   note.NoteTypeList,
		Todos:  todos,
	}

	if err := s.Create(n); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByID(userID, n.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.Type != note.NoteTypeList {
		t.Errorf("type = %q, want %q", got.Type, note.NoteTypeList)
	}

	if len(got.Todos) != len(todos) {
		t.Errorf("todos count = %d, want %d", len(got.Todos), len(todos))
	}

	for i, todo := range got.Todos {
		if todo.Text != todos[i].Text || todo.Completed != todos[i].Completed {
			t.Errorf("todo[%d] = {%q, %v}, want {%q, %v}", i, todo.Text, todo.Completed, todos[i].Text, todos[i].Completed)
		}
	}
}

func TestUpdateTodosReplace(t *testing.T) {
	s := setupTestDB(t)
	const userID = 1

	initialTodos := []note.Todo{
		{Text: "Old task 1", Completed: false},
		{Text: "Old task 2", Completed: true},
	}

	n := &note.Note{
		UserID: userID,
		Title:  "Editable list",
		Type:   note.NoteTypeList,
		Todos:  initialTodos,
	}

	if err := s.Create(n); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Update with new todos
	newTodos := []note.Todo{
		{Text: "New task 1", Completed: true},
		{Text: "New task 2", Completed: false},
		{Text: "New task 3", Completed: false},
	}
	n.Todos = newTodos

	if err := s.Update(n); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// Verify the update
	got, err := s.GetByID(userID, n.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}

	if len(got.Todos) != len(newTodos) {
		t.Errorf("todos count = %d, want %d", len(got.Todos), len(newTodos))
	}

	for i, todo := range got.Todos {
		if todo.Text != newTodos[i].Text || todo.Completed != newTodos[i].Completed {
			t.Errorf("todo[%d] = {%q, %v}, want {%q, %v}", i, todo.Text, todo.Completed, newTodos[i].Text, newTodos[i].Completed)
		}
	}
}
