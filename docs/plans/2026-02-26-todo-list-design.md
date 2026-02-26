# Todo Lists Feature Design

**Date:** 2026-02-26
**Feature:** Add todo/checklist support to notes (Google Keep style)

## Overview

Extend the Note model to support two types of notes:
- **Text notes** — Traditional notes with title and body text
- **List notes** — Checklists with a title and array of todos (text + completion status)

Users work with the entire todo list at once (no granular todo endpoints).

## Data Model

### NoteType Enum

```go
type NoteType string

const (
    NoteTypeText NoteType = "text"
    NoteTypeList NoteType = "list"
)
```

Type-safe constants to avoid string confusion.

### Todo Struct

```go
type Todo struct {
    Text      string `json:"text"`
    Completed bool   `json:"completed"`
}
```

Minimal structure: todo description and completion status only.

### Updated Note Struct

```go
type Note struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"index"`
    Type      NoteType  `gorm:"type:text"`     // "text" or "list"
    Title     string
    Body      *string   `gorm:"type:text"`     // nullable for list-type notes
    Todos     []Todo    `gorm:"type:json"`     // JSON array of todos
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Key decisions:**
- `Type` field determines note behavior
- `Body` is nullable (list notes don't use body)
- `Todos` stored as JSON in database — GORM handles serialization

## Database Schema

Add two columns to `notes` table:

| Column | Type | Default | Notes |
|--------|------|---------|-------|
| `type` | TEXT | `'text'` | Backward compatible default |
| `todos` | JSON | `NULL` | Empty or null for text notes |

Existing notes automatically get `type='text'` and `todos=NULL`.

## API Behavior

### Create Text Note

```http
POST /notes
Content-Type: application/json

{
  "type": "text",
  "title": "My Thoughts",
  "body": "Some text here..."
}
```

### Create List Note

```http
POST /notes
Content-Type: application/json

{
  "type": "list",
  "title": "Groceries",
  "todos": [
    {"text": "Milk", "completed": false},
    {"text": "Bread", "completed": true},
    {"text": "Eggs", "completed": false}
  ]
}
```

### Response (Both Types)

```json
{
  "id": 1,
  "type": "list",
  "title": "Groceries",
  "body": null,
  "todos": [
    {"text": "Milk", "completed": false},
    {"text": "Bread", "completed": true},
    {"text": "Eggs", "completed": false}
  ],
  "createdAt": "2026-02-26T10:30:00Z",
  "updatedAt": "2026-02-26T10:30:00Z"
}
```

### Update Note

Replace entire todos array (replace-all semantics):

```http
PUT /notes/{id}
Content-Type: application/json

{
  "type": "list",
  "title": "Groceries Updated",
  "todos": [
    {"text": "Milk", "completed": true},
    {"text": "Cheese", "completed": false}
  ]
}
```

No separate todo endpoints — all changes go through the main note endpoints.

## Validation Rules

In `internal/note/service.go`:

1. **Type must be valid:** `"text"` or `"list"` only
2. **Title required:** Non-empty for both types
3. **Type consistency:**
   - Text notes: `todos` must be empty or omitted
   - List notes: `body` may be null/empty, `todos` required (can be empty array)
4. **Todo text:** Can be any string (including empty, for simplicity)

## Handler Changes

Update `internal/handler/notes.go`:

- **Parse `type` field:** Validate against `NoteTypeText` and `NoteTypeList`
- **Handle todos array:** Unmarshal and validate structure
- **Response:** Return full note including todos (or empty array/null as appropriate)
- **No new endpoints:** Existing CRUD endpoints (GET, POST, PUT, DELETE) handle both types

## Database Storage

GORM's `type:json` tag automatically:
- Serializes `[]Todo` to JSON when saving
- Deserializes JSON to `[]Todo` when fetching

No custom marshaling needed.

## Testing

Add test cases to `handler/notes_test.go` and `store/sqlite_test.go`:

1. Create text note → verify `type="text"`, todos empty
2. Create list note → verify `type="list"`, todos populated
3. Update list note → verify todos are completely replaced
4. Fetch note → verify todos are deserialized correctly
5. Type validation → reject invalid type values
6. Edge case: empty todos array in list note

## Migration Path

Existing notes:
- Remain as `type='text'` (default)
- `todos` column initialized as `NULL`
- No data loss; all existing endpoints continue to work

## Future Extensions

Possible enhancements (out of scope):
- Sub-tasks/nested todos
- Due dates per todo
- Todo filtering/search
- Todo-level permissions
- Archiving completed todos

For now: keep it simple. Just text + completed status.
