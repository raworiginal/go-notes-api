package store

import (
	"github.com/raworiginal/go-notes-api/internal/user"
	"gorm.io/gorm"
)

// runUserMigrations runs schema migrations for the User model.
// It ensures the table exists with proper indexes and constraints.
func runUserMigrations(db *gorm.DB) error {
	// Ensure the table exists (idempotent - safe to call even if table already exists)
	if err := db.AutoMigrate(&user.User{}); err != nil {
		return err
	}

	return nil
}
