package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/raworiginal/go-notes-api/internal/handler"
	"github.com/raworiginal/go-notes-api/internal/middleware"
	"github.com/raworiginal/go-notes-api/internal/note"
	"github.com/raworiginal/go-notes-api/internal/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Config from flags/env
	port := flag.String("port", ":8080", "Server port")
	dbPath := flag.String("db", "notes.db", "SQLite database path")
	flag.Parse()

	// Init database
	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	// Auto-migrate the Note Schema (creates tables if it doesn't exist)
	if err := db.AutoMigrate(&note.Note{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Dependecy injection chain
	noteStore := store.NewSQLiteStore(db)
	noteService := note.NewService(noteStore)
	notesHandler := handler.NewNotesHandler(noteService)

	// Register routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /notes", notesHandler.GetAll)
	mux.HandleFunc("POST /notes", notesHandler.Create)
	mux.HandleFunc("GET /notes/{id}", notesHandler.GetByID)
	mux.HandleFunc("PUT /notes/{id}", notesHandler.Update)
	mux.HandleFunc("DELETE /notes/{id}", notesHandler.Delete)

	// Chain middleware
	var handler http.Handler = mux
	handler = middleware.Logging(handler)
	handler = middleware.RequestID(handler)

	// Config server with Timeouts
	server := &http.Server{
		Addr:         *port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting server on port %v", *port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
