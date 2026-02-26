package user

import "errors"

var (
	ErrNotFound           = errors.New("user not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
