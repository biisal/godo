package ui

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	agentAction "github.com/biisal/todo-cli/todos/actions/agent"
	todoAction "github.com/biisal/todo-cli/todos/actions/todo"
	"github.com/biisal/todo-cli/todos/models/todo"
	"github.com/biisal/todo-cli/todos/ui/styles"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var WrongTypeIdError = errors.New("ID Should be a number")

func SetUpDefalutKeys(key string, m *TeaModel) {
	switch key {
	case "ctrl+a":
		if len(m.Choices) > m.SelectedIndex+1 {
			m.SelectedIndex++
		} else {
			m.SelectedIndex = 0
		}
	case "ctrl+left", "ctrl+b":
		maxChoice := len(m.TodoModel.Choices)
		if m.TodoModel.SelectedIndex == 0 && maxChoice > 0 {
			m.TodoModel.SelectedIndex = maxChoice - 1
			return
		}
		m.TodoModel.SelectedIndex--
	case "ctrl+right", "ctrl+n":
		maxChoice := len(m.TodoModel.Choices)
		currentIndex := m.TodoModel.SelectedIndex
		m.TodoModel.SelectedIndex = (currentIndex + 1) % maxChoice
	}
}

func SetUpFormKey(key string, model *todo.TodoForm, m *TeaModel, cmds *[]tea.Cmd, msg tea.Msg) {
	switch key {
	case "tab":
		if model.Focus < model.InputCount-1 {
			model.Focus++
		} else {
			model.Focus = 0
		}
	case "ctrl+s":
		switch m.TodoModel.SelectedIndex {
		case 1:
			_, err := todoAction.AddTodo(m.TodoModel.AddModel.TitleInput.Value(), m.TodoModel.AddModel.DescInput.Value())
			if err != nil {
				m.ShowError(err, cmds)
				return
			}
			m.RefreshList()
			m.TodoModel.SelectedIndex = 0
			model.TitleInput.Reset()
			model.DescInput.Reset()
		case 2:
			id, err := strconv.Atoi(m.TodoModel.EditModel.IdInput.Value())
			if err != nil {
				m.ShowError(WrongTypeIdError, cmds)
				return
			}
			_, err = todoAction.ModifyTodo(id, m.TodoModel.EditModel.TitleInput.Value(), m.TodoModel.EditModel.DescInput.Value())
			if err != nil {
				m.ShowError(err, cmds)
				return
			}
			m.RefreshList()
			m.TodoModel.SelectedIndex = 0
			model.TitleInput.Reset()
			model.DescInput.Reset()
			model.IdInput.Reset()
		}
	}
	switch model.Focus {
	case 0:
		model.TitleInput.Focus()
		model.DescInput.Blur()
		model.IdInput.Blur()
		input, cmd := model.TitleInput.Update(msg)
		model.TitleInput = input
		*cmds = append(*cmds, cmd)
	case 1:
		model.DescInput.Focus()
		model.TitleInput.Blur()
		model.IdInput.Blur()
		input, cmd := model.DescInput.Update(msg)
		model.DescInput = input
		*cmds = append(*cmds, cmd)
	case 2:
		model.IdInput.Focus()
		model.DescInput.Blur()
		model.TitleInput.Blur()
		input, cmd := model.IdInput.Update(msg)
		model.IdInput = input
		*cmds = append(*cmds, cmd)
	}

}

