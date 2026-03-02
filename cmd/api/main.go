package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
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

	// Config from flags/env
	port := flag.String("port", os.Getenv("PORT"), "Server port")
	dbPath := flag.String("db", os.Getenv("DB_PATH"), "SQLite database path")
	jwtSecret := flag.String("secret", os.Getenv("JWT_SECRET"), "JWT secret key")
	corsOrigins := flag.String("cors", os.Getenv("CORS_ORIGINS"), "Comma-separated CORS allowed origins")
	flag.Parse()

	// Validate required config
	if *port == "" {
		log.Fatal("port is required: set PORT in .env or pass --port")
	}
	if *dbPath == "" {
		log.Fatal("db path is required: set DB_PATH in .env or pass --db")
	}
	if *jwtSecret == "" {
		log.Fatal("JWT secret is required: set JWT_SECRET in .env or pass --secret")
	}
	if *corsOrigins == "" {
		log.Fatal("CORS origins are required: set CORS_ORIGINS in .env or pass --cors")
	}

	// Parse comma-separated CORS origins
	allowedOrigins := strings.Split(strings.TrimSpace(*corsOrigins), ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	// Normalize port — accept both "3000" and ":3000"
	if !strings.HasPrefix(*port, ":") {
		*port = ":" + *port
	}

	// Init database
	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
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
	authHandler := handler.NewAuthHandler(userService, *jwtSecret)

	// Register routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/register", usersHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)

	// Protected routes (require authentication)
	protected := http.NewServeMux()
	protected.HandleFunc("GET /notes", notesHandler.GetAll)
	protected.HandleFunc("POST /notes", notesHandler.Create)
	protected.HandleFunc("GET /notes/{id}", notesHandler.GetByID)
	protected.HandleFunc("PUT /notes/{id}", notesHandler.Update)
	protected.HandleFunc("DELETE /notes/{id}", notesHandler.Delete)

	// Wrap protected routes with auth middleware
	mux.Handle("GET /notes", auth.Middleware(*jwtSecret)(protected))
	mux.Handle("POST /notes", auth.Middleware(*jwtSecret)(protected))
	mux.Handle("GET /notes/{id}", auth.Middleware(*jwtSecret)(protected))
	mux.Handle("PUT /notes/{id}", auth.Middleware(*jwtSecret)(protected))
	mux.Handle("DELETE /notes/{id}", auth.Middleware(*jwtSecret)(protected))

	// Chain middleware
	var handler http.Handler = mux
	handler = middleware.Logging(handler)
	handler = middleware.RequestID(handler)
	handler = middleware.CORS(allowedOrigins)(handler)

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
