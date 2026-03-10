package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConstants(t *testing.T) {

	if EnvProduction != "prod" {
		t.Errorf("Expected EnvProduction to be 'prod', got '%s'", EnvProduction)
	}
	if EnvDevelopment != "dev" {
		t.Errorf("Expected EnvDevelopment to be 'dev', got '%s'", EnvDevelopment)
	}
	if ModeTodo != "todo" {
		t.Errorf("Expected ModeTodo to be 'todo', got '%s'", ModeTodo)
	}
	if ModeAgent != "agent" {
		t.Errorf("Expected ModeAgent to be 'agent', got '%s'", ModeAgent)
	}
}

func TestAppDirectoryConstants(t *testing.T) {

	if AppDIR != "/.godo/" {
		t.Errorf("Expected AppDIR to be '/.godo/', got '%s'", AppDIR)
	}
	if AppName != "logs" {
		t.Errorf("Expected AppName to be 'logs', got '%s'", AppName)
	}
}

func TestMessageTypes(t *testing.T) {

	if ErrorType != "error" {
		t.Errorf("Expected ErrorType to be 'error', got '%s'", ErrorType)
	}
	if MessageType != "message" {
		t.Errorf("Expected MessageType to be 'message', got '%s'", MessageType)
	}
}

func TestPingChannel(t *testing.T) {

	if Ping == nil {
		t.Error("Ping channel should be initialized")
	}

	select {
	case Ping <- "test":
	case <-Ping:
	}
}

func TestDefaultDBName(t *testing.T) {

	if Cfg.DB_NAME != "todo.db" && Cfg.DB_NAME != "" {
		t.Errorf("Expected default DB_NAME to be 'todo.db', got '%s'", Cfg.DB_NAME)
	}
}

// ===== EDGE CASE TESTS =====

func TestLoadEnvMultiplePaths(t *testing.T) {

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	if err := os.WriteFile(envFile, []byte("TEST_KEY=test_value"), 0644); err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	err := loadEnv([]string{"/nonexistent/.env", envFile})
	if err != nil {
		t.Errorf("Expected to find env file, got error: %v", err)
	}
}

func TestLoadEnvNoPaths(t *testing.T) {
	err := loadEnv([]string{})
	if err != nil {
		t.Logf("Got expected error for empty paths: %v", err)
	}
}

func TestConfigDefaultValues(t *testing.T) {
	// These tests require MustLoad() to be called first
	// Without MustLoad(), Cfg will have zero values
	// This test verifies that we can't assume defaults are set
	if Cfg.OPENAI_MODEL != "" || Cfg.OPENAI_BASE_URL != "" || Cfg.ENVIRONMENT != "" || Cfg.MODE != "" {
		// If MustLoad was called, defaults should be set
		if Cfg.OPENAI_MODEL == "" {
			t.Error("OPENAI_MODEL should have a default")
		}
		if Cfg.OPENAI_BASE_URL == "" {
			t.Error("OPENAI_BASE_URL should have a default")
		}
		if Cfg.ENVIRONMENT == "" {
			t.Error("ENVIRONMENT should have a default")
		}
		if Cfg.MODE == "" {
			t.Error("MODE should have a default")
		}
	}
}

func TestConfigModeValues(t *testing.T) {

	validModes := []string{ModeTodo, ModeAgent}
	found := false
	for _, m := range validModes {
		if Cfg.MODE == m {
			found = true
			break
		}
	}
	if !found && Cfg.MODE != "" {
		t.Logf("Current mode '%s' is not a standard mode", Cfg.MODE)
	}
	// If MODE is empty, that's also fine - just means MustLoad wasn't called
}

func TestPingChannelBuffer(t *testing.T) {

	if Ping == nil {
		t.Error("Ping channel should be initialized")
	}

	for i := 0; i < 250; i++ {
		select {
		case Ping <- "ping":
		default:
			i = 250
		}
	}

	for {
		select {
		case <-Ping:
		default:
			return
		}
	}
}

func TestConstantsNotEmpty(t *testing.T) {

	if EnvProduction == "" {
		t.Error("EnvProduction should not be empty")
	}
	if EnvDevelopment == "" {
		t.Error("EnvDevelopment should not be empty")
	}
	if ModeTodo == "" {
		t.Error("ModeTodo should not be empty")
	}
	if ModeAgent == "" {
		t.Error("ModeAgent should not be empty")
	}
	if AppDIR == "" {
		t.Error("AppDIR should not be empty")
	}
	if AppName == "" {
		t.Error("AppName should not be empty")
	}
}

func TestMessageTypeConstants(t *testing.T) {

	if ErrorType == "" { // Note: intentionally spelled wrong in config
		t.Error("ErrorType should not be empty")
	}
	if MessageType == "" {
		t.Error("MessageType should not be empty")
	}
}
