package note

import (
	"testing"
)

// mockRepository is a mock implementation of Repository interface for testing
type mockRepository struct {
	notes map[int]*Note
	nextID int
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		notes: make(map[int]*Note),
		nextID: 1,
	}
}

func (m *mockRepository) GetAll(userID int) ([]*Note, error) {
	var result []*Note
	for _, note := range m.notes {
		if note.UserID == userID {
			result = append(result, note)
		}
	}
	return result, nil
}

func (m *mockRepository) GetByID(userID, id int) (*Note, error) {
	note, exists := m.notes[id]
	if !exists {
		return nil, ErrNotFound
	}
	if note.UserID != userID {
		return nil, ErrNotFound
	}
	return note, nil
}

func (m *mockRepository) Create(note *Note) error {
	note.ID = m.nextID
	m.notes[m.nextID] = note
	m.nextID++
	return nil
}

func (m *mockRepository) Update(note *Note) error {
	if _, exists := m.notes[note.ID]; !exists {
		return ErrNotFound
	}
	m.notes[note.ID] = note
	return nil
}

func (m *mockRepository) Delete(userID, id int) error {
	note, exists := m.notes[id]
	if !exists {
		return ErrNotFound
	}
	if note.UserID != userID {
		return ErrNotFound
	}
	delete(m.notes, id)
	return nil
}

// TestValidateNoteType tests that validateNoteType accepts valid types and rejects invalid ones
func TestValidateNoteType(t *testing.T) {
	tests := []struct {
		name    string
		noteType NoteType
		wantErr bool
	}{
		{
			name:    "valid text type",
			noteType: NoteTypeText,
			wantErr: false,
		},
		{
			name:    "valid list type",
			noteType: NoteTypeList,
			wantErr: false,
		},
		{
			name:    "invalid type",
			noteType: NoteType("invalid"),
			wantErr: true,
		},
		{
			name:    "empty type",
			noteType: NoteType(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNoteType(tt.noteType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNoteType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCreateTextNote tests that Service.CreateWithType properly creates a text note and preserves the type
func TestCreateTextNote(t *testing.T) {
	mockRepo := newMockRepository()
	service := NewService(mockRepo)

	title := "My Text Note"
	body := "This is the note body"

	note, err := service.CreateWithType(1, title, &body, NoteTypeText)
	if err != nil {
		t.Fatalf("CreateWithType() error = %v", err)
	}

	if note.Title != title {
		t.Errorf("Title = %q, want %q", note.Title, title)
	}

	if note.Type != NoteTypeText {
		t.Errorf("Type = %q, want %q", note.Type, NoteTypeText)
	}

	if note.Body == nil || *note.Body != body {
		t.Errorf("Body = %v, want %q", note.Body, body)
	}

	if note.UserID != 1 {
		t.Errorf("UserID = %d, want 1", note.UserID)
	}
}

// TestCreateListNoteWithTodos tests that Service.CreateWithType properly creates a list note with todos
func TestCreateListNoteWithTodos(t *testing.T) {
	mockRepo := newMockRepository()
	service := NewService(mockRepo)

	title := "My Todo List"
	todos := []Todo{
		{Text: "Task 1", Completed: false},
		{Text: "Task 2", Completed: true},
	}

	note, err := service.CreateWithType(1, title, nil, NoteTypeList, todos...)
	if err != nil {
		t.Fatalf("CreateWithType() error = %v", err)
	}

	if note.Title != title {
		t.Errorf("Title = %q, want %q", note.Title, title)
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

	if note.Todos[1].Text != "Task 2" {
		t.Errorf("Second todo text = %q, want %q", note.Todos[1].Text, "Task 2")
	}

	if note.UserID != 1 {
		t.Errorf("UserID = %d, want 1", note.UserID)
	}
}
