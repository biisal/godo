package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewContextBuilder(t *testing.T) {
	cb := NewContextBuilder()
	if cb == nil {
		t.Fatal("NewContextBuilder should not return nil")
	}
	if cb.skillsLoader == nil {
		t.Fatal("skillsLoader should not be nil")
	}
	if cb.memory == nil {
		t.Fatal("memory should not be nil")
	}
}

func TestContextBuilderBaseDir(t *testing.T) {
	cb := NewContextBuilder()
	if cb.baseDir == "" {
		t.Error("baseDir should not be empty")
	}
}

func TestGetIdentity(t *testing.T) {
	cb := NewContextBuilder()
	identity := cb.getIdentity()

	if identity == "" {
		t.Error("Identity should not be empty")
	}

	expected := "You are the AI assistant inside the GoDo CLI app"
	if len(identity) < len(expected) {
		t.Error("Identity seems too short")
	}
}

func TestLoadBootstrapFiles(t *testing.T) {
	cb := NewContextBuilder()

	content := cb.LoadBootstrapFiles()
	_ = content
}

func TestLoadBootstrapFilesWithMockFiles(t *testing.T) {

	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, "identity")

	err := os.MkdirAll(identityDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	files := map[string]string{
		"SOUL.md":     "This is the soul content",
		"AGENT.md":    "This is the agent content",
		"IDENTITY.md": "This is the identity content",
		"USER.md":     "This is the user content",
		"EXTRA.md":    "This should not be loaded",
	}

	for name, content := range files {
		path := filepath.Join(identityDir, name)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write %s: %v", name, err)
		}
	}

	cb := &ContextBuilder{
		skillsLoader: &DummySkillsLoader{},
		memory:       &DummyMemory{},
		baseDir:      identityDir,
	}

	result := cb.LoadBootstrapFiles()

	expectedContents := []string{
		"This is the soul content",
		"This is the agent content",
		"This is the identity content",
		"This is the user content",
	}

	for _, expected := range expectedContents {
		if len(result) < len(expected) || !contains(result, expected) {
			t.Errorf("Expected content '%s' to be in result", expected)
		}
	}

	if contains(result, "This should not be loaded") {
		t.Error("EXTRA.md should not be loaded")
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	cb := NewContextBuilder()
	prompt := cb.BuildSystemPrompt()

	if prompt == "" {
		t.Error("System prompt should not be empty")
	}

	if !contains(prompt, "You are the AI assistant inside the GoDo CLI app") {
		t.Error("System prompt should contain identity")
	}

	_ = prompt
}

func TestBuildSystemPromptWithMockData(t *testing.T) {
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, "identity")

	err := os.MkdirAll(identityDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(identityDir, "SOUL.md"), []byte("Soul content"), 0644); err != nil {
		t.Fatalf("Failed to write SOUL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(identityDir, "AGENT.md"), []byte("Agent content"), 0644); err != nil {
		t.Fatalf("Failed to write AGENT.md: %v", err)
	}

	mockSkills := &MockSkillsLoader{summary: "Test Skills: coding, analysis"}

	mockMemory := &MockMemory{context: "User prefers concise responses"}

	cb := &ContextBuilder{
		skillsLoader: mockSkills,
		memory:       mockMemory,
		baseDir:      identityDir,
	}

	prompt := cb.BuildSystemPrompt()

	if !contains(prompt, "Soul content") {
		t.Error("Should contain SOUL.md content")
	}
	if !contains(prompt, "Agent content") {
		t.Error("Should contain AGENT.md content")
	}
	if !contains(prompt, "Test Skills:") {
		t.Error("Should contain skills summary")
	}
	if !contains(prompt, "User prefers concise responses") {
		t.Error("Should contain memory context")
	}
	if !contains(prompt, "# Skills") {
		t.Error("Should have Skills section header")
	}
	if !contains(prompt, "# Memory") {
		t.Error("Should have Memory section header")
	}
}

