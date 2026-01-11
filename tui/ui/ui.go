package ui

import (
	"strings"

	"github.com/biisal/godo/config"
	"github.com/biisal/godo/logger"
	"github.com/biisal/godo/tui/models/agent"
	"github.com/biisal/godo/tui/models/todo"

	"github.com/biisal/godo/tui/ui/styles"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type agentResponseMsg struct {
	refresh bool
	err     error
}

var (
	TodoMode     = todo.Mode{Value: "todoMode", Label: "Todo Mode"}
	AgentMode    = todo.Mode{Value: "agentMode", Label: "Agent Mode"}
	TodoAddMode  = todo.Mode{Value: "todoAddMode", Label: "Add Todo"}
	TodoEditMode = todo.Mode{Value: "todoEditMode", Label: "Edit Todo"}
	TodoListMode = todo.Mode{Value: "todoListMode", Label: "Todo List"}
)

type TeaModel struct {
	IsShowHelp    bool
	Choices       []todo.Mode
	SelectedIndex int
	TodoModel     todo.TodoModel
	AgentModel    agent.AgentModel
	Width         int
	Height        int
	Error         error
	Theme         styles.Theme
	BgStyle       lipgloss.Style
	FLogger       *logger.Logger
	ChatContent   strings.Builder
}

//	func waitForActivity(ev chan string) tea.Cmd {
//		return func() tea.Msg {
//			return eventMsg(<-ev)
//		}
//	}
func getDescInput(placeholder string) textarea.Model {
	input := textarea.New()
	input.Placeholder = placeholder
	return input
}
func getTitleInput(focus bool, s ...string) textinput.Model {
	input := textinput.New()
	if len(s) > 0 {
		input.Prompt = s[0]
	}
	if len(s) > 1 {
		input.Placeholder = s[1]
	}
	if focus {
		input.Focus()
	}
	return input
}
func InitialModel(fLogger *logger.Logger) *TeaModel {

	promptInput := textinput.New()
	promptInput.Focus()
	teaModel := TeaModel{
		FLogger:       fLogger,
		SelectedIndex: 1,
		Choices:       []todo.Mode{TodoMode, AgentMode},
		TodoModel: todo.TodoModel{
			AddModel: todo.TodoForm{
				TitleInput: getTitleInput(true, "Title > ", "Enter todo title"),
				DescInput:  getDescInput("Enter todo description"),
				InputCount: 2,
			},
			EditModel: todo.TodoForm{
				IdInput:    getTitleInput(false, "Id > ", "Enter todo id"),
				TitleInput: getTitleInput(true, "Title > ", "Enter todo title"),
				DescInput:  getDescInput("Enter todo description"),
				InputCount: 3,
			},
			SelectedIndex: 0,
			Choices:       []todo.Mode{TodoListMode, TodoAddMode, TodoEditMode},
		},
		AgentModel: agent.AgentModel{
			PromptInput:  promptInput,
			ChatViewport: viewport.Model{},
		},
	}

	initialMode := teaModel.Choices[teaModel.SelectedIndex]
	if (initialMode == TodoMode && config.Cfg.MODE == "agent") ||
		(initialMode == AgentMode && config.Cfg.MODE == "todo") {
		teaModel.ToggleMode()
	}

	teaModel.RefreshList()
	return &teaModel
}

func (m *TeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	if m.Choices[m.SelectedIndex].Value == TodoMode.Value && m.TodoModel.Choices[m.TodoModel.SelectedIndex].Value == TodoListMode.Value {
		m.TodoModel.ListModel.List, cmd = m.TodoModel.ListModel.List.Update(msg)
		cmds = append(cmds, cmd)
	}
	switch msg := msg.(type) {
	case agentResponseMsg:
		if msg.refresh {
			m.RefreshList()
		}
		if msg.err != nil {
			m.ChatContent.WriteString(m.Theme.GetErrorInChatStyle().Width(m.Width).Render(msg.err.Error()) + "\n")
			m.AgentModel.ChatViewport.SetContent(m.ChatContent.String())
			m.AgentModel.ChatViewport.GotoBottom()
		} else {
			m.AgentModel.PromptInput.Reset()
		}
		return m, nil
	case config.StreamMsg:
		if msg.IsUser {
			m.ChatContent.WriteString(m.Theme.GetUserContentStyle().Width(m.Width).Render(msg.Text) + "\n")
			return m, nil
		}
		m.FLogger.Debug("Received to Update message: ", msg.Text)
		m.BuildAgentTextUI(msg.Text)
		return m, nil
	case clearErrorMsg:
		m.FLogger.Debug("Clearing Error")
		m.Error = nil
		return m, nil

	case tea.WindowSizeMsg:
		UpdateOnSize(msg, m)
		return m, nil
	case tea.KeyMsg:
		_, cmd := UpdateOnKey(msg, m)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *TeaModel) View() string {
	var s string
	hs, helpBarHeight := HelpBarView(m)

	maxHeight := m.Height - helpBarHeight
	var errorView string
	if m.Error != nil {
		errorView = m.Theme.GetEorrorStyle().Width(m.Width * 50 / 100).AlignHorizontal(lipgloss.Left).Render(m.Error.Error())
		maxHeight -= lipgloss.Height(errorView)
	}

	if m.IsShowHelp {
		s += HelpPageView(m, maxHeight)
	} else {
		switch m.Choices[m.SelectedIndex].Value {
		case TodoMode.Value:
			s += TodoView(m, maxHeight)
		case AgentMode.Value:
			s += AgentView(m, maxHeight)
		}
	}
	s = lipgloss.JoinVertical(lipgloss.Center, errorView, s, hs)
	return s
}

func (m *TeaModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
	)
}
