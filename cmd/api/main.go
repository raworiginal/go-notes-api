package main

import (
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/raworiginal/go-notes-api/internal/auth"
	"github.com/raworiginal/go-notes-api/internal/handler"
	"github.com/raworiginal/go-notes-api/internal/middleware"
	"github.com/raworiginal/go-notes-api/internal/note"
	"github.com/raworiginal/go-notes-api/internal/store"
	"github.com/raworiginal/go-notes-api/internal/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Load .env file (best-effort; no error if file absent)
	_ = godotenv.Load()

	// Load and validate configuration
	cfg := LoadConfig()

	// Init database
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	db.Exec("PRAGMA foreign_keys = ON")

	// Dependency injection chain
	noteStore := store.NewSQLiteNoteStore(db)
	noteService := note.NewService(noteStore)
	notesHandler := handler.NewNotesHandler(noteService)

	userStore := store.NewSQLiteUserStore(db)
	userService := user.NewService(userStore)
	usersHandler := handler.NewUsersHandler(userService)
	authHandler := handler.NewAuthHandler(userService, cfg.JWTSecret)

	// Register routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/register", usersHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Protected routes (require authentication)
	protected := http.NewServeMux()
	protected.HandleFunc("GET /notes", notesHandler.GetAll)
	protected.HandleFunc("POST /notes", notesHandler.Create)
	protected.HandleFunc("GET /notes/{id}", notesHandler.GetByID)
	protected.HandleFunc("PUT /notes/{id}", notesHandler.Update)
	protected.HandleFunc("DELETE /notes/{id}", notesHandler.Delete)

	// Wrap protected routes with auth middleware
	mux.Handle("GET /notes", auth.Middleware(cfg.JWTSecret)(protected))
	mux.Handle("POST /notes", auth.Middleware(cfg.JWTSecret)(protected))
	mux.Handle("GET /notes/{id}", auth.Middleware(cfg.JWTSecret)(protected))
	mux.Handle("PUT /notes/{id}", auth.Middleware(cfg.JWTSecret)(protected))
	mux.Handle("DELETE /notes/{id}", auth.Middleware(cfg.JWTSecret)(protected))

	// Chain middleware
	var handler http.Handler = mux
	handler = middleware.Logging(handler)
	handler = middleware.RequestID(handler)
	handler = middleware.CORS(cfg.CORSOrigins)(handler)

	// Config server with Timeouts
	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting server on port %v", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