func SetyUpListKey(key string, m *TeaModel, msg tea.KeyMsg, cmds *[]tea.Cmd) (tea.Model, *tea.Cmd) {
	switch key {
	case "enter":
		selected := m.TodoModel.ListModel.List.SelectedItem()
		if selected != nil {
			id := selected.(todo.Todo).ID
			todoAction.ToggleDone(id)
			item, err := todoAction.GetTodoById(id)
			if err != nil {
				m.fLogger.Info("ERROR GETTING ITEM WITH id ", id, " : ", err)
				return m, nil
			}
			m.TodoModel.ListModel.List.SetItem(m.TodoModel.ListModel.List.Index(), *item)
		}
	case "ctrl+e":
		selected := m.TodoModel.ListModel.List.SelectedItem()
		if selected != nil {
			todo := selected.(todo.Todo)
			m.TodoModel.SelectedIndex = 2
			m.TodoModel.EditModel.IdInput.SetValue(fmt.Sprintf("%d", todo.ID))
			m.TodoModel.EditModel.TitleInput.SetValue(todo.Title())
			m.TodoModel.EditModel.DescInput.SetValue(todo.Description())
		}
	case "delete":
		selected := m.TodoModel.ListModel.List.SelectedItem()
		if selected != nil {
			_, err := todoAction.DeleteTodo(selected.(todo.Todo).ID)
			if err != nil {
				m.ShowError(err, nil)
				return m, nil
			}
			m.RefreshList()
		}
	case "up", "down", "pgup", "pgdown", "home", "end":
		listModel, cmd := m.TodoModel.ListModel.List.Update(msg)
		m.TodoModel.ListModel.List = listModel
		m.UpdateDescriptionContent()
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
	if key == "ctrl+c" {
		m.IsExiting = true
		return m, tea.Quit
	}
	switch m.Choices[m.SelectedIndex].Value {
	case TodoMode.Value:
		switch m.TodoModel.Choices[m.TodoModel.SelectedIndex].Value {
		case TodoListMode.Value:
			m, c := SetyUpListKey(key, m, msg, &cmds)
			if c != nil {
				return m, *c
			}
		case TodoAddMode.Value:
			SetUpFormKey(key, &m.TodoModel.AddModel, m, &cmds, msg)
		case TodoEditMode.Value:
			SetUpFormKey(key, &m.TodoModel.EditModel, m, &cmds, msg)
		}
	case AgentMode.Value:
		switch key {
		case "up":
			m.AgentModel.ChatViewport.ScrollUp(1)
		case "down":
			m.AgentModel.ChatViewport.ScrollDown(1)
		case "enter":
			return m, tea.Cmd(func() tea.Msg {
				_, refresh, err := agentAction.AgentResponse(m.AgentModel.PromptInput.Value(), m.fLogger)
				if refresh {
					m.RefreshList()
				}
				if err != nil {
					m.ShowError(err, &cmds)
				}
				m.AgentModel.PromptInput.Reset()
				return m
			})
		}
		input, cmd := m.AgentModel.PromptInput.Update(msg)
		m.AgentModel.PromptInput = input
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func UpdateOnSize(msg tea.WindowSizeMsg, m *TeaModel) {
	m.Width = msg.Width
	m.Height = msg.Height
	m.RefreshList()
}

func (m *TeaModel) UpdateDescriptionContent() {
	var LabelStyle = lipgloss.NewStyle().Background(m.Theme.TitleBackround()).Padding(0, 1).Foreground(lipgloss.Color("#000000")).Bold(true)
	var rightContent string
	if selectedItem := m.TodoModel.ListModel.List.SelectedItem(); selectedItem != nil {
		m.fLogger.Info("SELECTED ITEM", selectedItem)
		if i, ok := selectedItem.(todo.Todo); ok {
			m.fLogger.Info("MATCHED", i)
			statusText := "Not Compelted"
			if i.Done {
				statusText = "Compelted"
			}
			rightContent = fmt.Sprintf("%s : %s\n\n%s : %s ", LabelStyle.Render("Status"), statusText, LabelStyle.Render("Description"), i.Description())
		}
	}
	m.fLogger.Info("RIGHT CONTENT IS : ", rightContent)
	m.TodoModel.ListModel.DescViewport.Height = m.Height * 50 / 100
	m.TodoModel.ListModel.DescViewport.SetContent(rightContent)
}

func (m *TeaModel) RefreshList() {
	todos, _ := todoAction.GetTodos()
	doneCount := 0
	items := []list.Item{}
	for _, todo := range todos {
		if todo.Done {
			doneCount++
		}
		items = append(items, todo)
	}
	innerWidth := m.Width * 60 / 100
	innerHeight := m.Height * 80 / 100
	m.fLogger.Info("THE WIDTH IS :", innerWidth)
	m.TodoModel.ListModel.List = list.New(items, todo.CustomDelegate{Width: innerWidth - 2, Theme: styles.Theme{}}, 0, 0)
	m.TodoModel.ListModel.List.SetSize(innerWidth, innerHeight)
	m.TodoModel.ListModel.List.Title = "Todos "
}

type clearErrorMsg struct{}

func (m *TeaModel) ShowError(err error, cmds *[]tea.Cmd) {
	m.Error = err
	*cmds = append(*cmds, func() tea.Msg {
		time.Sleep(4 * time.Second)
		return clearErrorMsg{}
	})
}

func (m *TeaModel) Exit(cmds *[]tea.Cmd) {
}
