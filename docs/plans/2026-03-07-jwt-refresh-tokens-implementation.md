# JWT Access & Refresh Token Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement a two-token authentication system with short-lived access tokens and persistent refresh tokens stored in the database.

**Architecture:** Extend the existing auth package with refresh token functions and database storage. Create a new RefreshToken model and repository following the existing GORM patterns. Add three new endpoints: modified login (returns both tokens), new refresh (validates and returns new access token), and new logout (revokes refresh token).

**Tech Stack:** Go 1.26, GORM with SQLite, golang-jwt/jwt v5, bcrypt for hashing

---

## Task 1: Update Config to Support Token Expiries

**Files:**
- Modify: `cmd/api/config.go`

**Step 1: Add token expiry fields to Config struct**

Open `cmd/api/config.go` and add these fields to the `Config` struct:
```go
type Config struct {
	// ... existing fields ...
	JWTSecret              string
	AccessTokenExpiry      time.Duration
	RefreshTokenExpiry     time.Duration
	// ... other fields ...
}
```

**Step 2: Update LoadConfig to parse token expiries from environment**

In the `LoadConfig()` function, add these lines after the existing JWT secret line:

```go
accessExpiry := os.Getenv("JWT_ACCESS_TOKEN_EXPIRY")
if accessExpiry == "" {
	accessExpiry = "15m"
}
cfg.AccessTokenExpiry, _ = time.ParseDuration(accessExpiry)

refreshExpiry := os.Getenv("JWT_REFRESH_TOKEN_EXPIRY")
if refreshExpiry == "" {
	refreshExpiry = "168h"
}
cfg.RefreshTokenExpiry, _ = time.ParseDuration(refreshExpiry)
```

**Step 3: Create or update `.env` file with defaults**

Add these lines to `.env`:
```
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h
```

**Step 4: Verify config loads correctly**

Run: `go build ./cmd/api`
Expected: No errors

**Step 5: Commit**

```bash
git add cmd/api/config.go .env
git commit -m "feat: add configurable JWT token expiries"
```

---

## Task 2: Create RefreshToken Model

**Files:**
- Create: `internal/auth/refresh_token.go`

**Step 1: Create RefreshToken model struct**

Create the file `internal/auth/refresh_token.go` with:

```go
package auth

import (
	"time"
)

// RefreshToken represents a refresh token stored in the database
type RefreshToken struct {
	ID        int       `gorm:"primaryKey"`
	UserID    int       `gorm:"index"`
	Token     string    `gorm:"uniqueIndex;index"`
	ExpiresAt time.Time
	RevokedAt *time.Time `gorm:"index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// IsValid checks if the refresh token is active and not expired
func (rt *RefreshToken) IsValid() bool {
	if rt.RevokedAt != nil {
		return false
	}
	return time.Now().Before(rt.ExpiresAt)
}

// Revoke marks the token as revoked
func (rt *RefreshToken) Revoke() {
	now := time.Now()
	rt.RevokedAt = &now
}
```

**Step 2: Verify file created**

Run: `go build ./internal/auth`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/auth/refresh_token.go
git commit -m "feat: add RefreshToken model"
```

---

## Task 3: Create RefreshToken Database Migration

**Files:**
- Create: `internal/store/refresh_token_migration.go`

**Step 1: Create migration file**

Create `internal/store/refresh_token_migration.go`:

```go
package store

import (
	"github.com/raworiginal/go-notes-api/internal/auth"
	"gorm.io/gorm"
)

// MigrateRefreshToken creates the refresh_tokens table
func MigrateRefreshToken(db *gorm.DB) error {
	return db.AutoMigrate(&auth.RefreshToken{})
}
```

**Step 2: Update main.go to run migration**

Open `cmd/api/main.go` and add this after `db.Exec("PRAGMA foreign_keys = ON")`:

```go
// Run migrations
if err := store.MigrateRefreshToken(db); err != nil {
	log.Fatalf("failed to migrate refresh token table: %v", err)
}
```

**Step 3: Verify migration compiles**

Run: `go build ./cmd/api`
Expected: No errors

**Step 4: Commit**

```bash
git add internal/store/refresh_token_migration.go cmd/api/main.go
git commit -m "feat: add RefreshToken database migration"
```

---

## Task 4: Create RefreshToken Repository

**Files:**
- Create: `internal/store/refresh_token_sqlite.go`

**Step 1: Create refresh token repository**

Create `internal/store/refresh_token_sqlite.go`:

