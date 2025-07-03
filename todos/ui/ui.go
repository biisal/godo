package ui

import (
	"github.com/biisal/todo-cli/config"
	"github.com/biisal/todo-cli/todos/models/agent"
	"github.com/biisal/todo-cli/todos/models/todo"

	"github.com/biisal/todo-cli/todos/ui/setup"
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

type item struct {
	id          int
	title, desc string
}

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
}

func getTitleInput(prompt, placeholder string) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.Prompt = prompt
	input.Focus()
	return input
}

func getDescInput(placeholder string) textarea.Model {
	input := textarea.New()
	input.Placeholder = placeholder
	input.FocusedStyle.Base = styles.DescStyle
	input.FocusedStyle.CursorLine = lipgloss.NewStyle()
	return input
}

type eventMsg string

func waitForActivity(ev chan string) tea.Cmd {
	return func() tea.Msg {
		return eventMsg(<-ev)
	}
}

func InitialModel() *TeaModel {
	idInput := textinput.New()
	idInput.Prompt = "ID > "

	teaModel := TeaModel{
		SelectedIndex: 0,
		Choices:       []todo.Mode{TodoMode, AgentMode},
		TodoModel: todo.TodoModel{
			AddModel: todo.TodoForm{
				TitleInput: getTitleInput("Title > ", "Enter title"),
				DescInput:  getDescInput("Enter description"),
				InputCount: 2,
			},
			EditModel: todo.TodoForm{
				IdInput:    idInput,
				TitleInput: getTitleInput("Title > ", "Edit title"),
				DescInput:  getDescInput("Edit description"),
				InputCount: 3,
			},
			SelectedIndex: 0,
			Choices:       []todo.Mode{TodoListMode, TodoAddMode, TodoEditMode},
		},
		AgentModel: agent.AgentModel{
			PromptInput: getTitleInput("> ", "Ask Agent and perform todo tasks.."),
			ChatViewport: viewport.Model{
				Height: 0,
				Width:  0,
			},
		},
	}
	teaModel.RefreshList()
	return &teaModel
}

func (m *TeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case agent.AgentResTeaMsg:
		if msg.Error != nil {
			m.ShowError(msg.Error, &cmds)
			return m, nil
		}
		m.AgentModel.PromptInput.SetValue("")
		m.AgentModel.History = msg.History
		return m, nil
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
	s := styles.TitleStyle.Render(styles.Logo) + "\n\n"
	if m.IsExiting {
		s += ExitView()
		return s
	}

	s += setup.SetUpChoice(m.Choices, m.SelectedIndex, "alt+right/left")
	switch m.Choices[m.SelectedIndex].Value {
	case TodoMode.Value:
		s += TodoView(m)
	case AgentMode.Value:
		s += AgentView(m)
	}

	if m.Error != nil {
		s += "\n\n" + styles.ErrorStyle.Width(m.Width-20).Render(m.Error.Error())
	}
	return s
}

func (m *TeaModel) Init() tea.Cmd {
	return tea.Batch(
		waitForActivity(config.Cfg.Event),
	)
}