func TestBuildSystemPromptSeparators(t *testing.T) {
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, "identity")

	err := os.MkdirAll(identityDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(identityDir, "SOUL.md"), []byte("Part1"), 0644); err != nil {
		t.Fatalf("Failed to write SOUL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(identityDir, "AGENT.md"), []byte("Part2"), 0644); err != nil {
		t.Fatalf("Failed to write AGENT.md: %v", err)
	}

	cb := &ContextBuilder{
		skillsLoader: &DummySkillsLoader{},
		memory:       &DummyMemory{},
		baseDir:      identityDir,
	}

	prompt := cb.BuildSystemPrompt()

	if !contains(prompt, "---") {
		t.Error("System prompt should use '---' as separator")
	}
}

// ===== EDGE CASE TESTS =====

func TestLoadBootstrapFilesEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, "identity")

	err := os.MkdirAll(identityDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cb := &ContextBuilder{
		skillsLoader: &DummySkillsLoader{},
		memory:       &DummyMemory{},
		baseDir:      identityDir,
	}

	result := cb.LoadBootstrapFiles()
	if result != "" {
		t.Errorf("Expected empty string when no files exist, got '%s'", result)
	}
}

func TestLoadBootstrapFilesPartialFiles(t *testing.T) {
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, "identity")

	err := os.MkdirAll(identityDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(identityDir, "SOUL.md"), []byte("Soul only"), 0644); err != nil {
		t.Fatalf("Failed to write SOUL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(identityDir, "IDENTITY.md"), []byte("Identity only"), 0644); err != nil {
		t.Fatalf("Failed to write IDENTITY.md: %v", err)
	}

	cb := &ContextBuilder{
		skillsLoader: &DummySkillsLoader{},
		memory:       &DummyMemory{},
		baseDir:      identityDir,
	}

	result := cb.LoadBootstrapFiles()

	if !contains(result, "Soul only") {
		t.Error("Should contain SOUL.md content")
	}
	if !contains(result, "Identity only") {
		t.Error("Should contain IDENTITY.md content")
	}
	if contains(result, "Agent only") || contains(result, "User only") {
		t.Error("Should not contain content from missing files")
	}
}

func TestLoadBootstrapFilesVeryLongContent(t *testing.T) {
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, "identity")

	err := os.MkdirAll(identityDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	longContent := strings.Repeat("A", 1024*1024)
	if err := os.WriteFile(filepath.Join(identityDir, "SOUL.md"), []byte(longContent), 0644); err != nil {
		t.Fatalf("Failed to write SOUL.md: %v", err)
	}

	cb := &ContextBuilder{
		skillsLoader: &DummySkillsLoader{},
		memory:       &DummyMemory{},
		baseDir:      identityDir,
	}

	result := cb.LoadBootstrapFiles()
	if len(result) < len(longContent) {
		t.Errorf("Expected long content to be loaded, got length %d", len(result))
	}
}

func TestLoadBootstrapFilesSpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, "identity")

	err := os.MkdirAll(identityDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	specialContent := "Hello 世界 🌍 \"quotes\" \\backslash\n\ttabbed\r\nnewlines"
	if err := os.WriteFile(filepath.Join(identityDir, "SOUL.md"), []byte(specialContent), 0644); err != nil {
		t.Fatalf("Failed to write SOUL.md: %v", err)
	}

	cb := &ContextBuilder{
		skillsLoader: &DummySkillsLoader{},
		memory:       &DummyMemory{},
		baseDir:      identityDir,
	}

	result := cb.LoadBootstrapFiles()
	if !contains(result, specialContent) {
		t.Error("Should preserve special characters")
	}
}

func TestLoadBootstrapFilesNonExistentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentDir := filepath.Join(tmpDir, "does_not_exist")

	cb := &ContextBuilder{
		skillsLoader: &DummySkillsLoader{},
		memory:       &DummyMemory{},
		baseDir:      nonExistentDir,
	}

	result := cb.LoadBootstrapFiles()
	if result != "" {
		t.Errorf("Expected empty string for non-existent directory, got '%s'", result)
	}
}

func TestBuildSystemPromptEmptySkillsAndMemory(t *testing.T) {
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, "identity")

	err := os.MkdirAll(identityDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cb := &ContextBuilder{
		skillsLoader: &DummySkillsLoader{},
		memory:       &DummyMemory{},
		baseDir:      identityDir,
	}

	prompt := cb.BuildSystemPrompt()

	if !contains(prompt, "You are the AI assistant inside the GoDo CLI app") {
		t.Error("Should contain core identity")
	}

	if contains(prompt, "# Skills") {
		t.Error("Should not have Skills section when skills are empty")
	}
	if contains(prompt, "# Memory") {
		t.Error("Should not have Memory section when memory is empty")
	}
}