```go
package store

import (
	"errors"

	"github.com/raworiginal/go-notes-api/internal/auth"
	"gorm.io/gorm"
)

// RefreshTokenStore defines operations for refresh tokens
type RefreshTokenStore interface {
	Create(rt *auth.RefreshToken) error
	GetByToken(token string) (*auth.RefreshToken, error)
	GetByUserID(userID int) (*auth.RefreshToken, error)
	Update(rt *auth.RefreshToken) error
	Delete(id int) error
}

// SQLiteRefreshTokenStore implements RefreshTokenStore
type SQLiteRefreshTokenStore struct {
	db *gorm.DB
}

// NewSQLiteRefreshTokenStore creates a new refresh token store
func NewSQLiteRefreshTokenStore(db *gorm.DB) *SQLiteRefreshTokenStore {
	return &SQLiteRefreshTokenStore{db}
}

// Create inserts a new refresh token
func (s *SQLiteRefreshTokenStore) Create(rt *auth.RefreshToken) error {
	if err := s.db.Create(rt).Error; err != nil {
		return err
	}
	return nil
}

// GetByToken retrieves a refresh token by its value
func (s *SQLiteRefreshTokenStore) GetByToken(token string) (*auth.RefreshToken, error) {
	var rt auth.RefreshToken
	if err := s.db.Where("token = ?", token).First(&rt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}
	return &rt, nil
}

// GetByUserID retrieves the active refresh token for a user
func (s *SQLiteRefreshTokenStore) GetByUserID(userID int) (*auth.RefreshToken, error) {
	var rt auth.RefreshToken
	if err := s.db.Where("user_id = ? AND revoked_at IS NULL", userID).First(&rt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}
	return &rt, nil
}

// Update saves changes to a refresh token
func (s *SQLiteRefreshTokenStore) Update(rt *auth.RefreshToken) error {
	return s.db.Save(rt).Error
}

// Delete removes a refresh token
func (s *SQLiteRefreshTokenStore) Delete(id int) error {
	return s.db.Delete(&auth.RefreshToken{}, id).Error
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/store`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/store/refresh_token_sqlite.go
git commit -m "feat: implement RefreshToken repository"
```

---

## Task 5: Add Token Hashing and Refresh Token Functions to Auth

**Files:**
- Modify: `internal/auth/token.go`

**Step 1: Add bcrypt import and hash function**

At the top of `internal/auth/token.go`, add to imports:
```go
"golang.org/x/crypto/bcrypt"
```

Then add this function before `GenerateToken`:

```go
// HashToken creates a bcrypt hash of the token
func HashToken(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CompareTokenHash verifies a token against its hash
func CompareTokenHash(hash, token string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
	return err == nil
}
```

**Step 2: Update GenerateToken to accept expiry duration**

Replace the `GenerateToken` function signature and body:

```go
// GenerateToken creates a signed JWT token with the specified expiry.
func GenerateToken(userID int, email, username, secret string, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID:           userID,
		Email:            email,
		Username:         username,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry))},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return token, nil
}
```

**Step 3: Add GenerateRefreshToken function**

Add this function after `GenerateToken`:

```go
// GenerateRefreshToken creates a random refresh token string
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
```

Add to imports at the top:
```go
"crypto/rand"
"encoding/base64"
```

**Step 4: Verify compilation**

Run: `go build ./internal/auth`
Expected: No errors

**Step 5: Commit**

```bash
git add internal/auth/token.go
git commit -m "feat: add token hashing and refresh token generation"
```

---

## Task 6: Update Login Handler

**Files:**
- Modify: `internal/handler/auth.go`

**Step 1: Add RefreshTokenStore to AuthHandler**

Update the `AuthHandler` struct:

```go
type AuthHandler struct {
	userService       *user.Service
	refreshTokenStore store.RefreshTokenStore
	jwtSecret         string
	accessExpiry      time.Duration
	refreshExpiry     time.Duration
}
```

**Step 2: Update NewAuthHandler constructor**

Replace the function:

```go
func NewAuthHandler(
	userService *user.Service,
	refreshTokenStore store.RefreshTokenStore,
	jwtSecret string,
	accessExpiry time.Duration,
	refreshExpiry time.Duration,
) *AuthHandler {
	return &AuthHandler{
		userService:       userService,
		refreshTokenStore: refreshTokenStore,
		jwtSecret:         jwtSecret,
		accessExpiry:      accessExpiry,
		refreshExpiry:     refreshExpiry,
	}
}
```

**Step 3: Update Login method to return both tokens**

Replace the `Login` method body after token generation:

```go
// Generate access token
accessToken, err := auth.GenerateToken(u.ID, u.Email, u.Username, h.jwtSecret, h.accessExpiry)
if err != nil {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"message": "Failed to generate access token"})
	return
}

