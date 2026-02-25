package note

// Service holds the business logic for notes

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAll() ([]*Note, error) {
	return s.repo.GetAll()
}

func (s *Service) GetByID(id int) (*Note, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(title string, body string) (*Note, error) {
	if title == "" {
		return nil, ErrInvalidInput
	}

	note := &Note{
		Title: title,
		Body:  body,
	}

	if err := s.repo.Create(note); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *Service) Update(id int, title string, body *string) (*Note, error) {
	note, err := s.repo.GetByID(id)
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

func (s *Service) Delete(id int) error {
	return s.repo.Delete(id)
}
