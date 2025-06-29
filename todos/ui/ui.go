package ui

import (
	"github.com/biisal/todo-cli/todos/models/todo"

	"github.com/biisal/todo-cli/todos/ui/setup"
	"github.com/biisal/todo-cli/todos/ui/styles"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	TodoMode     = todo.Mode{Value: "todoMode", Label: "Todo Mode"}
	AiMode       = todo.Mode{Value: "aiMode", Label: "AI Mode"}
	TodoAddMode  = todo.Mode{Value: "todoAddMode", Label: "Add Todo"}
	TodoEditMode = todo.Mode{Value: "todoEditMode", Label: "Edit Todo"}
	TodoListMode = todo.Mode{Value: "todoListMode", Label: "Todo List"}
)

type item struct {
	id          int
	title, desc string
}

type TeaModel struct {
	Choices       []todo.Mode
	SelectedIndex int
	TodoModel     todo.TodoModel
	Width         int
	Error         error
}

func getTitleInput(placeholder string) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.Prompt = "Title > "
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

func InitialModel() *TeaModel {
	idInput := textinput.New()
	idInput.Prompt = "ID > "

	teaModel := TeaModel{
		SelectedIndex: 0,
		Choices:       []todo.Mode{TodoMode, AiMode},
		TodoModel: todo.TodoModel{
			AddModel: todo.TodoForm{
				TitleInput: getTitleInput("Enter title"),
				DescInput:  getDescInput("Enter description"),
				InputCount: 2,
			},
			EditModel: todo.TodoForm{
				IdInput:    idInput,
				TitleInput: getTitleInput("Edit title"),
				DescInput:  getDescInput("Edit description"),
				InputCount: 3,
			},
			SelectedIndex: 0,
			Choices:       []todo.Mode{TodoListMode, TodoAddMode, TodoEditMode},
		},
	}
	teaModel.RefreshList()
	return &teaModel
}

func (m *TeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case clearErrorMsg:
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
	s := styles.TitleStyle.Render(styles.Logo)
	s += setup.SetUpChoice(m.Choices, m.SelectedIndex, "alt+right/left")
	switch m.Choices[m.SelectedIndex].Value {
	case TodoMode.Value:
		s += TodoView(m)
	case AiMode.Value:
		s += "AI MODE"
	}

	if m.Error != nil {
		s += "\n\n" + styles.ErrorStyle.Render(m.Error.Error())
	}
	return s
}

func (m *TeaModel) Init() tea.Cmd {
	return nil
}
