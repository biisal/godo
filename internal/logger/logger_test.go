package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	closeLog, err := Init(logFile, 0)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer closeLog()

	// Test that the log file was created
	info, err := os.Stat(logFile)
	if err != nil {
		t.Errorf("Log file was not created: %v", err)
	}
	_ = info
}

func TestInitWithDebugLevel(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "debug.log")

	closeLog, err := Init(logFile, 0)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Write a debug log
	slog.Info("test message")

	// Close to flush
	closeLog()

	// Read the log file
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logStr := string(content)
	if logStr == "" {
		t.Error("Log file should not be empty after initialization")
	}
	if !strings.Contains(logStr, "test message") {
		t.Error("Log file should contain the test message")
	}
}

func TestInitCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "newfile.log")

	_, err := Init(logFile, 0)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// File should exist now
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file should have been created")
	}
}

func TestInitAppendsToExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "append.log")

	// Create file with existing content
	existingContent := `{"time":"2024-01-01T00:00:00Z","level":"INFO","msg":"existing log"}`
	if err := os.WriteFile(logFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write log file: %v", err)
	}

	closeLog, _ := Init(logFile, 0)
	closeLog()

	content, _ := os.ReadFile(logFile)
	if !strings.Contains(string(content), "existing log") {
		t.Error("Existing content should be preserved")
	}
}

func TestInitReturnsCloseFunc(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "close_test.log")

	closeLog, _ := Init(logFile, 0)

	// Call close
	closeLog()

	// Try to open the file - should work since we're just reading
	_, err := os.OpenFile(logFile, os.O_RDONLY, 0644)
	if err != nil {
		t.Errorf("Should be able to open log file after close: %v", err)
	}
}

// ===== EDGE CASE TESTS =====

func TestInitWithInvalidPath(t *testing.T) {
	// Try to create log file in non-existent directory
	invalidPath := "/nonexistent/path/that/does/not/exist/test.log"

	_, err := Init(invalidPath, 0)
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestInitWithEmptyPath(t *testing.T) {
	// Empty path should fail
	_, err := Init("", 0)
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestInitMultipleTimes(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "multi.log")

	// Initialize multiple times
	for i := 0; i < 3; i++ {
		closeLog, err := Init(logFile, 0)
		if err != nil {
			t.Fatalf("Init %d failed: %v", i, err)
		}
		closeLog()
	}

	// File should still be readable
	_, err := os.ReadFile(logFile)
	if err != nil {
		t.Errorf("Should be able to read log file: %v", err)
	}
}

func TestInitWithDifferentLevels(t *testing.T) {
	tmpDir := t.TempDir()

	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	for _, level := range levels {
		logFile := filepath.Join(tmpDir, fmt.Sprintf("level_%d.log", level))
		closeLog, err := Init(logFile, level)
		if err != nil {
			t.Errorf("Init with level %v failed: %v", level, err)
		}
		closeLog()
	}
}

func TestInitLogFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "perm_test.log")

	closeLog, err := Init(logFile, 0)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Write some logs
	slog.Info("test")

	// Close
	closeLog()

	// Check file permissions
	info, err := os.Stat(logFile)
	if err != nil {
		t.Fatalf("Failed to stat log file: %v", err)
	}

	// File should be readable (at minimum)
	if info.Mode().Perm() == 0 {
		t.Error("Log file should have non-zero permissions")
	}
}

func TestSlogWritesAtDifferentLevels(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "levels.log")

	closeLog, _ := Init(logFile, slog.LevelDebug)

	// Write logs at different levels
	slog.Debug("debug message")
	slog.Info("info message")
	slog.Warn("warn message")
	slog.Error("error message")

	closeLog()

	// Read and verify all messages were logged
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logStr := string(content)
	if !strings.Contains(logStr, "debug message") {
		t.Error("Debug message should be logged")
	}
	if !strings.Contains(logStr, "info message") {
		t.Error("Info message should be logged")
	}
	if !strings.Contains(logStr, "warn message") {
		t.Error("Warn message should be logged")
	}
	if !strings.Contains(logStr, "error message") {
		t.Error("Error message should be logged")
	}
}
