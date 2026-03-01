package ui

import (
	"log/slog"
	"strings"

	"github.com/biisal/godo/internal/config"
	"github.com/biisal/godo/internal/tui/ui/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

func RenderListView(m *TeaModel, maxHeight int) string {
	leftWidth := m.Width * 60 / 100
	left := styles.TodoListStyle.Height(maxHeight).Width(leftWidth).Render(m.TodoModel.ListModel.List.View())

	m.UpdateDescriptionContent()
	right := styles.TodoDescViewportStyle.Width(m.Width - leftWidth - 1).
		BorderForeground(styles.Colors().Border).
		Height(maxHeight).
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
		topPart := styles.CenteredTitleStyle.Width(m.Width).Render("Add Todo")
		topHight := lipgloss.Height(topPart)
		inputView := styles.TodoInputStyle.
			Width(m.Width).
			Render(titleInput.View())
		inputHeight := lipgloss.Height(inputView)
		restHeight := maxHeight - topHight - inputHeight

		descView := styles.TodoDescStyle.Width(m.Width).Height(restHeight)
		descX, descY := descView.GetFrameSize()
		descInput.SetHeight(restHeight - descY)
		descInput.SetWidth(m.Width - descX)

		s = lipgloss.JoinVertical(lipgloss.Center, topPart, inputView, descView.Render(descInput.View()))
	case TodoListMode.Value:
		return RenderListView(m, maxHeight)
	case TodoEditMode.Value:
		titleInput := m.TodoModel.EditModel.TitleInput
		descInput := m.TodoModel.EditModel.DescInput
		idInput := m.TodoModel.EditModel.IdInput
		titleInput.Width = m.Width
		idInput.Width = m.Width
		topPart := styles.CenteredTitleStyle.Width(m.Width).Render("Edit Todo : " + m.TodoModel.EditModel.IdInput.Value())
		topHight := lipgloss.Height(topPart)
		inputView := styles.TodoInputStyle.
			Width(m.Width).
			Render(titleInput.View())
		inputHeight := lipgloss.Height(inputView)

		idInputView := styles.TodoIDInputStyle.
			Width(m.Width).
			Render(idInput.View())
		idInputHeight := lipgloss.Height(idInputView)
		descHeight := maxHeight - topHight - inputHeight - idInputHeight
		descInput.SetHeight(descHeight)
		descInput.SetWidth(m.Width)

		s = lipgloss.JoinVertical(lipgloss.Center, topPart, inputView, descInput.View(), idInputView)

	}
	return s
}

func (m *TeaModel) AgentPromtInputView() (string, int) {
	inputHeight := 1
	marginX := 4
	s := styles.InstructionStyle.Foreground(styles.Colors().Accent).Background(styles.Colors().Secondary).Render(config.Cfg.OPENAI_MODEL)
	s += lipgloss.NewStyle().Background(styles.Colors().Secondary).Render(" " + m.AgentModel.StateText)

	fullInput := lipgloss.JoinVertical(lipgloss.Left, m.AgentModel.PromptInput.View(), s)
	inputView := styles.AgentPromptStyle.
		Width(m.Width - marginX*2).
		Height(inputHeight - 1).
		Padding(1).
		Render(fullInput)

	bottomPart := lipgloss.Place(
		m.Width,
		lipgloss.Height(inputView),
		lipgloss.Center,
		lipgloss.Bottom,
		inputView,
	)
	return bottomPart, lipgloss.Height(bottomPart)
}

