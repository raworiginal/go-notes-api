package store

import (
	"errors"

	"github.com/raworiginal/go-notes-api/internal/user"
	"gorm.io/gorm"
)

type SQLiteUserStore struct {
	db *gorm.DB
}

func NewSQLiteUserStore(db *gorm.DB) *SQLiteUserStore {
	return &SQLiteUserStore{db}
}

func (s *SQLiteUserStore) Create(u *user.User) error {
	return s.db.Create(u).Error
}

func (s *SQLiteUserStore) GetByID(id int) (*user.User, error) {
	var u user.User
	if err := s.db.First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *SQLiteUserStore) GetByEmail(email string) (*user.User, error) {
	var u user.User
	if err := s.db.Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
