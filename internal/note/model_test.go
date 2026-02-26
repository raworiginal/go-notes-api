package note

import (
	"encoding/json"
	"testing"
	"time"
)

// TestNoteTypeConstants verifies NoteType constants are correct
func TestNoteTypeConstants(t *testing.T) {
	if NoteTypeText != "text" {
		t.Errorf("NoteTypeText = %q, want %q", NoteTypeText, "text")
	}
	if NoteTypeList != "list" {
		t.Errorf("NoteTypeList = %q, want %q", NoteTypeList, "list")
	}
}

// TestTodoJSONMarshal verifies Todo JSON serialization/deserialization
func TestTodoJSONMarshal(t *testing.T) {
	todo := Todo{
		Text:      "Buy groceries",
		Completed: false,
	}

	// Marshal to JSON
	data, err := json.Marshal(todo)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal back
	var unmarshalled Todo
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshalled.Text != todo.Text {
		t.Errorf("Text = %q, want %q", unmarshalled.Text, todo.Text)
	}
	if unmarshalled.Completed != todo.Completed {
		t.Errorf("Completed = %v, want %v", unmarshalled.Completed, todo.Completed)
	}
}

// TestNoteWithTodos verifies Note struct can hold Todos array
func TestNoteWithTodos(t *testing.T) {
	todos := []Todo{
		{Text: "Task 1", Completed: true},
		{Text: "Task 2", Completed: false},
	}

	note := Note{
		ID:        1,
		UserID:    1,
		Title:     "My List",
		Type:      NoteTypeList,
		Todos:     todos,
		Body:      nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if note.Type != NoteTypeList {
		t.Errorf("Type = %q, want %q", note.Type, NoteTypeList)
	}

	if len(note.Todos) != 2 {
		t.Errorf("Todos length = %d, want 2", len(note.Todos))
	}

	if note.Todos[0].Text != "Task 1" {
		t.Errorf("First todo text = %q, want %q", note.Todos[0].Text, "Task 1")
	}

	if !note.Todos[0].Completed {
		t.Errorf("First todo completed = %v, want true", note.Todos[0].Completed)
	}

	if note.Todos[1].Text != "Task 2" {
		t.Errorf("Second todo text = %q, want %q", note.Todos[1].Text, "Task 2")
	}

	if note.Todos[1].Completed {
		t.Errorf("Second todo completed = %v, want false", note.Todos[1].Completed)
	}
}
