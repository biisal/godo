package ui

import (
	"strings"

	"github.com/biisal/todo-cli/todos/models"
	"github.com/biisal/todo-cli/todos/ui/setup"
	"github.com/biisal/todo-cli/todos/ui/styles"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	TodoMode     = models.Mode{Value: "todoMode", Label: "Todo Mode"}
	AiMode       = models.Mode{Value: "aiMode", Label: "AI Mode"}
	TodoAddMode  = models.Mode{Value: "todoAddMode", Label: "Add Todo"}
	TodoEditMode = models.Mode{Value: "todoEditMode", Label: "Edit Todo"}
	TodoListMode = models.Mode{Value: "todoListMode", Label: "Todo List"}
)

type item struct {
	id          int
	title, desc string
}

type TeaModel struct {
	Choices       []models.Mode
	SelectedIndex int
	TodoModel     models.TodoModel
	Width         int
	Error         error
}

func getTitleInput(placeholder string) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
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
	idInput.Prompt = "ID >"

	teaModel := TeaModel{
		SelectedIndex: 0,
		Choices:       []models.Mode{TodoMode, AiMode},
		TodoModel: models.TodoModel{
			AddModel: models.TodoForm{
				TitleInput: getTitleInput("Enter title"),
				DescInput:  getDescInput("Enter description"),
				InputCount: 2,
			},
			EditModel: models.TodoForm{
				IdInput:    idInput,
				TitleInput: getTitleInput("Edit title"),
				DescInput:  getDescInput("Edit description"),
				InputCount: 3,
			},
			SelectedIndex: 0,
			Choices:       []models.Mode{TodoListMode, TodoAddMode, TodoEditMode},
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
				if i, ok := selectedItem.(models.Todo); ok {
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
