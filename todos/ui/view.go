package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/biisal/todo-cli/config"
	agentaction "github.com/biisal/todo-cli/todos/actions/agent"
	todoaction "github.com/biisal/todo-cli/todos/actions/todo"
	"github.com/biisal/todo-cli/todos/models/agent"
	// "github.com/biisal/todo-cli/todos/ui/setup"
	"github.com/biisal/todo-cli/todos/ui/styles"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func RenderListView(m *TeaModel) string {
	listView := m.TodoModel.ListModel.List.View()

	return lipgloss.Place(
		m.Width,
		m.Height,
		lipgloss.Left,
		lipgloss.Top,
		lipgloss.NewStyle().
			Background(m.Theme.GetBackground()).
			Width(m.Width).
			Render(listView),
	)
}
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
		s = RenderListView(m)
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
	// s += "\n\n"
	// s += setup.SetUpChoice(m.TodoModel.Choices, m.TodoModel.SelectedIndex, "ctrl+right/left")

	return s
}

func AgentView(m *TeaModel) string {
	var s string
	var chatContent string
	historyLen := len(agentaction.History)
	if historyLen == 0 {
		// chatContent = styles.PurpleStyle.Margin(0, 1).Render("Made in India with ❤️ by Biisal") + "\n"
		logo := styles.Logo + styles.ChatInstructions
		chatContent = styles.PurpleStyle.AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center).Height(m.AgentModel.ChatViewport.Height).Width(m.AgentModel.ChatViewport.Width).Render(logo)
	} else {
		for _, msg := range agentaction.History {
			if msg.IsToolReq {
				continue
			}
			content := ""
			for _, part := range msg.Parts {
				content += part.Text
			}
			if content == "" {
				continue
			}
			if msg.Role == agent.UserRole {
				chatContent += m.Theme.GetUserContentStyle().Width(m.Width).Render(content) + "\n"
			} else {
				chatContent += m.Theme.GetAgentContentStyle().Width(m.Width).Render(content)
			}
		}
	}
	m.AgentModel.ChatViewport.SetContent(chatContent)
	chatView := lipgloss.NewStyle().Background(lipgloss.Color("#130F1A")).Render(m.AgentModel.ChatViewport.View())
	topPart := lipgloss.Place(
		m.Width,
		m.AgentModel.ChatViewport.Height,
		lipgloss.Center,
		lipgloss.Top,
		chatView,
	)
	inputView := styles.ChatInputStyle.AlignHorizontal(lipgloss.Left).Width(m.Width).BorderTop(true).
		Border(lipgloss.NormalBorder()).
		BorderBottom(false).
		BorderRight(false).
		BorderLeft(false).
		BorderForeground(lipgloss.Color("#666666")).Render(m.AgentModel.PromptInput.View())
	bottomPart := lipgloss.Place(
		m.Width,
		lipgloss.Height(inputView),
		lipgloss.Left,
		lipgloss.Bottom,
		inputView,
	)
	middleLeft := m.Theme.GetInstructionStyle().
		Render("Quit: Esc / Ctrl+C")

	middleRight := m.Theme.GetInstructionStyle().
		Render("Scroll: PageUp / PageDown")
	middle := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Background(m.Theme.GetBackground()).Width(m.Width/2).Render(middleLeft),
		lipgloss.NewStyle().Background(m.Theme.GetBackground()).Width(m.Width/2).AlignHorizontal(lipgloss.Right).Render(middleRight),
	)
	s += lipgloss.JoinVertical(lipgloss.Top, topPart, middle, bottomPart)
	return s
}

func ExitView() string {
	timeSpent := time.Since(config.StartTime).Round(time.Second).String()
	timeStr := fmt.Sprintf("%v", timeSpent)
	totalTodos, completed, remains, err := todoaction.GetTodosInfo()
	var s string
	if err == nil {
		columns := []table.Column{
			{
				Title: "SUMMURY",
				Width: 20,
			},
			{
				Title: "",
				Width: 20,
			},
		}
		rows := []table.Row{
			{"Time Spent", timeStr},
			{"Total Todos", strconv.Itoa(totalTodos)},
			{"Completed", strconv.Itoa(completed)},
			{"Remains", strconv.Itoa(remains)},
		}
		t := table.New(table.WithColumns(columns), table.WithRows(rows))
		t.SetHeight(5)

		s = styles.BoxStyle.Render(t.View()) + "\n\n"
	}
	return s
}
