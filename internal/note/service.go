package note

// Service holds the business logic for notes

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAll(userID int) ([]*Note, error) {
	return s.repo.GetAll(userID)
}

func (s *Service) GetByID(userID, id int) (*Note, error) {
	return s.repo.GetByID(userID, id)
}

func (s *Service) Create(userID int, title string, body string) (*Note, error) {
	if title == "" {
		return nil, ErrInvalidInput
	}

	note := &Note{
		UserID: userID,
		Title:  title,
		Body:   body,
	}

	if err := s.repo.Create(note); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *Service) Update(userID, id int, title string, body *string) (*Note, error) {
	note, err := s.repo.GetByID(userID, id)
	if err != nil {
		return nil, err
	}
	if title != "" {
		note.Title = title
	}
	if body != nil {
		note.Body = *body
	}
	if err := s.repo.Update(note); err != nil {
		return nil, err
	}

	return note, nil
}

func (s *Service) Delete(userID, id int) error {
	return s.repo.Delete(userID, id)
}
