package ui

import (
	// "fmt"
	"strings"

	"github.com/biisal/todo-cli/todos/models"
	"github.com/biisal/todo-cli/todos/ui/setup"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"

	// "github.com/charmbracelet/bubbles/viewport"
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
}

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	FocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	BlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Italic(true).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#4F6892"))
	BoxStyle     = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89b4fa")).
			Margin(0, 0)
)

func InitialModel() *TeaModel {
	titleInput := textinput.New()
	titleInput.Placeholder = "Enter title"
	titleInput.Focus()

	descInput := textarea.New()
	descInput.Placeholder = "Enter description"

	descInput.FocusedStyle.Base = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#89b4fa")).
		Padding(0, 1)

	descInput.FocusedStyle.CursorLine = lipgloss.NewStyle()

	teaModel := TeaModel{
		SelectedIndex: 0,
		Choices:       []models.Mode{TodoMode, AiMode},
		TodoModel: models.TodoModel{
			AddModel: models.TodoAdd{
				TitleInput: titleInput,
				DescInput:  descInput,
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
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		addModel, listModel := &m.TodoModel.AddModel, &m.TodoModel.ListModel
		addModel.TitleInput.Width = m.Width - 60
		addModel.DescInput.SetWidth(m.Width - 10)

		if listModel.List.Height() > 0 {
			listHeight := listModel.List.Height()
			listModel.DescViewport.Height = listHeight
			listModel.DescViewport.Width = m.Width/2 - 9
		}
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
	s += "TODO CLI\n\n"
	s += setup.SetUpChoice(m.Choices, m.SelectedIndex)
	switch m.Choices[m.SelectedIndex].Value {
	case TodoMode.Value:
		switch m.TodoModel.Choices[m.TodoModel.SelectedIndex].Value {
		case TodoAddMode.Value:
			if m.TodoModel.AddModel.Focus == 0 {
				s += BoxStyle.Render(FocusedStyle.Render(m.TodoModel.AddModel.TitleInput.View())) + "\n"
				s += BlurredStyle.Render(m.TodoModel.AddModel.DescInput.View())
			} else {
				s += BlurredStyle.Render(m.TodoModel.AddModel.TitleInput.View()) + "\n"
				s += FocusedStyle.Render(m.TodoModel.AddModel.DescInput.View())
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
		}
		s += "\n\n"
		s += setup.SetUpChoice(m.TodoModel.Choices, m.TodoModel.SelectedIndex)

	case AiMode.Value:
		s += "AI MODE"
	}

	return s
}

func (m *TeaModel) Init() tea.Cmd {
	return nil
}
