package memory

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory SQLite database with the memories table.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS memories (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(sqlStmt); err != nil {
		t.Fatalf("Failed to create memories table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewMemoryStore(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)
	if store == nil {
		t.Fatal("NewMemoryStore should not return nil")
	}
}

// ===== Save Tests =====

func TestSaveBasic(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	err := store.Save("favorite_language", "Go")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	entries, err := store.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Key != "favorite_language" || entries[0].Content != "Go" {
		t.Errorf("Unexpected entry: key=%q content=%q", entries[0].Key, entries[0].Content)
	}
}

func TestSaveUpsert(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	if err := store.Save("lang", "Python"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if err := store.Save("lang", "Go"); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	entries, err := store.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry after upsert, got %d", len(entries))
	}
	if entries[0].Content != "Go" {
		t.Errorf("Expected updated content 'Go', got %q", entries[0].Content)
	}
}

func TestSaveMultiple(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	keys := []string{"name", "language", "editor", "os"}
	for _, k := range keys {
		if err := store.Save(k, "value_"+k); err != nil {
			t.Fatalf("Save(%q) failed: %v", k, err)
		}
	}

	entries, err := store.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(entries) != len(keys) {
		t.Fatalf("Expected %d entries, got %d", len(keys), len(entries))
	}
}

func TestSaveEmptyKey(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	err := store.Save("", "some content")
	if err == nil {
		t.Error("Save with empty key should return an error")
	}
}

func TestSaveEmptyContent(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	err := store.Save("key", "")
	if err == nil {
		t.Error("Save with empty content should return an error")
	}
}

func TestSaveWhitespaceOnly(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	err := store.Save("   ", "content")
	if err == nil {
		t.Error("Save with whitespace-only key should return an error")
	}

	err = store.Save("key", "   ")
	if err == nil {
		t.Error("Save with whitespace-only content should return an error")
	}
}

func TestSaveTrimsWhitespace(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	if err := store.Save("  lang  ", "  Go  "); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	entries, err := store.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if entries[0].Key != "lang" {
		t.Errorf("Expected trimmed key 'lang', got %q", entries[0].Key)
	}
	if entries[0].Content != "Go" {
		t.Errorf("Expected trimmed content 'Go', got %q", entries[0].Content)
	}
}

// ===== Delete Tests =====

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	store.Save("lang", "Go")
	store.Save("editor", "vim")

	if err := store.Delete("lang"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	entries, err := store.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry after delete, got %d", len(entries))
	}
	if entries[0].Key != "editor" {
		t.Errorf("Expected remaining key 'editor', got %q", entries[0].Key)
	}
}

func TestDeleteNonExistent(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	// Should not error when deleting a non-existent key
	err := store.Delete("does_not_exist")
	if err != nil {
		t.Fatalf("Delete of non-existent key should not error, got: %v", err)
	}
}

// ===== Search Tests =====

func TestSearchByKey(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	store.Save("favorite_language", "Go")
	store.Save("favorite_editor", "vim")
	store.Save("os", "macOS")

	results, err := store.Search("favorite")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results searching 'favorite', got %d", len(results))
	}
}

func TestSearchByContent(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	store.Save("lang", "Go programming language")
	store.Save("hobby", "programming robots")
	store.Save("pet", "golden retriever")

	results, err := store.Search("programming")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results searching 'programming', got %d", len(results))
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	store.Save("a", "1")
	store.Save("b", "2")
	store.Save("c", "3")

	// Empty query should return all
	results, err := store.Search("")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("Expected 3 results for empty query, got %d", len(results))
	}
}

func TestSearchNoResults(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	store.Save("lang", "Go")

	results, err := store.Search("python")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("Expected 0 results, got %d", len(results))
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	store.Save("Language", "GoLang")

	results, err := store.Search("golang")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result for case-insensitive search, got %d", len(results))
	}
}

// ===== GetAll Tests =====

func TestGetAllEmpty(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	entries, err := store.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("Expected 0 entries, got %d", len(entries))
	}
}

func TestGetAllOrder(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	store.Save("first", "1")
	store.Save("second", "2")
	store.Save("third", "3")

	// Update the first one so it becomes most recent
	store.Save("first", "updated")

	entries, err := store.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}
	// Most recently updated should be first
	if entries[0].Key != "first" {
		t.Errorf("Expected most recently updated key 'first' at index 0, got %q", entries[0].Key)
	}
}

// ===== GetMemoryContext Tests =====

func TestGetMemoryContextEmpty(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	result := store.GetMemoryContext()
	if result != "" {
		t.Errorf("Expected empty string for empty store, got %q", result)
	}
}

func TestGetMemoryContextNilDB(t *testing.T) {
	store := NewMemoryStore(nil)
	result := store.GetMemoryContext()
	if result != "" {
		t.Errorf("Expected empty string for nil DB, got %q", result)
	}
}

func TestGetMemoryContextFormatted(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	store.Save("name", "Avisek")
	store.Save("lang", "Go")

	result := store.GetMemoryContext()

	if result == "" {
		t.Fatal("Expected non-empty memory context")
	}
	if !containsSubstr(result, "**name**") || !containsSubstr(result, "Avisek") {
		t.Error("Memory context should contain formatted name entry")
	}
	if !containsSubstr(result, "**lang**") || !containsSubstr(result, "Go") {
		t.Error("Memory context should contain formatted lang entry")
	}
}

func TestGetMemoryContextBulletPoints(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	store.Save("a", "1")
	store.Save("b", "2")

	result := store.GetMemoryContext()

	if !containsSubstr(result, "- **a**: 1") {
		t.Errorf("Expected bullet point format, got %q", result)
	}
	if !containsSubstr(result, "- **b**: 2") {
		t.Errorf("Expected bullet point format, got %q", result)
	}
}

// ===== Special Characters Tests =====

func TestSaveSpecialCharacters(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	if err := store.Save("greeting", "Hello 世界 🌍"); err != nil {
		t.Fatalf("Save with special chars failed: %v", err)
	}

	entries, err := store.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if entries[0].Content != "Hello 世界 🌍" {
		t.Errorf("Special characters not preserved: got %q", entries[0].Content)
	}
}

func TestSaveMultilineContent(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(db)

	multiline := "line1\nline2\nline3"
	if err := store.Save("multi", multiline); err != nil {
		t.Fatalf("Save multiline failed: %v", err)
	}

	entries, err := store.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if entries[0].Content != multiline {
		t.Errorf("Multiline not preserved: got %q", entries[0].Content)
	}
}

// Helper
func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
