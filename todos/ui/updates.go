package ui

import (
	"github.com/biisal/todo-cli/todos/actions"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func SetUpDefalutKeys(key string, m *TeaModel) {
	switch key {
	case "right":
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
	case "left":
		if m.SelectedIndex > 0 {
			m.SelectedIndex--
		}
	}
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
			switch key {
			case "up", "down", "pgup", "pgdown", "home", "end", "enter":
				listModel, cmd := m.TodoModel.ListModel.List.Update(msg)
				m.TodoModel.ListModel.List = listModel
				m.updateDescriptionContent()
				return m, cmd
			case "j", "k":
				vp, cmd := m.TodoModel.ListModel.DescViewport.Update(msg)
				m.TodoModel.ListModel.DescViewport = vp
				return m, cmd
			}
		case TodoAddMode.Value:
			if m.TodoModel.AddModel.Focus == 0 {
				m.TodoModel.AddModel.TitleInput.Focus()
				m.TodoModel.AddModel.DescInput.Blur()
			} else {
				m.TodoModel.AddModel.TitleInput.Blur()
				m.TodoModel.AddModel.DescInput.Focus()
			}

			// Update active input
			if m.TodoModel.AddModel.Focus == 0 {
				input, cmd := m.TodoModel.AddModel.TitleInput.Update(msg)
				m.TodoModel.AddModel.TitleInput = input
				cmds = append(cmds, cmd)
			} else {
				input, cmd := m.TodoModel.AddModel.DescInput.Update(msg)
				m.TodoModel.AddModel.DescInput = input
				cmds = append(cmds, cmd)
			}

			switch key {
			case "tab":
				if m.TodoModel.AddModel.Focus > 0 {
					m.TodoModel.AddModel.Focus--
				} else {
					m.TodoModel.AddModel.Focus++
				}
			case "ctrl+s":
				if err := m.Add(); err == nil {
					m.RefreshList()
				}

			}
		}

	}
	return m, tea.Batch(cmds...)
}

func (m *TeaModel) updateDescriptionContent() {
	var rightContent = "Description:\n\n"
	if selectedItem := m.TodoModel.ListModel.List.SelectedItem(); selectedItem != nil {
		if i, ok := selectedItem.(item); ok {
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
	todos, _ := actions.GetTodos(true)
	items := []list.Item{}
	for _, todo := range todos {
		items = append(items, item{title: todo.Title, desc: todo.Description})
	}
	if m.TodoModel.ListModel.List.Height() == 0 {
		h, _ := docStyle.GetFrameSize()
		todoList := list.New(items, list.NewDefaultDelegate(), 0, h*3)

		todoList.SetShowHelp(false)
		todoList.Title = "TODO LIST"
		todoList.Styles.Title = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Underline(true)
		todoList.SetShowPagination(true)
		m.TodoModel.ListModel.List = todoList

		listHeight := todoList.Height()
		viewportHeight := listHeight - 4 // Account for border and padding
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		vp := viewport.New(m.Width/2-9, viewportHeight)
		m.TodoModel.ListModel.DescViewport = vp

		m.updateDescriptionContent()
	} else {
		m.TodoModel.ListModel.List.SetItems(items)
	}
}

func (m TeaModel) Add() error {
	_, err := actions.AddTodo(m.TodoModel.AddModel.TitleInput.Value(), m.TodoModel.AddModel.DescInput.Value())
	return err
}
