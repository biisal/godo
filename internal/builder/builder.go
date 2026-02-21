package builder

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/biisal/godo/internal/config"
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
	// Resolve base directory dynamically based on current working directory
	baseDir, _ := os.Getwd()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("Could not determine user home directory, falling back to current directory", "err", err)
		homeDir = baseDir
	}
	var identityDir string

	identityDir = filepath.Join(homeDir, config.AppDIR, "content", "identity")

	return &ContextBuilder{
		skillsLoader: NewFileSkillsLoader(filepath.Join(homeDir, config.AppDIR, "content")), // resolves to homeDir/.godo/content/skills
		memory:       &DummyMemory{},
		baseDir:      identityDir,
	}
}

func (cb *ContextBuilder) getIdentity() string {
	return `You are the AI assistant inside the GoDo CLI app.
Use the tools available to fulfill the user's requests.
Always respond in plain text. Do NOT use markdown formatting (no headers, bold, italic, bullet points, or code blocks) as the output is displayed in a terminal.`
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

	// Core identity section
	parts = append(parts, cb.getIdentity())

	// Bootstrap files (SOUL.md, AGENT.md, IDENTITY.md, USER.md)
	bootstrapContent := cb.LoadBootstrapFiles()
	if bootstrapContent != "" {
		parts = append(parts, bootstrapContent)
	}

	// Skills summary
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
