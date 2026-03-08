package store

import (
	"github.com/raworiginal/go-notes-api/internal/auth"
	"gorm.io/gorm"
)

func MigrationRefreshToken(db *gorm.DB) error {
	return db.AutoMigrate(&auth.RefreshToken{})
}
