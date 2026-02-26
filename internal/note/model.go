// Package note is for managing data for notes
package note

import "time"

// NoteType represents the type of note
type NoteType string

const (
	NoteTypeText NoteType = "text"
	NoteTypeList NoteType = "list"
)

// Todo represents a single todo item in a list note
type Todo struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

type Note struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id" gorm:"index"`
	Title     string    `json:"title"`
	Body      *string   `json:"body"`
	Type      NoteType  `json:"type"`
	Todos     []Todo    `json:"todos" gorm:"type:json;serializer:json"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