// Generate refresh token
refreshTokenStr, err := auth.GenerateRefreshToken()
if err != nil {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"message": "Failed to generate refresh token"})
	return
}

// Hash and store refresh token
hashedToken, err := auth.HashToken(refreshTokenStr)
if err != nil {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"message": "Failed to hash refresh token"})
	return
}

rt := &auth.RefreshToken{
	UserID:    u.ID,
	Token:     hashedToken,
	ExpiresAt: time.Now().Add(h.refreshExpiry),
}

if err := h.refreshTokenStore.Create(rt); err != nil {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"message": "Failed to store refresh token"})
	return
}

// Return both tokens
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
	"accessToken":  accessToken,
	"refreshToken": refreshTokenStr,
})
```

**Step 4: Verify compilation**

Run: `go build ./internal/handler`
Expected: No errors

**Step 5: Commit**

```bash
git add internal/handler/auth.go
git commit -m "feat: update login to return access and refresh tokens"
```

---

## Task 7: Add Refresh Handler

**Files:**
- Modify: `internal/handler/auth.go`

**Step 1: Add Refresh method to AuthHandler**

Add this method to the `AuthHandler` struct:

```go
// Refresh method POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON"})
		return
	}

	// Get stored refresh token from database
	storedRT, err := h.refreshTokenStore.GetByToken(req.RefreshToken)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid refresh token"})
		return
	}

	// Check if token is valid (not revoked, not expired)
	if !storedRT.IsValid() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Refresh token expired or revoked"})
		return
	}

	// Verify the stored hash matches the provided token
	if !auth.CompareTokenHash(storedRT.Token, req.RefreshToken) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid refresh token"})
		return
	}

	// Get user to include in new access token
	u, err := h.userService.GetByID(storedRT.UserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "User not found"})
		return
	}

	// Generate new access token
	newAccessToken, err := auth.GenerateToken(u.ID, u.Email, u.Username, h.jwtSecret, h.accessExpiry)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Failed to generate token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"accessToken": newAccessToken})
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/handler`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/handler/auth.go
git commit -m "feat: add refresh endpoint handler"
```

---

## Task 8: Add Logout Handler

**Files:**
- Modify: `internal/handler/auth.go`

**Step 1: Add Logout method to AuthHandler**

Add this method to the `AuthHandler` struct:

```go
// Logout method POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract userID from context (set by auth middleware)
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "User not authenticated"})
		return
	}

	// Get user's refresh token
	rt, err := h.refreshTokenStore.GetByUserID(userID)
	if err != nil {
		// No token found is not an error - user might not have one
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
		return
	}

	// Revoke the refresh token
	rt.Revoke()
	if err := h.refreshTokenStore.Update(rt); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Failed to logout"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/handler`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/handler/auth.go
git commit -m "feat: add logout endpoint handler"
```

---

## Task 9: Register New Routes in main.go

**Files:**
- Modify: `cmd/api/main.go`

**Step 1: Initialize refresh token store**

In `main()`, after creating `authHandler`, add:

```go
refreshTokenStore := store.NewSQLiteRefreshTokenStore(db)
```

**Step 2: Update AuthHandler creation**

Replace the `authHandler` line with:

```go
authHandler := handler.NewAuthHandler(
	userService,
	refreshTokenStore,
	cfg.JWTSecret,
	cfg.AccessTokenExpiry,
	cfg.RefreshTokenExpiry,
)
```

**Step 3: Register new routes**

In the route registration section, add these routes after the login route:

```go
mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
mux.HandleFunc("POST /auth/logout", auth.Middleware(cfg.JWTSecret)(http.HandlerFunc(authHandler.Logout)))
```

**Step 4: Verify compilation**

Run: `go build ./cmd/api`
Expected: No errors

**Step 5: Commit**

```bash
git add cmd/api/main.go
git commit -m "feat: register refresh and logout endpoints"
```

---

## Task 10: Add User Service GetByID Method

**Files:**
- Modify: `internal/user/service.go`

**Step 1: Check if GetByID exists**

Open `internal/user/service.go` and look for a `GetByID` method. If it doesn't exist, add it:

```go
// GetByID retrieves a user by ID
func (s *Service) GetByID(id int) (*User, error) {
	return s.repository.GetByID(id)
}
```

**Step 2: Check if repository has GetByID**

If it doesn't exist, add it to `internal/user/repository.go`:

```go
// GetByID retrieves a user by ID
func (r *SQLiteUserRepository) GetByID(id int) (*User, error) {
	var u User
	if err := r.db.First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}
