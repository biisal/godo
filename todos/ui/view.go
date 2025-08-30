package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/biisal/todo-cli/config"
	agentaction "github.com/biisal/todo-cli/todos/actions/agent"
	todoaction "github.com/biisal/todo-cli/todos/actions/todo"
	"github.com/biisal/todo-cli/todos/models/agent"

	"github.com/biisal/todo-cli/todos/ui/styles"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func RenderListView(m *TeaModel, maxHeight int) string {
	left := lipgloss.NewStyle().Height(maxHeight).Width(m.Width * 60 / 100).Render(m.TodoModel.ListModel.List.View())
	m.UpdateDescriptionContent()
	right := lipgloss.NewStyle().Width(m.Width-lipgloss.Width(left)).Border(lipgloss.RoundedBorder()).
		BorderForeground(m.Theme.GetBorderColor()).
		Height(maxHeight).
		BorderLeft(true).BorderBottom(false).
		Padding(1, 0, 0, 1).
		BorderRight(false).BorderTop(false).
		Render(m.TodoModel.ListModel.DescViewport.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}
func TodoView(m *TeaModel, maxHeight int) string {
	var s string
	switch m.TodoModel.Choices[m.TodoModel.SelectedIndex].Value {
	case TodoAddMode.Value:
		titleInput := m.TodoModel.AddModel.TitleInput
		descInput := m.TodoModel.AddModel.DescInput
		titleInput.Width = m.Width
		topPart := lipgloss.NewStyle().Align(lipgloss.Center).Width(m.Width).Render("Add Todo")
		topHight := lipgloss.Height(topPart)
		inputView := styles.ChatInputStyle.
			AlignHorizontal(lipgloss.Left).
			Width(m.Width).
			Height(5).
			BorderTop(true).
			Border(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderRight(false).
			BorderLeft(false).
			BorderForeground(lipgloss.Color("#666666")).
			Render(titleInput.View())
		inputHeight := lipgloss.Height(inputView)
		restHeight := maxHeight - topHight - inputHeight
		descInput.SetHeight(restHeight)
		descInput.SetWidth(m.Width)

		s = lipgloss.JoinVertical(lipgloss.Center, topPart, inputView, descInput.View())
	case TodoListMode.Value:
		return RenderListView(m, maxHeight)
	case TodoEditMode.Value:
		titleInput := m.TodoModel.EditModel.TitleInput
		descInput := m.TodoModel.EditModel.DescInput
		idInput := m.TodoModel.EditModel.IdInput
		titleInput.Width = m.Width
		idInput.Width = m.Width
		topPart := lipgloss.NewStyle().Align(lipgloss.Center).Width(m.Width).Render("Edit Todo : " + m.TodoModel.EditModel.IdInput.Value())
		topHight := lipgloss.Height(topPart)
		inputView := styles.ChatInputStyle.
			AlignHorizontal(lipgloss.Left).
			Width(m.Width).
			Height(5).
			BorderTop(true).
			Border(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderRight(false).
			BorderLeft(false).
			BorderForeground(lipgloss.Color("#666666")).
			Render(titleInput.View())
		inputHeight := lipgloss.Height(inputView)

		idInputView := styles.ChatInputStyle.
			AlignHorizontal(lipgloss.Left).
			Width(m.Width).
			Height(1).
			BorderTop(true).
			Border(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderRight(false).
			BorderLeft(false).
			BorderForeground(lipgloss.Color("#666666")).
			Render(idInput.View())
		idInputHeight := lipgloss.Height(idInputView)
		descHeight := maxHeight - topHight - inputHeight - idInputHeight
		descInput.SetHeight(descHeight)
		descInput.SetWidth(m.Width)

		s = lipgloss.JoinVertical(lipgloss.Center, topPart, inputView, descInput.View(), idInputView)

	}
	return s
}

func AgentView(m *TeaModel, maxHeight int) string {
	var chatContent string
	historyLen := len(agentaction.History)

	if historyLen == 0 {
		m.fLogger.Info("No history - showing welcome screen")
		logo := styles.Logo + styles.ChatInstructions
		chatContent = styles.PurpleStyle.
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Height(m.AgentModel.ChatViewport.Height).
			Width(m.AgentModel.ChatViewport.Width).
			Render(logo)
	} else {
		m.fLogger.Info("Processing chat history...")
		processedMsgs := 0
		for i, msg := range agentaction.History {
			if msg.IsToolReq {
				m.fLogger.FInfo("Skipping tool request message at index %d", i)
				continue
			}

			content := ""
			for _, part := range msg.Parts {
				content += part.Text
			}

			if content == "" {
				m.fLogger.FInfo("Skipping empty message at index %d", i)
				continue
			}

			if msg.Role == agent.UserRole {
				m.fLogger.FInfo("Rendering user message %d", i)
				chatContent += m.Theme.GetUserContentStyle().Width(m.Width).Render(content) + "\n"
			} else {
				m.fLogger.FInfo("Rendering agent message %d", i)
				chatContent += m.Theme.GetAgentContentStyle().Width(m.Width).Render(content)
			}
			processedMsgs++
		}
		m.fLogger.FInfo("Processed %d messages", processedMsgs)
	}
	outerChatIViewHeight := maxHeight * 80 / 100
	m.AgentModel.ChatViewport.SetContent(chatContent)

	chatstyle := lipgloss.NewStyle()
	cWidth, cHeight := chatstyle.GetFrameSize()

	m.AgentModel.ChatViewport.Width = m.Width - cWidth
	m.AgentModel.ChatViewport.Height = outerChatIViewHeight - cHeight

	chatView := chatstyle.Render(m.AgentModel.ChatViewport.View())

	// Top section
	topPart := lipgloss.Place(
		m.Width,
		outerChatIViewHeight,
		lipgloss.Center,
		lipgloss.Top,
		chatView,
	)

	middleLeft := m.Theme.GetInstructionStyle().Render("Quit: Esc / Ctrl+C")
	middleRight := m.Theme.GetInstructionStyle().Render("Scroll: PageUp / PageDown")
	middle := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(m.Width/2).Render(middleLeft),
		lipgloss.NewStyle().Width(m.Width/2).AlignHorizontal(lipgloss.Right).Render(middleRight),
	)

	middleHeight := lipgloss.Height(middle)
	topHeight := lipgloss.Height(topPart)

	inputHeight := maxHeight - topHeight - middleHeight
	inputView := styles.ChatInputStyle.
		AlignHorizontal(lipgloss.Left).
		Width(m.Width).
		Height(inputHeight - 1).
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

	result := lipgloss.JoinVertical(lipgloss.Top, topPart, middle, bottomPart)

	totalHeight := lipgloss.Height(result)
	m.fLogger.FDebug("Final height - Expected: %d, Actual: %d", maxHeight, totalHeight)
	m.fLogger.FDebug("Input Height %d | Input View Height %d", inputHeight, lipgloss.Height(inputView))
	if totalHeight > maxHeight {
		overflow := totalHeight - maxHeight
		m.fLogger.FWarn("WARNING: Content exceeds screen height by %d pixels", overflow)
	}

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

func HelpBarView(m *TeaModel) (string, int) {
	var s string
	modeUi := ""
	if m.Choices[m.SelectedIndex].Value == TodoMode.Value {
		totalChoices := len(m.TodoModel.Choices)
		for i, choice := range m.TodoModel.Choices {
			defalutStyles := m.Theme.ModeStyle()
			if i == totalChoices-1 {
				m.fLogger.Debug("Adding margin right")
				defalutStyles = defalutStyles.MarginRight(1)
			}
			if choice.Value == m.TodoModel.Choices[m.TodoModel.SelectedIndex].Value {
				modeUi += defalutStyles.Background(m.Theme.ListTheme.SelectedColor()).Render(choice.Label)
			} else {
				modeUi += defalutStyles.Render(choice.Label)
			}
		}
	}
	for _, choice := range m.Choices {
		if choice.Value == m.Choices[m.SelectedIndex].Value {
			modeUi += m.Theme.ModeStyle().Background(m.Theme.ListTheme.SelectedColor()).Render(choice.Label)
		} else {
			modeUi += m.Theme.ModeStyle().Render(choice.Label)
		}
	}
	rightPart := m.Theme.GetInstructionStyle().AlignHorizontal(lipgloss.Right).Render(modeUi)
	rightWidth := lipgloss.Width(rightPart)

	leftPart := m.Theme.GetInstructionStyle().Width(m.Width - rightWidth).Render("Quit: Esc / Ctrl+C")

	s = lipgloss.JoinHorizontal(lipgloss.Top, leftPart, rightPart)
	m.fLogger.Debug("HELPBR TOOK HEIGHT OF ", lipgloss.Height(s))
	return s, lipgloss.Height(s)
}
