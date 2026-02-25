# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests
go test ./...

# Run a single test file/package
go test ./internal/handler/
go test ./internal/store/

# Run tests with verbose output
go test -v ./...

# Build
go build ./cmd/api

# Run the server
go run ./cmd/api

# Add dependencies
go get gorm.io/gorm
go get gorm.io/driver/sqlite
```

## Architecture

This is a REST API for a notes app using idiomatic Go with `net/http` standard library (no third-party router). Storage is SQLite via GORM.

**Layer flow:** `cmd/api/main.go` → wires `store` → `service` → `handler`

### `internal/note/` — Domain layer (no external imports)

- `model.go` — `Note` struct with `ID`, `Title`, `Body`, `CreatedAt`, `UpdatedAt`
- `repository.go` — `Repository` interface (the key abstraction decoupling business logic from storage)
- `service.go` — `Service` struct wrapping `Repository`; owns all validation logic
- `errors.go` — sentinel errors `ErrNotFound` and `ErrInvalidInput`

### `internal/store/` — Storage implementation

- `sqlite.go` — `SQLiteStore` implements `note.Repository` using GORM
- Maps `gorm.ErrRecordNotFound` → `note.ErrNotFound` to keep domain errors clean

### `internal/handler/` — HTTP layer

- `notes.go` — `NotesHandler` holds `*note.Service`; uses Go 1.26 `ServeMux` with `{id}` path params and `r.PathValue("id")`
- Maps domain errors → HTTP status codes using `errors.Is`: `ErrNotFound`→404, `ErrInvalidInput`→400

### `internal/middleware/`

- `requestid.go` — injects UUID per request into `context.Context` using unexported `contextKey` type
- `logging.go` — wraps `ResponseWriter` to capture status code; logs method, path, status, duration, request ID

### `cmd/api/main.go`

- Reads config from env vars/flags (port, DB path)
- Chains middleware: `RequestID → Logging → mux`
- Sets `http.Server` timeouts (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`)

## API Endpoints

| Method | Path          | Description    |
|--------|---------------|----------------|
| GET    | /notes        | List all notes |
| POST   | /notes        | Create a note  |
| GET    | /notes/{id}   | Get by ID      |
| PUT    | /notes/{id}   | Update         |
| DELETE | /notes/{id}   | Delete         |

## Testing Approach

- `handler/notes_test.go` — table-driven tests using `httptest.NewRecorder()`
- `store/sqlite_test.go` — integration tests against an in-memory SQLite DB (`:memory:`)

## Key Patterns

- Repository interface in `note/` package; concrete implementation in `store/` — never import `store` from `note`
- Error wrapping with `fmt.Errorf("%w", err)` throughout service and store layers
- Middleware signature: `func(http.Handler) http.Handler`
- Go module: `github.com/raworiginal/go-notes-api`