func AgentView(m *TeaModel, maxHeight int) string {
	bottomPart, bottomHeight := m.AgentPromtInputView()

	outerChatIViewHeight := maxHeight - bottomHeight

	chatstyle := styles.AgentChatViewStyle.Foreground(styles.Colors().Accent)
	cWidth, cHeight := chatstyle.GetFrameSize()

	chatWidth := m.Width
	// var shellPanel string
	// if m.AgentModel.ShellContent.Len() > 0 {
	// 	chatWidth = m.Width * 60 / 100
	// 	shellWidth := m.Width - chatWidth - 2
	// 	m.AgentModel.ShellViewport.Width = shellWidth
	// 	m.AgentModel.ShellViewport.Height = outerChatIViewHeight - 2
	// 	shellPanel = styles.ShellSidePanelStyle.
	// 		Width(shellWidth).
	// 		Height(outerChatIViewHeight - 2).
	// 		Render(m.AgentModel.ShellViewport.View())
	// }

	m.AgentModel.ChatViewport.Height = outerChatIViewHeight - cHeight
	m.AgentModel.ChatViewport.Width = chatWidth - cWidth
	content := m.ChatContent.String()
	if m.ThinkContent.Len() > 0 {
		content += styles.ThinkingTokenStyle.Render(m.ThinkContent.String())
	}

	if strings.TrimSpace(content) == "" {
		welcomeTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.Colors().Primary).
			MarginBottom(1).
			Render("✦ GODO Agent")
		welcomeHint := styles.InstructionStyle.Render("Type a message below to get started")
		welcomeMsg := lipgloss.JoinVertical(lipgloss.Center, welcomeTitle, welcomeHint)
		content = lipgloss.Place(
			chatWidth-cWidth,
			outerChatIViewHeight-cHeight,
			lipgloss.Center,
			lipgloss.Center,
			welcomeMsg,
		)
	}

	m.AgentModel.ChatViewport.SetContent(wordwrap.String(content, chatWidth-cWidth-1))
	chatView := chatstyle.Width(chatWidth).Render(m.AgentModel.ChatViewport.View())

	// var combinedView string
	// if shellPanel != "" {
	// 	combinedView = lipgloss.JoinHorizontal(lipgloss.Top, chatView, shellPanel)
	// } else {
	// 	combinedView = chatView
	// }
	combinedView := chatView

	topPart := lipgloss.Place(
		m.Width,
		outerChatIViewHeight,
		lipgloss.Left,
		lipgloss.Top,
		combinedView,
	)

	result := lipgloss.JoinVertical(lipgloss.Top, topPart, bottomPart)

	totalHeight := lipgloss.Height(result)
	if totalHeight > maxHeight {
		overflow := totalHeight - maxHeight
		slog.Warn("content exceeds screen height", "overflow", overflow)
	}
	return result
}

func HelpPageView(m *TeaModel, maxHeight int) string {
	leftHelp := `Help

Global:
  ctrl+b     toggle help
  ctrl+a     toggle todo <-> agent 
  ctrl+c     quit
  
Navigation:  
  ctrl+h/←   previous choice
  ctrl+l/→   next choice

Todo List:
  enter      toggle done
  ctrl+e     edit todo
  j/k        next/previous todo 
  `

	rightHelp := `Forms:
  tab        next field
  ctrl+s     save
  
Agent:
  enter      send message
  up/down    scroll chat
  /clear     clear history

Actions:
  delete     remove todo

Press ctrl+u to close`

	style := lipgloss.NewStyle().
		Padding(2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Colors().Border).
		Foreground(styles.Colors().Foreground).
		Width(m.Width * 40 / 100)

	leftPanel := style.Render(leftHelp)
	rightPanel := style.Render(rightHelp)

	combined := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	return lipgloss.Place(m.Width, maxHeight,
		lipgloss.Center, lipgloss.Center,
		combined)
}

