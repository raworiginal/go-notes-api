// Package store handles the interactions with the sqlite db
package store

import (
	"errors"

	"github.com/raworiginal/go-notes-api/internal/note"
	"gorm.io/gorm"
)

type SQLiteStore struct {
	db *gorm.DB
}

func NewSQLiteStore(db *gorm.DB) *SQLiteStore {
	return &SQLiteStore{db}
}

func (s *SQLiteStore) GetByID(id int) (*note.Note, error) {
	var n note.Note
	if err := s.db.First(&n, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, note.ErrNotFound
		}
		return nil, err
	}
	return &n, nil
}

func (s *SQLiteStore) GetAll() ([]*note.Note, error) {
	var notes []*note.Note
	if err := s.db.Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}

func (s *SQLiteStore) Update(n *note.Note) error {
	result := s.db.Save(n)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return note.ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) Create(n *note.Note) error {
	result := s.db.Create(n)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *SQLiteStore) Delete(id int) error {
	result := s.db.Delete(&note.Note{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
