# Plan: Go Notes API

## Context

Building a REST API for a notes app from scratch in idiomatic Go. The goal is to practice
Go patterns through deliberate structure choices — not just getting it working, but getting
it *right*. The project is currently an empty directory at `/home/raworiginal/Code/projects/go-notes-api`.

Go version: 1.26 (enables modern ServeMux routing with `{id}` path params and `r.PathValue`)

User preferences:
- Router: `net/http` standard library only
- Storage: SQLite via GORM
- Patterns to practice: interface-based design, error handling, middleware/context, testing

---

## Project Structure

```
go-notes-api/
├── cmd/
│   └── api/
│       └── main.go              # Wire everything together, start server
├── internal/
│   ├── note/
│   │   ├── model.go             # Note struct, domain types
│   │   ├── repository.go        # Repository interface (the key abstraction)
│   │   ├── service.go           # Business logic layer
│   │   └── errors.go            # Sentinel errors (ErrNotFound, ErrInvalidInput)
│   ├── handler/
│   │   ├── notes.go             # HTTP handlers, JSON encode/decode, error mapping
│   │   └── notes_test.go        # Table-driven tests using httptest
│   ├── middleware/
│   │   ├── logging.go           # Request logging (method, path, duration, status)
│   │   └── requestid.go         # Inject request ID into context
│   └── store/
│       ├── sqlite.go            # SQLite implementation of note.Repository
│       └── sqlite_test.go       # Integration tests against real SQLite DB
├── go.mod
└── go.sum
```

---

## Implementation Plan

### Step 1 — Initialize module and dependencies

```bash
go mod init github.com/raworiginal/go-notes-api
go get -u gorm.io/gorm
go get -u gorm.io/driver/sqlite
```

### Step 2 — Domain layer (`internal/note/`)

**`model.go`** — the `Note` struct:
```go
type Note struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Body      string    `json:"body"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**`errors.go`** — sentinel errors for domain logic:
```go
var ErrNotFound    = errors.New("note not found")
var ErrInvalidInput = errors.New("invalid input")
```

**`repository.go`** — interface that decouples business logic from storage:
```go
type Repository interface {
    Create(ctx context.Context, n *Note) error
    GetByID(ctx context.Context, id int) (*Note, error)
    List(ctx context.Context) ([]*Note, error)
    Update(ctx context.Context, n *Note) error
    Delete(ctx context.Context, id int) error
}
```

**`service.go`** — wraps repository, owns validation logic:
```go
type Service struct { repo Repository }
func NewService(repo Repository) *Service
```
> **User contribution point**: Implement validation logic in `service.go`'s `Create` and `Update` methods — what makes a note valid? How should errors be wrapped and returned?

### Step 3 — Storage layer (`internal/store/`)

**`note_sqlite.go`** — `SQLiteNoteStore` struct implementing `note.Repository`:
- Initializes GORM DB with SQLite dialect: `gorm.Open(sqlite.Open(path), &gorm.Config{})`
- Runs migrations on startup using `db.AutoMigrate(&Note{})`
- Implements all 5 interface methods using GORM's query builder
- Maps `gorm.ErrRecordNotFound` → `note.ErrNotFound` for clean domain errors

> **User contribution point**: Implement the `List` method — decide if/how to support ordering or scoping (e.g., only non-deleted records).

### Step 4 — Middleware (`internal/middleware/`)

**`requestid.go`** — generates a UUID per request, stores it in `context.Context`:
```go
type contextKey string
const RequestIDKey contextKey = "requestID"

func RequestID(next http.Handler) http.Handler { ... }
```

**`logging.go`** — wraps a `ResponseWriter` to capture status code, logs after request:
```go
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}
func Logging(next http.Handler) http.Handler { ... }
```

> **User contribution point**: Implement `Logging` middleware in `logging.go` — include the request ID from context, method, path, status code, and duration.

### Step 5 — HTTP Handlers (`internal/handler/`)

**`notes.go`** — `NotesHandler` struct holding a `*note.Service`:
- Uses Go 1.26 ServeMux patterns: `GET /notes`, `POST /notes`, `GET /notes/{id}`, etc.
- `r.PathValue("id")` to extract path params
- Helper `writeJSON` / `readJSON` for encoding/decoding
- Helper `errorResponse` to map domain errors → HTTP status codes:
  - `note.ErrNotFound` → 404
  - `note.ErrInvalidInput` → 400
  - other → 500

> **User contribution point**: Implement `errorResponse` — mapping domain errors to HTTP responses is a critical design decision. Use `errors.Is` for sentinel error detection.

### Step 6 — Entry point (`cmd/api/main.go`)

- Parse config (port, DB path) from env vars or flags
- Create store, service, handler
- Register routes on `http.NewServeMux()`
- Chain middleware: `RequestID → Logging → mux`
- Start `http.Server` with timeouts (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`)

### Step 7 — Tests

**`handler/notes_test.go`** — table-driven tests using `httptest.NewRecorder()`:
```go
tests := []struct {
    name       string
    method     string
    path       string
    body       string
    wantStatus int
}{...}
```

**`store/sqlite_test.go`** — integration tests against an in-memory SQLite DB.

---

## API Endpoints

| Method | Path          | Description        |
|--------|---------------|--------------------|
| GET    | /notes        | List all notes     |
| POST   | /notes        | Create a note      |
| GET    | /notes/{id}   | Get note by ID     |
| PUT    | /notes/{id}   | Update a note      |
| DELETE | /notes/{id}   | Delete a note      |

---

## Key Idiomatic Patterns

| Pattern | Where |
|---------|-------|
| Interface-based repository | `note/repository.go` + `store/sqlite.go` |
| Sentinel errors + `errors.Is` | `note/errors.go` + `handler/notes.go` |
| Error wrapping with `fmt.Errorf("%w", err)` | Throughout service/store layers |
| GORM for data access | `store/sqlite.go` with `gorm.DB` |
| Mapping GORM errors → domain errors | `store/sqlite.go` (map `gorm.ErrRecordNotFound` to `note.ErrNotFound`) |
| Middleware as `func(http.Handler) http.Handler` | `middleware/` package |
| Unexported context key type | `middleware/requestid.go` |
| Table-driven tests | `handler/notes_test.go` |
| `httptest` for handler testing | `handler/notes_test.go` |

---

## Verification

```bash
# Run all tests
go test ./...

# Start the server
go run ./cmd/api

# Smoke test with curl
curl -X POST localhost:8080/notes -d '{"title":"hello","body":"world"}'
curl localhost:8080/notes
curl localhost:8080/notes/1
curl -X PUT localhost:8080/notes/1 -d '{"title":"updated","body":"world"}'
curl -X DELETE localhost:8080/notes/1
```
