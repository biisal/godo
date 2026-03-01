package builder

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/biisal/godo/internal/config"
	"github.com/biisal/godo/internal/memory"
)

// SkillsLoader interface placeholder for future implementation
type SkillsLoader interface {
	BuildSkillsSummary() string
}

// Memory interface placeholder for future implementation
type Memory interface {
	GetMemoryContext() string
}

// DummySkillsLoader for when no skills loader is provided
type DummySkillsLoader struct{}

func (d *DummySkillsLoader) BuildSkillsSummary() string { return "" }

// DummyMemory for when no memory is provided
type DummyMemory struct{}

func (d *DummyMemory) GetMemoryContext() string { return "" }

type ContextBuilder struct {
	skillsLoader SkillsLoader
	memory       Memory
	baseDir      string
}

func NewContextBuilder() *ContextBuilder {
	baseDir, _ := os.Getwd()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("Could not determine user home directory, falling back to current directory", "err", err)
		homeDir = baseDir
	}
	identityDir := filepath.Join(homeDir, config.AppDIR, "content", "identity")

	return &ContextBuilder{
		skillsLoader: NewFileSkillsLoader(filepath.Join(homeDir, config.AppDIR, "content")),
		memory:       memory.NewMemoryStore(config.Cfg.DB),
		baseDir:      identityDir,
	}
}

func (cb *ContextBuilder) getIdentity() string {
	return `You are the AI assistant inside the GoDo CLI app.
Use the tools available to fulfill the user's requests.
Always respond in plain text. Do NOT use markdown formatting (no headers, bold, italic, bullet points, or code blocks) as the output is displayed in a terminal.
When you learn important facts about the user or their preferences, save them with the SaveMemory tool so you remember across sessions.
Use RecallMemories when you need to look up previously saved information.`
}

func (cb *ContextBuilder) LoadBootstrapFiles() string {
	files := []string{"SOUL.md", "AGENT.md", "IDENTITY.md", "USER.md"}
	var combined []string

	for _, file := range files {
		path := filepath.Join(cb.baseDir, file)
		content, err := os.ReadFile(path)
		if err == nil {
			combined = append(combined, string(content))
		}
	}

	return strings.Join(combined, "\n\n")
}

func (cb *ContextBuilder) BuildSystemPrompt() string {
	parts := []string{}

	parts = append(parts, cb.getIdentity())

	bootstrapContent := cb.LoadBootstrapFiles()
	if bootstrapContent != "" {
		parts = append(parts, bootstrapContent)
	}

	skillsSummary := cb.skillsLoader.BuildSkillsSummary()
	if skillsSummary != "" {
		parts = append(parts, fmt.Sprintf(`# Skills
%s`, skillsSummary))
	}

	// Memory context
	memoryContext := cb.memory.GetMemoryContext()
	if memoryContext != "" {
		parts = append(parts, "# Memory\n\n"+memoryContext)
	}

	// Join with "---" separator
	return strings.Join(parts, "\n\n---\n\n")
}
