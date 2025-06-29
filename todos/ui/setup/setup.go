package setup

import (
	"github.com/biisal/todo-cli/todos/models/todo"
	"github.com/biisal/todo-cli/todos/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

func SetUpChoice(choices []todo.Mode, matchIndex int, helpText string) string {
	var s string
	for i, choice := range choices {
		if i == matchIndex {
			selectedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FFFF")).
				Bold(true)
			s += selectedStyle.Render("["+choice.Label+"]") + "  "
		} else {
			normalStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666"))
			s += normalStyle.Render("["+choice.Label+"]") + "  "
		}
	}

	termWidth, _ := lipgloss.Size(s)
	help := styles.HelpStyle.Width(termWidth).Render(helpText)
	choiceBox := styles.BoxStyle.Render(s)

	return help + "\n" + choiceBox + "\n"
}
