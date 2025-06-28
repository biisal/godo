package models

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

type Todo struct {
	ID          int
	Title       string
	Description string
	Done        bool
}

type Mode struct {
	Value string
	Label string
}

type TodoAdd struct {
	TitleInput textinput.Model
	DescInput  textarea.Model
	Focus      int
}

type TodoList struct {
	List         list.Model
	DescViewport viewport.Model
}

type TodoModel struct {
	AddModel      TodoAdd
	ListModel     TodoList
	Choices       []Mode
	SelectedIndex int
}
