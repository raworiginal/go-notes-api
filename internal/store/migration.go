package store

import (
	"github.com/raworiginal/go-notes-api/internal/note"
	"gorm.io/gorm"
)

// runMigrations runs schema migrations for the Note model.
// It creates the table if it doesn't exist and adds new columns with proper defaults.
func runMigrations(db *gorm.DB) error {
	// First, ensure the table exists with AutoMigrate
	if err := db.AutoMigrate(&note.Note{}); err != nil {
		return err
	}

	// Add the type column if it doesn't exist with default value 'text'
	if !db.Migrator().HasColumn(&note.Note{}, "type") {
		if err := db.Migrator().AddColumn(&note.Note{}, "type"); err != nil {
			return err
		}
		// Set default value for existing rows
		if err := db.Model(&note.Note{}).Where("type = ?", "").Update("type", note.NoteTypeText).Error; err != nil {
			return err
		}
	}

	// Add the todos column if it doesn't exist
	if !db.Migrator().HasColumn(&note.Note{}, "todos") {
		if err := db.Migrator().AddColumn(&note.Note{}, "todos"); err != nil {
			return err
		}
	}

	return nil
}
