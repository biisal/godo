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
	// return lipgloss.Place(
	// 	m.Width,
	// 	m.Height,
	// 	lipgloss.Left,
	// 	lipgloss.Top,
	// 	lipgloss.NewStyle().
	// 		Width(m.Width).Background(m.Theme.GetBackground()).
	// 		Render(m.TodoModel.ListModel.List.View()),
	// )
	left := lipgloss.NewStyle().Width(m.Width * 60 / 100).PaddingTop(1).Render(m.TodoModel.ListModel.List.View())
	m.UpdateDescriptionContent()
	right := lipgloss.NewStyle().Width(m.Width-lipgloss.Width(left)).Border(lipgloss.RoundedBorder()).
		BorderForeground(m.Theme.GetBorderColor()).
		Height(m.Height).
		BorderLeft(true).BorderBottom(false).
		Padding(1, 0, 0, 1).
		BorderRight(false).BorderTop(false).
		Render(m.TodoModel.ListModel.DescViewport.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
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
	m.fLogger.Println("=== AgentView: Starting render ===")
	m.fLogger.Printf("Screen dimensions - Width: %d, Height: %d", m.Width, m.Height)

	var chatContent string
	historyLen := len(agentaction.History)
	m.fLogger.Printf("Chat history length: %d", historyLen)

	if historyLen == 0 {
		m.fLogger.Println("No history - showing welcome screen")
		logo := styles.Logo + styles.ChatInstructions
		chatContent = styles.PurpleStyle.
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Height(m.AgentModel.ChatViewport.Height).
			Width(m.AgentModel.ChatViewport.Width).
			Render(logo)
	} else {
		m.fLogger.Println("Processing chat history...")
		processedMsgs := 0
		for i, msg := range agentaction.History {
			if msg.IsToolReq {
				m.fLogger.Printf("Skipping tool request message at index %d", i)
				continue
			}

			content := ""
			for _, part := range msg.Parts {
				content += part.Text
			}

			if content == "" {
				m.fLogger.Printf("Skipping empty message at index %d", i)
				continue
			}

			if msg.Role == agent.UserRole {
				m.fLogger.Printf("Rendering user message %d", i)
				chatContent += m.Theme.GetUserContentStyle().Width(m.Width).Render(content) + "\n"
			} else {
				m.fLogger.Printf("Rendering agent message %d", i)
				chatContent += m.Theme.GetAgentContentStyle().Width(m.Width).Render(content)
			}
			processedMsgs++
		}
		m.fLogger.Printf("Processed %d messages", processedMsgs)
	}

	m.AgentModel.ChatViewport.SetContent(chatContent)

	// Chat viewport styling
	chatstyle := lipgloss.NewStyle()

	cWidth, cHeight := chatstyle.GetFrameSize()
	m.fLogger.Printf("Chat frame size - Width: %d, Height: %d", cWidth, cHeight)

	outerHeight := m.AgentModel.ChatViewport.Height
	m.fLogger.Printf("Original viewport height: %d", outerHeight)

	// Adjust viewport for border
	m.AgentModel.ChatViewport.Width = m.Width - cWidth
	m.AgentModel.ChatViewport.Height = outerHeight - cHeight
	m.fLogger.Printf("Adjusted viewport - Width: %d, Height: %d",
		m.AgentModel.ChatViewport.Width, m.AgentModel.ChatViewport.Height)

	chatView := chatstyle.Render(m.AgentModel.ChatViewport.View())

	// Restore viewport height
	m.AgentModel.ChatViewport.Height = outerHeight
	m.fLogger.Printf("Viewport height restored to: %d", outerHeight)

	// Top section
	topPart := lipgloss.Place(
		m.Width,
		outerHeight,
		lipgloss.Center,
		lipgloss.Top,
		chatView,
	)

	// Middle section (instructions)
	middleLeft := m.Theme.GetInstructionStyle().Render("Quit: Esc / Ctrl+C")
	middleRight := m.Theme.GetInstructionStyle().Render("Scroll: PageUp / PageDown")
	middle := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(m.Width/2).Render(middleLeft),
		lipgloss.NewStyle().Width(m.Width/2).AlignHorizontal(lipgloss.Right).Render(middleRight),
	)

	middleHeight := lipgloss.Height(middle)
	m.fLogger.Printf("Middle section height: %d", middleHeight)

	// Bottom section (input)
	inputHeight := m.Height - outerHeight - middleHeight
	m.fLogger.Printf("Input height calculation: %d - %d - %d = %d",
		m.Height, outerHeight, middleHeight, inputHeight)

	if inputHeight < 1 {
		inputHeight = 1
		m.fLogger.Println("Input height set to minimum: 1")
	}

	inputView := styles.ChatInputStyle.
		AlignHorizontal(lipgloss.Left).
		Width(m.Width).
		Height(inputHeight).
		BorderTop(true).
		Border(lipgloss.NormalBorder()).
		BorderBottom(false).
		BorderRight(false).
		BorderLeft(false).
		BorderForeground(lipgloss.Color("#666666")).
		Render(m.AgentModel.PromptInput.View())

	bottomPart := lipgloss.Place(
		m.Width,
		lipgloss.Height(inputView),
		lipgloss.Left,
		lipgloss.Bottom,
		inputView,
	)

	// Join all sections
	result := lipgloss.JoinVertical(lipgloss.Top, topPart, middle, bottomPart)

	// Final verification and adjustment
	totalHeight := lipgloss.Height(result)
	m.fLogger.Printf("Final height - Expected: %d, Actual: %d", m.Height, totalHeight)

	if totalHeight > m.Height {
		overflow := totalHeight - m.Height
		m.fLogger.Printf("WARNING: Content exceeds screen height by %d pixels", overflow)

		// Reduce input height by overflow amount
		adjustedInputHeight := inputHeight - overflow
		if adjustedInputHeight < 1 {
			adjustedInputHeight = 1
		}

		m.fLogger.Printf("Adjusting input height: %d -> %d", inputHeight, adjustedInputHeight)

		// Recreate input with adjusted height
		adjustedInputView := styles.ChatInputStyle.
			AlignHorizontal(lipgloss.Left).
			Width(m.Width).
			Height(adjustedInputHeight).
			BorderTop(true).
			Border(lipgloss.NormalBorder()).
			BorderBottom(false).
			BorderRight(false).
			BorderLeft(false).
			BorderForeground(lipgloss.Color("#666666")).
			Render(m.AgentModel.PromptInput.View())

		adjustedBottomPart := lipgloss.Place(
			m.Width,
			lipgloss.Height(adjustedInputView),
			lipgloss.Left,
			lipgloss.Bottom,
			adjustedInputView,
		)

		// Rejoin with adjusted input
		result = lipgloss.JoinVertical(lipgloss.Top, topPart, middle, adjustedBottomPart)

		finalHeight := lipgloss.Height(result)
		m.fLogger.Printf("After adjustment - Final height: %d", finalHeight)
	}

	m.fLogger.Println("=== AgentView: Render complete ===")
	return result
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

func HelpBarView(m *TeaModel) string {
	var s string
	s += lipgloss.NewStyle().Width(m.Width).Background(m.Theme.GetBackground()).Foreground(lipgloss.Color("#ffffff")).Render("this is a help test")
	m.fLogger.Println("HELPBR TOOK HEIGHT OF ", lipgloss.Height(s))
	return s
}
