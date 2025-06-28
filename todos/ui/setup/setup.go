package setup

import (
	"github.com/biisal/todo-cli/todos/models"
	"github.com/charmbracelet/lipgloss"
)

func SetUpChoice(choices []models.Mode, matchIndex int) string {
	var s string
	for i, choice := range choices {
		if i == matchIndex {
			selectedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FFFF")).
				Bold(true).
				Underline(true)
			s += selectedStyle.Render("[ "+choice.Label+" ]") + "  "
		} else {

			normalStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666"))
			s += normalStyle.Render("[ "+choice.Label+" ]") + "  "
		}
	}
	return s + "\n\n"
}
