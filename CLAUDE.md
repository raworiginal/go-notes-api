# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

**Using Makefile (recommended):**
```bash
make build    # Build binary to bin/api
make run      # Build and run the server
make test     # Run all tests
make clean    # Remove binaries
make help     # Show all targets
```

**Direct Go commands:**
```bash
# Run all tests
go test ./...

# Run a single test file/package
go test ./internal/handler/
go test ./internal/store/
go test ./internal/user/

# Run tests with verbose output
go test -v ./...

# Build to bin/
go build -o bin/api ./cmd/api

# Run the server (loads config from .env)
./bin/api

# Add dependencies
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto
```

## Environment Setup

Copy `.env.example` to `.env` and configure:
```bash
PORT=8080
DB_PATH=notes.db
JWT_SECRET=<your-secret-key>
```

All three variables are required. See `cmd/api/main.go` for validation on startup.

## Architecture

This is a REST API for a notes app with user authentication, using idiomatic Go with `net/http` (no third-party router). Storage is SQLite via GORM. Authentication uses JWT tokens with 24-hour expiry.

**Layer flow:** `cmd/api/main.go` → wires `store` → `service` → `handler` and `middleware`

### `internal/user/` — User domain layer

- `model.go` — `User` struct with `ID`, `Username`, `Email`, `PasswordHash`, timestamps
- `repository.go` — `Repository` interface for user storage operations
- `service.go` — `Service` handles `Authenticate()`, `Create()`, user validation; uses `golang.org/x/crypto/bcrypt` for password hashing
- `errors.go` — sentinel errors `ErrInvalidCredentials`, `ErrUserExists`, `ErrInvalidInput`

### `internal/note/` — Note domain layer

- `model.go` — `Note` struct with `ID`, `Title`, `Body`, `CreatedAt`, `UpdatedAt`
- `repository.go` — `Repository` interface (abstraction decoupling business logic from storage)
- `service.go` — `Service` struct wrapping `Repository`; owns validation logic
- `errors.go` — sentinel errors `ErrNotFound` and `ErrInvalidInput`

### `internal/auth/` — Authentication

- `token.go` — JWT token generation and validation using `github.com/golang-jwt/jwt/v5`; `GenerateToken()` creates 24-hour tokens with `Claims` containing `UserID`, `Email`, `Username`
- `middleware.go` — HTTP middleware that validates JWT from `Authorization: Bearer <token>` header, injects `userID` into context via `userIDKey` (unexported type), provides `UserIDFromContext()` helper; returns 401 for missing/invalid/expired tokens

### `internal/store/` — Storage implementation

- `note_sqlite.go` — `SQLiteNoteStore` implements `note.Repository` using GORM; maps `gorm.ErrRecordNotFound` → `note.ErrNotFound`
- `note_migration.go` — creates `notes` table with `user_id` foreign key constraint
- `user_sqlite.go` — `SQLiteUserStore` implements `user.Repository`; `FindByEmail()` for authentication
- `user_migration.go` — creates `users` table with unique constraints on `username` and `email`

### `internal/handler/` — HTTP layer

- `notes.go` — `NotesHandler` holds `*note.Service`; uses Go 1.26 `ServeMux` with `{id}` path params and `r.PathValue("id")`; extracts `userID` from context via `auth.UserIDFromContext()` and filters notes by owner
- `auth.go` — `AuthHandler` handles `POST /login`: validates credentials, generates JWT token
- `users.go` — `UsersHandler` handles `POST /users/register`: creates new user with hashed password

### `internal/middleware/`

- `requestid.go` — injects UUID per request into context using unexported `contextKey` type
- `logging.go` — wraps `ResponseWriter` to capture status code; logs method, path, status, duration, request ID

### `cmd/api/main.go`

- Validates required config: `PORT`, `DB_PATH`, `JWT_SECRET` (fails fast if missing)
- Normalizes port (accepts both "3000" and ":3000")
- Enables SQLite foreign keys: `PRAGMA foreign_keys = ON`
- Chains middleware: `RequestID → Logging → mux`
- Routes: public (`/users/register`, `/login`) and protected (`/notes/*` routes wrapped with `auth.Middleware`)
- Sets `http.Server` timeouts (15s read/write, 60s idle)

## API Endpoints

### Public Routes
| Method | Path             | Request Body         | Response                 |
|--------|------------------|----------------------|--------------------------|
| POST   | /users/register  | `{username, email, password}` | `{id, username, email, created_at}` |
| POST   | /login           | `{email, password}`  | `{token}`                |

### Protected Routes (require `Authorization: Bearer <token>`)
| Method | Path             | Description                  |
|--------|------------------|------------------------------|
| GET    | /notes           | List authenticated user's notes |
| POST   | /notes           | Create a note for user       |
| GET    | /notes/{id}      | Get note by ID (owner check) |
| PUT    | /notes/{id}      | Update note (owner check)    |
| DELETE | /notes/{id}      | Delete note (owner check)    |

## Testing Approach

- `handler/notes_test.go` — table-driven tests using `httptest.NewRecorder()`
- `store/note_sqlite_test.go` — integration tests against in-memory SQLite (`:memory:`)
- Tests use `auth.GenerateToken()` for protected route testing

## Key Patterns

- Domain layer (`internal/note/`, `internal/user/`) has no external dependencies; `Repository` interfaces enable storage abstraction
- Error wrapping with `fmt.Errorf("%w", err)` throughout; handlers map domain errors → HTTP status using `errors.Is()`
- Middleware signature: `func(http.Handler) http.Handler`
- Context injection pattern: unexported `contextKey` type prevents key collisions (see `internal/auth/middleware.go` and `internal/middleware/requestid.go`)
- Protected routes: auth validation happens in middleware; handlers extract `userID` from context to scope queries
- Go module: `github.com/raworiginal/go-notes-api`
- Dependencies: `gorm.io/gorm`, `gorm.io/driver/sqlite`, `github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto`
