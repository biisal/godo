package ui

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	// "github.com/biisal/todo-cli/config"
	agentAction "github.com/biisal/todo-cli/todos/actions/agent"
	todoAction "github.com/biisal/todo-cli/todos/actions/todo"
	"github.com/biisal/todo-cli/todos/models/todo"
	"github.com/biisal/todo-cli/todos/ui/styles"

	// "github.com/biisal/todo-cli/todos/ui/styles"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	// "github.com/charmbracelet/lipgloss"
	// "github.com/charmbracelet/bubbles/spinner"
	// "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	case "ctrl+right":
		if len(m.TodoModel.Choices) > m.TodoModel.SelectedIndex+1 {
			m.TodoModel.SelectedIndex++
		}
	case "ctrl+left":
		if m.TodoModel.SelectedIndex > 0 {
			m.TodoModel.SelectedIndex--
		}
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
				m.fLogger.Println("ERROR GETTING ITEM WITH id ", id, " : ", err)
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
	if key == "esc" || key == "ctrl+c" {
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
	m.fLogger.Printf("=== UpdateOnSize: Window resized to %dx%d ===", msg.Width, msg.Height)

	// m.BgStyle = lipgloss.NewStyle().Background(m.Theme.GetBackground()).Width(msg.Width).Height(msg.Height)
	// v, h := m.BgStyle.GetFrameSize()

	m.Width = msg.Width
	m.Height = msg.Height - 2

	// m.fLogger.Printf("Background frame size - Width: %d, Height: %d", v, h)
	m.fLogger.Printf("Usable area - Width: %d, Height: %d", m.Width, m.Height)

	// Set chat viewport to 80% of available height
	chatViewportHeight := m.Height * 80 / 100
	m.AgentModel.ChatViewport.Width = m.Width
	m.AgentModel.ChatViewport.Height = chatViewportHeight

	m.fLogger.Printf("Chat viewport set to - Width: %d, Height: %d (80%% of %d)",
		m.AgentModel.ChatViewport.Width, m.AgentModel.ChatViewport.Height, m.Height)

	m.RefreshList()
	m.fLogger.Println("=== UpdateOnSize: Complete ===")
}

func (m *TeaModel) UpdateDescriptionContent() {
	var LabelStyle = lipgloss.NewStyle().Background(m.Theme.TitleBackround()).Padding(0, 1).Foreground(lipgloss.Color("#000000")).Bold(true)
	var rightContent string
	if selectedItem := m.TodoModel.ListModel.List.SelectedItem(); selectedItem != nil {
		m.fLogger.Println("SELECTED ITEM", selectedItem)
		if i, ok := selectedItem.(todo.Todo); ok {
			m.fLogger.Println("MATCHED", i)
			statusText := "Not Compelted"
			if i.Done {
				statusText = "Compelted"
			}
			rightContent = fmt.Sprintf("%s : %s\n\n%s : %s ", LabelStyle.Render("Status"), statusText, LabelStyle.Render("Description"), i.Description())
		}
	}
	m.fLogger.Println("RIGHT CONTENT IS : ", rightContent)
	m.TodoModel.ListModel.DescViewport.Height = m.Height * 50 / 100
	m.TodoModel.ListModel.DescViewport.SetContent(rightContent)
}

func (m *TeaModel) RefreshList() {
	defer func() {
		// m.updateDescriptionContent()
		// m.TodoModel.ListModel.List.SetShowStatusBar(false)

	}()
	todos, _ := todoAction.GetTodos()
	doneCount := 0
	// totalTodos := len(todos)
	items := []list.Item{}
	for _, todo := range todos {
		if todo.Done {
			doneCount++
		}
		items = append(items, todo)
	}
	innerWidth := m.Width * 60 / 100
	innerHeight := m.Height * 80 / 100
	m.fLogger.Println("THE WIDTH IS :", innerWidth)
	m.TodoModel.ListModel.List = list.New(items, todo.CustomDelegate{Width: innerWidth - 2, Theme: styles.Theme{}}, 0, 0)
	// m.TodoModel.ListModel.List = list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.TodoModel.ListModel.List.SetSize(innerWidth, innerHeight)
	m.TodoModel.ListModel.List.Title = "Todos "
	// m.TodoModel.ListModel.List = list.New(items, todo.CustomDelegate{Width: width, Bg: m.Theme.GetBackground(), Forground: m.Theme.GetGrayColor(), SelectedForground: m.Theme.GetPurpleColoe()}, width, m.Height)
	// m.TodoModel.ListModel.List = list.New(items, list.NewDefaultDelegate(), width, height)
	// m.TodoModel.ListModel.List.SetSize(width, height)
	// m.TodoModel.ListModel.List.SetHeight(m.Height * 80 / 100)
	// defaultStyle := lipgloss.NewStyle().Width(width).MarginLeft(2).Background(m.Theme.GetBackground())
	// m.TodoModel.ListModel.List.Styles = list.Styles{
	// 	StatusBar:             defaultStyle,
	// 	PaginationStyle:       defaultStyle,
	// 	ActivePaginationDot:   defaultStyle,
	// 	InactivePaginationDot: defaultStyle,
	// 	StatusEmpty:           defaultStyle,
	// 	Spinner:               defaultStyle,
	// 	HelpStyle:             defaultStyle,
	// 	Title:                 defaultStyle.Padding(1, 1).Width(0).MarginLeft(2),
	// 	ArabicPagination:      defaultStyle,
	// }
	// if totalTodos == 0 {
	// 	return
	// }
	// statusText := fmt.Sprintf("\n\n[%d/%d] Done", doneCount, len(todos))
	// var status string
	// if doneCount > totalTodos/2 {
	// 	status = styles.GreenStyle.Background(m.Theme.GetBackground()).Width(m.TodoModel.ListModel.List.Width()).Render(statusText)
	// } else {
	// 	status = styles.RedStyle.Background(m.Theme.GetBackground()).Render(statusText)
	// }
	// m.TodoModel.ListModel.List.NewStatusMessage(status)

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
	// *cmds = append(*cmds, func() tea.Msg {
	// 	time.Sleep(time.Second * 2)
	// 	return tea.Quit()
	// })
}
