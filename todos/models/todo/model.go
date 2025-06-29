package todo

import (
	"fmt"
	"io"
	"strings"

	"github.com/biisal/todo-cli/todos/ui/styles"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type Todo struct {
	ID              int    `json:"id"`
	TitleText       string `json:"title"`
	DescriptionText string `json:"description"`
	Done            bool   `json:"done"`
}

type Mode struct {
	Value string
	Label string
}

type TodoForm struct {
	IdInput    textinput.Model
	TitleInput textinput.Model
	DescInput  textarea.Model
	Focus      int
	InputCount int
}

type TodoList struct {
	List         list.Model
	DescViewport viewport.Model
}

type TodoModel struct {
	AddModel      TodoForm
	ListModel     TodoList
	EditModel     TodoForm
	Choices       []Mode
	SelectedIndex int
}

func (i Todo) Title() string       { return i.TitleText }
func (i Todo) Description() string { return i.DescriptionText }
func (i Todo) FilterValue() string { return i.TitleText }

type TodoListDelegate struct{}

func (d TodoListDelegate) Height() int { return 1 }

func (d TodoListDelegate) Spacing() int                            { return 1 }
func (d TodoListDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d TodoListDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Todo)
	if !ok {
		return
	}
	maxWidth := max(0, m.Width()-8)
	if len(i.TitleText) > maxWidth {
		i.TitleText = strings.TrimSpace(i.Title()[:maxWidth]) + "..."
	}
	if len(i.Description()) > maxWidth {
		i.DescriptionText = strings.TrimSpace(i.Description()[:maxWidth]) + "..."
	}
	doneStr := "○"
	if i.Done {
		doneStr = "✓"
	}
	str := fmt.Sprintf("[%s] %s\n    %s", doneStr, i.TitleText, i.DescriptionText)
	fn := styles.ItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return styles.SelectedItemStyle.Render(strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
