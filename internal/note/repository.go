package note

type Repository interface {
	GetAll() ([]*Note, error)
	GetByID(id int) (*Note, error)
	Create(note *Note) error
	Update(note *Note) error
	Delete(id int) error
}
