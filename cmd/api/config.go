package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"
)

// Config holds all application configuration from env/flags.
type Config struct {
	Port               string
	DBPath             string
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	CORSOrigins        []string
}

// LoadConfig reads configuration from environment variables and flags.
// Flags override environment variables.
func LoadConfig() *Config {
	// Define flags with env var defaults
	port := flag.String("port", os.Getenv("PORT"), "Server port")
	dbPath := flag.String("db", os.Getenv("DB_PATH"), "SQLite database path")
	jwtSecret := flag.String("secret", os.Getenv("JWT_SECRET"), "JWT secret key")
	accessExpiry := os.Getenv("JWT_ACCESS_TOKEN_EXPIRY")
	if accessExpiry == "" {
		accessExpiry = "15m"
	}
	parsedAccessTokenExpiry, _ := time.ParseDuration(accessExpiry)
	refreshExpiry := os.Getenv("JWT_REFRESH_TOKEN_EXPIRY")
	if refreshExpiry == "" {
		refreshExpiry = "168h"
	}
	parsedRefreshTokenExpiry, _ := time.ParseDuration(refreshExpiry)
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

	// Normalize port
	if !strings.HasPrefix(*port, ":") {
		*port = ":" + *port
	}

	// Parse comma-separated CORS origins
	allowedOrigins := strings.Split(strings.TrimSpace(*corsOrigins), ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	return &Config{
		Port:               *port,
		DBPath:             *dbPath,
		JWTSecret:          *jwtSecret,
		CORSOrigins:        allowedOrigins,
		AccessTokenExpiry:  parsedAccessTokenExpiry,
		RefreshTokenExpiry: parsedRefreshTokenExpiry,
	}
}
