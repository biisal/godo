package models

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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

func (i Todo) Title() string       { return i.TitleText }
func (i Todo) Description() string { return i.DescriptionText }
func (i Todo) FilterValue() string { return i.TitleText }
