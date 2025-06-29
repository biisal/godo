package ui

import (
	"fmt"
	"time"

	"github.com/biisal/todo-cli/todos/actions"
	"github.com/biisal/todo-cli/todos/models"
	"github.com/biisal/todo-cli/todos/ui/styles"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func SetUpDefalutKeys(key string, m *TeaModel) {
	switch key {
	case "alt+right":
		if len(m.Choices) > m.SelectedIndex+1 {
			m.SelectedIndex++
		}
	case "ctrl+right":
		if len(m.TodoModel.Choices) > m.TodoModel.SelectedIndex+1 {
			m.TodoModel.SelectedIndex++
		}
	case "ctrl+left":
		if m.TodoModel.SelectedIndex > 0 {
			m.TodoModel.SelectedIndex--
		}
	case "alt+left":
		if m.SelectedIndex > 0 {
			m.SelectedIndex--
		}
	}
}

func SetUpFormKey(key string, model *models.TodoForm, m *TeaModel, cmds *[]tea.Cmd, msg tea.Msg) {
	switch key {
	case "tab":
		if model.Focus < model.InputCount-1 {
			model.Focus++
		} else {
			model.Focus = 0
		}
	case "ctrl+s":
		if err := m.Add(); err == nil {
			m.RefreshList()
			m.TodoModel.SelectedIndex = 0
			model.TitleInput.Reset()
			model.DescInput.Reset()
			m.ShowError(fmt.Errorf("This is a test"), cmds)
		}
	}
	switch model.Focus {
	case 0:
		model.TitleInput.Focus()
		model.DescInput.Blur()
		model.ID.Blur()
		input, cmd := model.TitleInput.Update(msg)
		model.TitleInput = input
		*cmds = append(*cmds, cmd)
	case 1:
		model.DescInput.Focus()
		model.TitleInput.Blur()
		model.ID.Blur()
		input, cmd := model.DescInput.Update(msg)
		model.DescInput = input
		*cmds = append(*cmds, cmd)
	case 2:
		model.ID.Focus()
		model.DescInput.Blur()
		model.TitleInput.Blur()
		input, cmd := model.ID.Update(msg)
		model.ID = input
		*cmds = append(*cmds, cmd)
	}

}

func SetyUpListKey(key string, m *TeaModel, msg tea.KeyMsg) (tea.Model, *tea.Cmd) {
	switch key {
	case "enter":
		selected := m.TodoModel.ListModel.List.SelectedItem()
		if selected != nil {
			id := selected.(models.Todo).ID
			actions.ToggleDone(id)
			m.RefreshList()
		}
	case "ctrl+e":
		selected := m.TodoModel.ListModel.List.SelectedItem()
		if selected != nil {
			todo := selected.(models.Todo)
			m.TodoModel.SelectedIndex = 2
			m.TodoModel.EditModel.ID.SetValue(fmt.Sprintf("%d", todo.ID))
			m.TodoModel.EditModel.TitleInput.SetValue(todo.Title())
			m.TodoModel.EditModel.DescInput.SetValue(todo.Description())
		}
	case "up", "down", "pgup", "pgdown", "home", "end":
		listModel, cmd := m.TodoModel.ListModel.List.Update(msg)
		m.TodoModel.ListModel.List = listModel
		m.updateDescriptionContent()
		return m, &cmd
	case "j", "k":
		vp, cmd := m.TodoModel.ListModel.DescViewport.Update(msg)
		m.TodoModel.ListModel.DescViewport = vp
		return m, &cmd
	}
	return m, nil
}

func UpdateOnKey(msg tea.KeyMsg, m *TeaModel) (tea.Model, tea.Cmd) {
	key := msg.String()
	var cmds []tea.Cmd
	defer SetUpDefalutKeys(key, m)
	if key == "esc" || key == "ctrl+c" {
		return m, tea.Quit
	}
	switch m.Choices[m.SelectedIndex].Value {
	case TodoMode.Value:
		switch m.TodoModel.Choices[m.TodoModel.SelectedIndex].Value {
		case TodoListMode.Value:
			m, c := SetyUpListKey(key, m, msg)
			if c != nil {
				return m, *c
			}
		case TodoAddMode.Value:
			SetUpFormKey(key, &m.TodoModel.AddModel, m, &cmds, msg)
		case TodoEditMode.Value:
			SetUpFormKey(key, &m.TodoModel.EditModel, m, &cmds, msg)
		}
	}
	return m, tea.Batch(cmds...)
}

func UpdateOnSize(msg tea.WindowSizeMsg, m *TeaModel) {
	m.Width = msg.Width
	titleWidth, descWidth, idWidth := m.Width-60, m.Width-10, m.Width-70
	addModel, listModel, editModel := &m.TodoModel.AddModel, &m.TodoModel.ListModel, &m.TodoModel.EditModel
	addModel.TitleInput.Width = titleWidth
	addModel.DescInput.SetWidth(descWidth)

	editModel.TitleInput.Width = titleWidth
	editModel.DescInput.SetWidth(descWidth)
	editModel.ID.Width = idWidth
	if listModel.List.Height() > 0 {
		listHeight := listModel.List.Height()
		listModel.DescViewport.Height = listHeight
		listModel.DescViewport.Width = m.Width/2 - 9
	}

}

type clearErrorMsg struct{}

func (m *TeaModel) ShowError(err error, cmds *[]tea.Cmd) {
	m.Error = err
	*cmds = append(*cmds, func() tea.Msg {
		time.Sleep(4 * time.Second)
		return clearErrorMsg{}
	})
}

func (m *TeaModel) updateDescriptionContent() {
	var rightContent = "Description:\n\n"
	if selectedItem := m.TodoModel.ListModel.List.SelectedItem(); selectedItem != nil {
		if i, ok := selectedItem.(models.Todo); ok {
			desc := i.Description()
			wrappedStyle := lipgloss.NewStyle().
				Width(m.TodoModel.ListModel.DescViewport.Width)
			wrappedDesc := wrappedStyle.Render(desc)
			rightContent += wrappedDesc
		}
	}
	m.TodoModel.ListModel.DescViewport.SetContent(rightContent)
}

func (m *TeaModel) RefreshList() {
	defer func() {
		m.updateDescriptionContent()
		m.TodoModel.ListModel.List.SetShowStatusBar(false)

	}()
	todos, _ := actions.GetTodos()
	totalTodos := len(todos)
	items := []list.Item{}
	for _, todo := range todos {
		items = append(items, todo)
	}
	if m.TodoModel.ListModel.List.Height() == 0 {
		h, _ := styles.DocStyle.GetFrameSize()
		todoList := list.New(items, models.TodoListDelegate{}, 0, h*3)
		sp := spinner.New()
		todoList.SetShowHelp(false)
		todoList.Title = "TODO LIST"
		todoList.Styles.Title = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Underline(true)
		todoList.SetShowPagination(true)
		todoList.SetSpinner(sp.Spinner)

		m.TodoModel.ListModel.List = todoList
		m.TodoModel.ListModel.List.ToggleSpinner()

		listHeight := todoList.Height()
		viewportHeight := listHeight - 4 // Account for border and padding
		vp := viewport.New(m.Width/2-9, viewportHeight)
		m.TodoModel.ListModel.DescViewport = vp

	} else {
		m.TodoModel.ListModel.List.SetItems(items)
	}
	if totalTodos == 0 {
		return
	}
	doneCount := 0
	for _, todo := range todos {
		if todo.Done {
			doneCount++
		}
	}
	statusText := fmt.Sprintf("\n\n[%d/%d] Done", doneCount, len(todos))
	var status string
	if doneCount > totalTodos/2 {
		status = styles.GreenStyle.Render(statusText)
	} else {
		status = styles.RedStyle.Render(statusText)
	}
	m.TodoModel.ListModel.List.NewStatusMessage(status)

}

func (m TeaModel) Add() error {
	_, err := actions.AddTodo(m.TodoModel.AddModel.TitleInput.Value(), m.TodoModel.AddModel.DescInput.Value())
	return err
}
