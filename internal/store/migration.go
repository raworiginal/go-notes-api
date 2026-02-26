package store

import (
	"github.com/raworiginal/go-notes-api/internal/note"
	"gorm.io/gorm"
)

// runMigrations runs schema migrations for the Note model.
// It ensures the table exists and adds new columns with proper defaults.
func runMigrations(db *gorm.DB) error {
	// Ensure the table exists (idempotent - safe to call even if table already exists)
	if err := db.AutoMigrate(&note.Note{}); err != nil {
		return err
	}

	// Add type column if it doesn't exist
	if !db.Migrator().HasColumn(&note.Note{}, "type") {
		if err := db.Migrator().AddColumn(&note.Note{}, "type"); err != nil {
			return err
		}
		// Set default type for existing rows (handle both NULL and empty string)
		if err := db.Model(&note.Note{}).
			Where("type = ? OR type IS NULL", "").
			Update("type", note.NoteTypeText).Error; err != nil {
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
