package note

type Repository interface {
	GetAll(userID int) ([]*Note, error)
	GetByID(userID, id int) (*Note, error)
	Create(note *Note) error
	Update(note *Note) error
	Delete(userID, id int) error
}
