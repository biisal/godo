package ui

import (
	"github.com/biisal/todo-cli/config"
	"github.com/biisal/todo-cli/logger"
	"github.com/biisal/todo-cli/todos/models/agent"
	"github.com/biisal/todo-cli/todos/models/todo"

	"github.com/biisal/todo-cli/todos/ui/styles"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func listenForStreamUpdates(logger *logger.Logger) tea.Cmd {
	return func() tea.Msg {
		// This will block until a message is available
		logger.Debug("Listening for stream updates")

		msg := <-config.StreamResponse
		logger.Debug("Received message: ", msg)
		return StreamMsg{Text: msg}
	}
}

type StreamMsg struct {
	Text string
}

var (
	TodoMode     = todo.Mode{Value: "todoMode", Label: "Todo Mode"}
	AgentMode    = todo.Mode{Value: "agentMode", Label: "Agent Mode"}
	TodoAddMode  = todo.Mode{Value: "todoAddMode", Label: "Add Todo"}
	TodoEditMode = todo.Mode{Value: "todoEditMode", Label: "Edit Todo"}
	TodoListMode = todo.Mode{Value: "todoListMode", Label: "Todo List"}
)

type TeaModel struct {
	IsExiting     bool
	EventMsg      string
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
	ChatContent   string
}

type eventMsg string

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
		SelectedIndex: 0,
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
	case StreamMsg:
		m.FLogger.Debug("Received message in Update: ", msg.Text)
		if msg.Text == "DONE" {
			m.AgentModel.StreamChunk = ""
		} else {
			m.AgentModel.StreamChunk += msg.Text
		}
		return m, listenForStreamUpdates(m.FLogger)
	case clearErrorMsg:
		m.FLogger.Debug("Clearing Error")
		m.Error = nil
		return m, nil

	case tea.WindowSizeMsg:
		UpdateOnSize(msg, m)
		return m, nil
	case eventMsg:
		m.EventMsg = string(msg)
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

	switch m.Choices[m.SelectedIndex].Value {
	case TodoMode.Value:
		s += TodoView(m, maxHeight)
	case AgentMode.Value:
		s += AgentView(m, maxHeight)
	}
	s = lipgloss.JoinVertical(lipgloss.Center, errorView, s, hs)
	return s
}

func (m *TeaModel) Init() tea.Cmd {
	return tea.Batch(
		listenForStreamUpdates(m.FLogger),
		textinput.Blink,
	)
}
