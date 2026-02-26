// Package user for managing user data
package user

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Register creates a new user account. Validate inputs, check for duplicate
// email/username, hash the password with bcrypt, then persist.
func (s *Service) Register(username, email, password string) (*User, error) {
	if email == "" {
		return nil, fmt.Errorf("%w: email cannot be empty", ErrInvalidInput)
	}

	if username == "" {
		return nil, fmt.Errorf("%w: username cannot be empty", ErrInvalidInput)
	}
	if len(password) <= 7 {
		return nil, fmt.Errorf("%w: password must be at least 8 characters", ErrInvalidInput)
	}
	_, err := s.repo.GetByEmail(email)
	if err == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &User{
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) GetByID(id int) (*User, error) {
	return s.repo.GetByID(id)
}

// Authenticate verifies credentials and returns the user if valid.
// Used during JWT login (step 2).
func (s *Service) Authenticate(email, password string) (*User, error) {
	u, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}
