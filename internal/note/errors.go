package note

import "errors"

var (
	ErrNotFound     = errors.New("note not found")
	ErrInvalidInput = errors.New("invalid input")
)