```

**Step 3: Verify compilation**

Run: `go build ./internal/user`
Expected: No errors

**Step 4: Commit**

```bash
git add internal/user/service.go internal/user/repository.go
git commit -m "feat: add GetByID method to user service"
```

---

## Task 11: Write Tests for Token Generation and Validation

**Files:**
- Create: `internal/auth/token_test.go`

**Step 1: Create test file**

Create `internal/auth/token_test.go`:

```go
package auth

import (
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	userID := 1
	email := "test@example.com"
	username := "testuser"
	secret := "test-secret"
	expiry := 15 * time.Minute

	token, err := GenerateToken(userID, email, username, secret, expiry)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Validate the token
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected Email %s, got %s", email, claims.Email)
	}
	if claims.Username != username {
		t.Errorf("expected Username %s, got %s", username, claims.Username)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token1, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	token2, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if token1 == token2 {
		t.Fatal("expected different tokens")
	}

	if token1 == "" || token2 == "" {
		t.Fatal("expected non-empty tokens")
	}
}

func TestHashAndCompareToken(t *testing.T) {
	token := "test-token-value"

	hash, err := HashToken(token)
	if err != nil {
		t.Fatalf("HashToken failed: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	if !CompareTokenHash(hash, token) {
		t.Fatal("expected token to match hash")
	}

	if CompareTokenHash(hash, "wrong-token") {
		t.Fatal("expected token to not match wrong hash")
	}
}

func TestExpiredToken(t *testing.T) {
	userID := 1
	email := "test@example.com"
	username := "testuser"
	secret := "test-secret"
	expiry := -1 * time.Minute // Already expired

	token, err := GenerateToken(userID, email, username, secret, expiry)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = ValidateToken(token, secret)
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}
```

**Step 2: Run tests**

Run: `go test ./internal/auth -v`
Expected: All tests pass

**Step 3: Commit**

```bash
git add internal/auth/token_test.go
git commit -m "test: add token generation and validation tests"
```

---

## Task 12: Write Tests for RefreshToken Model

**Files:**
- Create: `internal/auth/refresh_token_test.go`

**Step 1: Create test file**

Create `internal/auth/refresh_token_test.go`:

```go
package auth

import (
	"testing"
	"time"
)

func TestRefreshTokenIsValid(t *testing.T) {
	// Valid token
	rt := &RefreshToken{
		ID:        1,
		UserID:    1,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		RevokedAt: nil,
	}

	if !rt.IsValid() {
		t.Fatal("expected token to be valid")
	}

	// Expired token
	rt.ExpiresAt = time.Now().Add(-1 * time.Hour)
	if rt.IsValid() {
		t.Fatal("expected token to be expired")
	}

	// Revoked token
	rt.ExpiresAt = time.Now().Add(1 * time.Hour)
	now := time.Now()
	rt.RevokedAt = &now
	if rt.IsValid() {
		t.Fatal("expected token to be revoked")
	}
}

func TestRefreshTokenRevoke(t *testing.T) {
	rt := &RefreshToken{
		ID:        1,
		UserID:    1,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		RevokedAt: nil,
	}

	if rt.RevokedAt != nil {
		t.Fatal("expected RevokedAt to be nil initially")
	}

	rt.Revoke()

	if rt.RevokedAt == nil {
		t.Fatal("expected RevokedAt to be set")
	}

	if !time.Now().Before(rt.RevokedAt.Add(1 * time.Second)) {
		t.Fatal("expected RevokedAt to be recent")
	}
}
```

**Step 2: Run tests**

Run: `go test ./internal/auth -v`
Expected: All tests pass

**Step 3: Commit**

```bash
git add internal/auth/refresh_token_test.go
git commit -m "test: add refresh token model tests"
```

---

## Final Verification

**Step 1: Run all tests**

```bash
go test ./... -v
```

Expected: All tests pass

**Step 2: Build the application**

```bash
go build ./cmd/api
```

Expected: No errors

**Step 3: Verify database migrations**

The database will be automatically migrated on startup.

---

## Summary

You've now implemented:
- ✅ Configurable token expiries via `.env`
- ✅ RefreshToken model with revocation support
- ✅ Database migration and repository
- ✅ Token hashing and random generation
- ✅ Updated login to return both tokens
- ✅ Refresh endpoint to exchange refresh tokens
- ✅ Logout endpoint to revoke tokens
- ✅ Comprehensive tests
- ✅ Proper error handling and validation

The system now supports the two-token architecture with database-backed refresh tokens.
