package todo

import (
	"fmt"
	"io"

	"github.com/biisal/godo/todos/ui/styles"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

type CustomDelegate struct {
	Theme             styles.Theme
	Width             int
	Bg                lipgloss.Color
	Forground         lipgloss.Color
	SelectedForground lipgloss.Color
}

func (d CustomDelegate) Height() int                               { return 2 }
func (d CustomDelegate) Spacing() int                              { return 1 }
func (d CustomDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d CustomDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(Todo)
	if !ok {
		return
	}

	title := item.Title()
	desc := item.Description()

	rowStyle := d.Theme.ListTheme.RowStyle().Margin(0, 0).Padding(0, 2).BorderLeft(false)

	if item.Done && index != m.Index() {
		rowStyle = rowStyle.Foreground(d.Theme.GetDarkGreenColor())
	} else if item.Done && index == m.Index() {
		rowStyle = rowStyle.BorderForeground(d.Theme.GetGreenColor()).Padding(0, 1).BorderLeft(true).Foreground(d.Theme.GetGreenColor())

	} else if index == m.Index() {
		rowStyle = rowStyle.BorderLeft(true).
			Padding(0, 1).
			Foreground(d.Theme.ListTheme.SelectedColor())
	}
	cropedDesc := desc
	padding := 8
	if d.Width > padding && len(desc) >= d.Width-padding {
		cropedDesc = desc[:d.Width-padding] + "..."
	}
	row := rowStyle.Render(fmt.Sprintf("%s\n%s", title, cropedDesc))
	fmt.Fprint(w, row)
}
