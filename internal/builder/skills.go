package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileSkillsLoader struct {
	skillsDir string
}

func NewFileSkillsLoader(baseDir string) *FileSkillsLoader {
	return &FileSkillsLoader{
		skillsDir: filepath.Join(baseDir, "skills"),
	}
}

func (l *FileSkillsLoader) BuildSkillsSummary() string {
	entries, err := os.ReadDir(l.skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "" // No skills directory
		}
		return fmt.Sprintf("Error reading skills directory: %v", err)
	}

	var skills []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			skillName := strings.TrimSuffix(entry.Name(), ".md")
			skills = append(skills, fmt.Sprintf("- [%s]: Available in skills/%s", skillName, entry.Name()))
		}
	}

	if len(skills) == 0 {
		return ""
	}

	return strings.Join(skills, "\n")
}
