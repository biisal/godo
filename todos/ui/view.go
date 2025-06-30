package ui

import (
	"strings"

	"github.com/biisal/todo-cli/todos/models/agent"
	"github.com/biisal/todo-cli/todos/models/todo"
	"github.com/biisal/todo-cli/todos/ui/setup"
	"github.com/biisal/todo-cli/todos/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

func TodoView(m *TeaModel) string {
	var s string
	switch m.TodoModel.Choices[m.TodoModel.SelectedIndex].Value {
	case TodoAddMode.Value:
		titleInput := m.TodoModel.AddModel.TitleInput
		descInput := m.TodoModel.AddModel.DescInput
		switch m.TodoModel.AddModel.Focus {
		case 0:
			s += styles.BoxStyle.Render(styles.FocusedStyle.Render(titleInput.View())) + "\n"
			s += descInput.View() + "\n"
		case 1:
			s += styles.BlurredStyle.Render(titleInput.View()) + "\n"
			s += descInput.View() + "\n"
		}

	case TodoListMode.Value:
		listModel := m.TodoModel.ListModel
		listModel.List.SetWidth(m.Width/2 - 5)

		var rightContent = "Description:\n"
		if selectedItem := listModel.List.SelectedItem(); selectedItem != nil {
			if i, ok := selectedItem.(todo.Todo); ok {
				rightContent += i.Description()
			}
		}
		separator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")).
			Width(m.Width).
			Render(strings.Repeat("─ ", m.Width/2)) + "\n"
		leftStyle := lipgloss.NewStyle().
			Width(m.Width/2 - 5).
			Height(listModel.List.Height()).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89b4fa"))
		listHeight := listModel.List.Height()

		rightStyle := lipgloss.NewStyle().
			Width(m.Width/2 - 5).
			Height(listHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89b4fa"))

		rightSide := rightStyle.Render(m.TodoModel.ListModel.DescViewport.View())
		view := lipgloss.JoinHorizontal(lipgloss.Center, leftStyle.Render(listModel.List.View()), rightSide)
		s += view + "\n" + separator

	case TodoEditMode.Value:
		titleInput := m.TodoModel.EditModel.TitleInput
		descInput := m.TodoModel.EditModel.DescInput
		idInput := m.TodoModel.EditModel.IdInput
		switch m.TodoModel.EditModel.Focus {
		case 0:
			s += styles.BoxStyle.Render(styles.FocusedStyle.Render(titleInput.View())) + "\n"
			s += descInput.View() + "\n"
			s += styles.BlurredStyle.Render(idInput.View())
		case 1:
			s += styles.BlurredStyle.Render(titleInput.View()) + "\n"
			s += descInput.View() + "\n"
			s += styles.BlurredStyle.Render(idInput.View())
		default:
			s += styles.BlurredStyle.Render(titleInput.View()) + "\n"
			s += descInput.View() + "\n"
			s += styles.BoxStyle.Render(styles.FocusedStyle.Render(idInput.View())) + "\n"
		}

	}
	s += "\n\n"
	s += setup.SetUpChoice(m.TodoModel.Choices, m.TodoModel.SelectedIndex, "ctrl+right/left")

	return s
}

func AgentView(m *TeaModel) string {
	var s string
	for _, msg := range m.AgentModel.History {
		content := strings.TrimSpace(msg.Content)
		if content == "" || strings.HasPrefix(content, "<function") {
			continue
		}
		if msg.Role == agent.UserRole {
			s += styles.GreenStyle.Width(m.Width-20).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#00FF00")).Render(content) + "\n"
		} else {
			s += styles.DescStyle.Width(m.Width-20).Render(content) + "\n"
		}
	}
	s += m.AgentModel.Response + "\n"
	s += styles.BoxStyle.Render(m.AgentModel.PromptInput.View())

	return s
}
