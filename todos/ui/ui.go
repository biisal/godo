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
	fLogger       *logger.Logger
}

type eventMsg string

func waitForActivity(ev chan string) tea.Cmd {
	return func() tea.Msg {
		return eventMsg(<-ev)
	}
}
func getTitleInput(s ...string) textinput.Model {
	input := textinput.New()
	if len(s) > 0 {
		input.Prompt = s[0]
	}
	if len(s) > 1 {
		input.Placeholder = s[1]
	}
	input.Focus()
	return input
}
func InitialModel(fLogger *logger.Logger) *TeaModel {
	idInput := textinput.New()
	idInput.Prompt = "ID > "

	promptInput := textinput.New()
	promptInput.Focus()
	teaModel := TeaModel{
		fLogger:       fLogger,
		SelectedIndex: 0,
		Choices:       []todo.Mode{TodoMode, AgentMode},
		TodoModel: todo.TodoModel{
			AddModel: todo.TodoForm{
				TitleInput: getTitleInput("Title > ", "Enter todo title"),
				DescInput:  textarea.New(),
				InputCount: 2,
			},
			EditModel: todo.TodoForm{
				IdInput:    idInput,
				TitleInput: getTitleInput("Title > ", "Enter todo title"),
				DescInput:  textarea.New(),
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
	m.TodoModel.ListModel.List, cmd = m.TodoModel.ListModel.List.Update(msg)
	cmds = append(cmds, cmd)
	switch msg := msg.(type) {
	case clearErrorMsg:
		m.Error = nil
		return m, nil
	case tea.WindowSizeMsg:
		UpdateOnSize(msg, m)
		return m, nil
	case eventMsg:
		m.EventMsg = string(msg)
		return m, waitForActivity(config.Cfg.Event)
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

	switch m.Choices[m.SelectedIndex].Value {
	case TodoMode.Value:
		s = TodoView(m, maxHeight)
	case AgentMode.Value:
		s = AgentView(m, maxHeight)
	}
	s = lipgloss.JoinVertical(lipgloss.Center, s, hs)
	return s
}

func (m *TeaModel) Init() tea.Cmd {
	return tea.Batch(
		waitForActivity(config.Cfg.Event), textinput.Blink,
	)
}
