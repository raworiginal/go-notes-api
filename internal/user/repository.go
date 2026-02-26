package user

type Repository interface {
	Create(u *User) error
	GetByID(id int) (*User, error)
	GetByEmail(email string) (*User, error)
}