func HelpBarView(m *TeaModel) (string, int) {
	var s string
	var modeUI strings.Builder
	if m.Choices[m.SelectedIndex].Value == TodoMode.Value {
		totalChoices := len(m.TodoModel.Choices)
		for i, choice := range m.TodoModel.Choices {
			defalutStyles := styles.ModeStyle
			if i == totalChoices-1 {
				defalutStyles = defalutStyles.MarginRight(1)
			}
			if choice.Value == m.TodoModel.Choices[m.TodoModel.SelectedIndex].Value {
				modeUI.WriteString(defalutStyles.Background(styles.Colors().Primary).Foreground(styles.Colors().PrimaryForeground).Render(choice.Label))
			} else {
				modeUI.WriteString(defalutStyles.Render(choice.Label))
			}
		}
	}
	for _, choice := range m.Choices {
		if choice.Value == m.Choices[m.SelectedIndex].Value {
			modeUI.WriteString(styles.ModeStyle.Background(styles.Colors().Primary).Foreground(styles.Colors().PrimaryForeground).Render(choice.Label))
		} else {
			modeUI.WriteString(styles.ModeStyle.Render(choice.Label))
		}
	}
	rightPart := styles.InstructionStyle.AlignHorizontal(lipgloss.Right).Render(modeUI.String())
	rightWidth := lipgloss.Width(rightPart)

	leftPart := styles.InstructionStyle.Width(m.Width - rightWidth).Render("Help: Ctrl+b")

	s = lipgloss.JoinHorizontal(lipgloss.Top, leftPart, rightPart)
	return s, lipgloss.Height(s)
}

func (m *TeaModel) BuildAgentTextUI(text string, msgType string) {
	defer m.AgentModel.ChatViewport.GotoBottom()
	switch msgType {
	case "stream_start":
		m.ThinkContent.Reset()
	case "stream_end":
		if m.ThinkContent.Len() > 0 {
			m.ChatContent.WriteString(styles.ThinkingTokenStyle.Render(m.ThinkContent.String()))
			m.ThinkContent.Reset()
		}
		m.ChatContent.WriteString(lipgloss.NewStyle().MarginBottom(1).Render())
	case "thinking":
		m.ThinkContent.WriteString(text)
	case "messageStatus":
		if m.ThinkContent.Len() > 0 {
			m.ChatContent.WriteString(styles.ThinkingTokenStyle.Render(m.ThinkContent.String()))
			m.ThinkContent.Reset()
		}
		m.ChatContent.WriteString(styles.StatusOutputStyle.Render(text) + "\n")
		slog.Debug("CHAT CONTENT", "content", map[string]any{
			"fullContent": m.ChatContent.String(),
		})
	default:
		if m.ThinkContent.Len() > 0 {
			m.ChatContent.WriteString(styles.ThinkingTokenStyle.Render(m.ThinkContent.String()))
			m.ThinkContent.Reset()
		}
		m.ChatContent.WriteString(text)
	}
}

func (m *TeaModel) RenderChatContentFromHistory() {
	var chatBlocks []string
	for _, msg := range m.AgentBot.History {
		text := msg.Content
		switch msg.Role {
		case "user":
			chatBlocks = append(chatBlocks, styles.UserContentStyle.Width(m.Width).Render(text))
		default:
			if msg.Reasoning != "" {
				chatBlocks = append(chatBlocks, styles.ThinkingTokenStyle.Render(msg.Reasoning))
			}
			if text != "" {
				chatBlocks = append(chatBlocks, styles.AgentContentStyle.Width(m.Width).Render(text))
			}
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, chatBlocks...)
	m.ChatContent.Reset()
	m.ChatContent.WriteString(content)
	m.ChatContent.WriteString(lipgloss.NewStyle().MarginBottom(1).Render())
	m.AgentModel.ChatViewport.SetContent(content)
	if len(content) > 0 && m.AgentModel.ChatViewport.Height > 0 {
		m.AgentModel.ChatViewport.GotoBottom()
	}
}

// func (m *TeaModel) renderAgentMarkdown(text string) string {
// 	if strings.TrimSpace(text) == "" {
// 		return text
// 	}
//
// 	renderer, err := glamour.NewTermRenderer(
// 		glamour.WithAutoStyle(),
// 		glamour.WithWordWrap(m.Width*70/100),
// 	)
// 	if err != nil {
// 		return styles.AgentContentStyle.Render(text)
// 	}
//
// 	rendered, err := renderer.Render(text)
// 	if err != nil {
// 		return styles.AgentContentStyle.Render(text)
// 	}
//
// 	return strings.TrimRight(rendered, "\n")
// }
