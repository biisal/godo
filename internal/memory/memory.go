package memory

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// MemoryEntry represents a single memory stored in the database.
type MemoryEntry struct {
	ID        int
	Key       string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MemoryStore provides persistent key-value memory backed by SQLite.
type MemoryStore struct {
	db *sql.DB
}

// NewMemoryStore creates a new MemoryStore using the given database connection.
func NewMemoryStore(db *sql.DB) *MemoryStore {
	return &MemoryStore{db: db}
}

// Save upserts a memory entry by key.
func (m *MemoryStore) Save(key, content string) error {
	key = strings.TrimSpace(key)
	content = strings.TrimSpace(content)
	if key == "" {
		return fmt.Errorf("memory key cannot be empty")
	}
	if content == "" {
		return fmt.Errorf("memory content cannot be empty")
	}

	sqlStmt := `
		INSERT INTO memories (key, content, created_at, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			content = excluded.content,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := m.db.Exec(sqlStmt, key, content)
	return err
}

// Delete removes a memory entry by key.
func (m *MemoryStore) Delete(key string) error {
	_, err := m.db.Exec("DELETE FROM memories WHERE key = ?", key)
	return err
}

// GetAll returns all memory entries ordered by most recently updated.
func (m *MemoryStore) GetAll() ([]MemoryEntry, error) {
	rows, err := m.db.Query("SELECT id, key, content, created_at, updated_at FROM memories ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("Error closing rows on getall memory", "error", err)
		}
	}()

	var entries []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		if err := rows.Scan(&e.ID, &e.Key, &e.Content, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// Search finds memory entries where the key or content matches the query (case-insensitive).
func (m *MemoryStore) Search(query string) ([]MemoryEntry, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return m.GetAll()
	}

	likePattern := "%" + query + "%"
	rows, err := m.db.Query(
		"SELECT id, key, content, created_at, updated_at FROM memories WHERE key LIKE ? OR content LIKE ? ORDER BY updated_at DESC",
		likePattern, likePattern,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("Error closing rows on searching memory", "error", err)
		}
	}()

	var entries []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		if err := rows.Scan(&e.ID, &e.Key, &e.Content, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// GetMemoryContext implements the builder.Memory interface.
// Returns all memories formatted for injection into the system prompt.
func (m *MemoryStore) GetMemoryContext() string {
	if m.db == nil {
		return ""
	}
	entries, err := m.GetAll()
	if err != nil || len(entries) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&sb, "- **%s**: %s\n", e.Key, e.Content)
	}
	return sb.String()
}
