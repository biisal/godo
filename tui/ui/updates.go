package ui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	// "time"

	"github.com/biisal/godo/config"
	agentAction "github.com/biisal/godo/tui/actions/agent"
	todoAction "github.com/biisal/godo/tui/actions/todo"
	"github.com/biisal/godo/tui/models/todo"
	"github.com/biisal/godo/tui/ui/styles"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var WrongTypeIdError = errors.New("ID Should be a number")

func SetUpDefalutKeys(key string, m *TeaModel) {
	switch key {
	case "ctrl+b":
		m.IsShowHelp = !m.IsShowHelp
	case "ctrl+a":
		m.ToggleMode()
	case "ctrl+left", "ctrl+h":
		maxChoice := len(m.TodoModel.Choices)
		if m.TodoModel.SelectedIndex == 0 && maxChoice > 0 {
			m.TodoModel.SelectedIndex = maxChoice - 1
			return
		}
		m.TodoModel.SelectedIndex--
	case "ctrl+right", "ctrl+l":
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
				*cmds = append(*cmds, m.ShowError(err))
				return
			}
			m.RefreshList()
			m.TodoModel.SelectedIndex = 0
			model.TitleInput.Reset()
			model.DescInput.Reset()
		case 2:
			id, err := strconv.Atoi(m.TodoModel.EditModel.IdInput.Value())
			if err != nil {
				*cmds = append(*cmds, m.ShowError(WrongTypeIdError))
				return
			}
			_, err = todoAction.ModifyTodo(id, m.TodoModel.EditModel.TitleInput.Value(), m.TodoModel.EditModel.DescInput.Value())
			if err != nil {
				*cmds = append(*cmds, m.ShowError(err))
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
				m.FLogger.Info("ERROR GETTING ITEM WITH id ", id, " : ", err)
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
				cmd := m.ShowError(err)
				return m, &cmd
			}
			m.RefreshList()
		}
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
			promtInput := strings.TrimSpace(m.AgentModel.PromptInput.Value())
			if promtInput == "" {
				return m, nil
			}
			if promtInput == "/clear" {
				if err := agentAction.TruncateChats(); err != nil {
					return m, m.ShowError(err)
				}
				m.ChatContent.Reset()
				m.AgentModel.PromptInput.Reset()
				return m, nil
			}

			config.StreamResponse <- config.StreamMsg{IsUser: true, Text: promtInput}
			return m, tea.Cmd(func() tea.Msg {
				_, refresh, err := agentAction.AgentResponse(promtInput, m.FLogger)
				return agentResponseMsg{
					refresh: refresh,
					err:     err,
				}
			})
		}
		input, cmd := m.AgentModel.PromptInput.Update(msg)
		m.AgentModel.PromptInput = input
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func UpdateOnSize(msg tea.WindowSizeMsg, m *TeaModel) tea.Cmd {
	m.Width = msg.Width
	m.Height = msg.Height
	m.RefreshList()
	m.RenderChatContentFromHistory()
	return nil
}

func (m *TeaModel) UpdateDescriptionContent() {
	var LabelStyle = lipgloss.NewStyle().Background(m.Theme.TitleBackround()).Padding(0, 1).Foreground(lipgloss.Color("#000000")).Bold(true)
	var rightContent string
	if selectedItem := m.TodoModel.ListModel.List.SelectedItem(); selectedItem != nil {
		m.FLogger.Info("SELECTED ITEM", selectedItem)
		if i, ok := selectedItem.(todo.Todo); ok {
			m.FLogger.Info("MATCHED", i)
			statusText := "Not Compelted"
			if i.Done {
				statusText = "Compelted"
			}
			rightContent = fmt.Sprintf("%s : %s\n\n%s : %s ", LabelStyle.Render("Status"), statusText, LabelStyle.Render("Description"), i.Description())
		}
	}
	m.FLogger.Info("RIGHT CONTENT IS : ", rightContent)
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
	m.FLogger.Info("THE WIDTH IS :", innerWidth)
	m.TodoModel.ListModel.List = list.New(items, todo.CustomDelegate{Width: innerWidth - 2, Theme: styles.Theme{}}, 0, 0)
	m.TodoModel.ListModel.List.SetSize(innerWidth, innerHeight)
	m.TodoModel.ListModel.List.Title = "Todos "
	m.TodoModel.ListModel.List.SetShowStatusBar(false)
}

func (m *TeaModel) ToggleMode() {
	if len(m.Choices) > m.SelectedIndex+1 {
		m.SelectedIndex++
	} else {
		m.SelectedIndex = 0
	}
	switch m.Choices[m.SelectedIndex].Value {
	case TodoMode.Value: {
		config.Cfg.MODE = "todo"
	}
	case AgentMode.Value: {
		config.Cfg.MODE = "agent"
	}
	}
	config.SaveCfg()
}

type clearErrorMsg struct{}

func (m *TeaModel) ShowError(err error) tea.Cmd {
	m.Error = err
	return tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m *TeaModel) Exit(cmds *[]tea.Cmd) {
}
