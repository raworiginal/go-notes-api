package note

import "fmt"

// Service holds the business logic for notes

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// validateNoteType validates that the NoteType is one of the valid types
func validateNoteType(t NoteType) error {
	switch t {
	case NoteTypeText, NoteTypeList:
		return nil
	default:
		return fmt.Errorf("%w: invalid note type %q", ErrInvalidInput, t)
	}
}

func (s *Service) GetAll(userID int) ([]*Note, error) {
	return s.repo.GetAll(userID)
}

func (s *Service) GetByID(userID, id int) (*Note, error) {
	return s.repo.GetByID(userID, id)
}

func (s *Service) Create(userID int, title string, body string) (*Note, error) {
	return s.CreateWithType(userID, title, body, NoteTypeText)
}

func (s *Service) CreateWithType(userID int, title string, body string, noteType NoteType, todos ...Todo) (*Note, error) {
	if title == "" {
		return nil, ErrInvalidInput
	}

	// Default type to text if empty
	if noteType == "" {
		noteType = NoteTypeText
	}

	// Validate the type
	if err := validateNoteType(noteType); err != nil {
		return nil, err
	}

	bodyPtr := &body
	note := &Note{
		UserID: userID,
		Title:  title,
		Body:   bodyPtr,
		Type:   noteType,
		Todos:  todos,
	}

	if err := s.repo.Create(note); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *Service) Update(userID, id int, title string, body *string) (*Note, error) {
	return s.UpdateWithType(userID, id, title, body, "")
}

func (s *Service) UpdateWithType(userID, id int, title string, body *string, noteType NoteType) (*Note, error) {
	note, err := s.repo.GetByID(userID, id)
	if err != nil {
		return nil, err
	}
	if title != "" {
		note.Title = title
	}
	if body != nil {
		note.Body = body
	}
	if noteType != "" {
		if err := validateNoteType(noteType); err != nil {
			return nil, err
		}
		note.Type = noteType
	}
	if err := s.repo.Update(note); err != nil {
		return nil, err
	}

	return note, nil
}

func (s *Service) Delete(userID, id int) error {
	return s.repo.Delete(userID, id)
}