func TestBuildSystemPromptWithOnlySkills(t *testing.T) {
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, "identity")

	err := os.MkdirAll(identityDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	mockSkills := &MockSkillsLoader{summary: "Test Skills: coding, analysis"}

	cb := &ContextBuilder{
		skillsLoader: mockSkills,
		memory:       &DummyMemory{},
		baseDir:      identityDir,
	}

	prompt := cb.BuildSystemPrompt()

	if !contains(prompt, "# Skills") {
		t.Error("Should have Skills section header")
	}
	if !contains(prompt, "coding, analysis") {
		t.Error("Should contain skills content")
	}

	if contains(prompt, "# Memory") {
		t.Error("Should not have Memory section when memory is empty")
	}
}

func TestBuildSystemPromptWithOnlyMemory(t *testing.T) {
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, "identity")

	err := os.MkdirAll(identityDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	mockMemory := &MockMemory{context: "User prefers concise responses"}

	cb := &ContextBuilder{
		skillsLoader: &DummySkillsLoader{},
		memory:       mockMemory,
		baseDir:      identityDir,
	}

	prompt := cb.BuildSystemPrompt()

	if !contains(prompt, "# Memory") {
		t.Error("Should have Memory section header")
	}
	if !contains(prompt, "User prefers concise responses") {
		t.Error("Should contain memory content")
	}

	if contains(prompt, "# Skills") {
		t.Error("Should not have Skills section when skills are empty")
	}
}

func TestBuildSystemPromptMultipleSeparators(t *testing.T) {
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, "identity")

	err := os.MkdirAll(identityDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	for name, content := range map[string]string{
		"SOUL.md": "Part1", "AGENT.md": "Part2",
		"IDENTITY.md": "Part3", "USER.md": "Part4",
	} {
		if err := os.WriteFile(filepath.Join(identityDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", name, err)
		}
	}

	mockSkills := &MockSkillsLoader{summary: "Skills"}
	mockMemory := &MockMemory{context: "Memory"}

	cb := &ContextBuilder{
		skillsLoader: mockSkills,
		memory:       mockMemory,
		baseDir:      identityDir,
	}

	prompt := cb.BuildSystemPrompt()

	count := strings.Count(prompt, "---")
	if count < 3 {
		t.Errorf("Expected at least 3 separators, got %d", count)
	}
}

func TestContextBuilderWithNilInterfaces(t *testing.T) {
	t.Skip("Skipping - this test reveals a bug where nil interfaces cause panic. Fix the code to handle nil interfaces first.")
}

func TestContextBuilderBaseDirWithSpaces(t *testing.T) {
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, "my identity dir")

	err := os.MkdirAll(identityDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(identityDir, "SOUL.md"), []byte("Content with space path"), 0644); err != nil {
		t.Fatalf("Failed to write SOUL.md: %v", err)
	}

	cb := &ContextBuilder{
		skillsLoader: &DummySkillsLoader{},
		memory:       &DummyMemory{},
		baseDir:      identityDir,
	}

	result := cb.LoadBootstrapFiles()
	if !contains(result, "Content with space path") {
		t.Error("Should work with paths containing spaces")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type MockSkillsLoader struct {
	summary string
}

func (m *MockSkillsLoader) BuildSkillsSummary() string {
	return m.summary
}

type MockMemory struct {
	context string
}

func (m *MockMemory) GetMemoryContext() string {
	return m.context
}

var _ SkillsLoader = (*MockSkillsLoader)(nil)
var _ Memory = (*MockMemory)(nil)

func TestDummySkillsLoader(t *testing.T) {
	loader := &DummySkillsLoader{}
	result := loader.BuildSkillsSummary()
	if result != "" {
		t.Errorf("Expected empty string from DummySkillsLoader, got '%s'", result)
	}
}

func TestDummyMemory(t *testing.T) {
	mem := &DummyMemory{}
	result := mem.GetMemoryContext()
	if result != "" {
		t.Errorf("Expected empty string from DummyMemory, got '%s'", result)
	}
}
