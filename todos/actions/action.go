package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/biisal/todo-cli/todos/models"
)

type Action struct {
}

const (
	TodoFilePath = "./todos.json"
)

func GetTodos() ([]models.Todo, error) {
	path := TodoFilePath
	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
		if _, err = os.Create(path); err != nil {
			return nil, err
		}
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var todos []models.Todo
	err = json.NewDecoder(file).Decode(&todos)
	if err != nil {
		if err == io.EOF {
			return []models.Todo{}, nil
		}
		return nil, err
	}
	return todos, nil
}

func WriteTodos(todos []models.Todo) error {
	path := TodoFilePath
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(todos)
}

func AddTodo(title, description string) ([]models.Todo, error) {
	title, description = strings.TrimSpace(title), strings.TrimSpace(description)
	if title == "" || description == "" {
		return nil, fmt.Errorf("title and description cannot be empty")
	}
	todos, err := GetTodos()
	if err != nil {
		return nil, err
	}
	todos = append([]models.Todo{{
		TitleText:       title,
		DescriptionText: description,
		Done:            false,
	}}, todos...)
	for i := range todos {
		todos[i].ID = i
	}
	WriteTodos(todos)
	return todos, err
}

func DeleteTodo(id int) ([]models.Todo, error) {
	todos, err := GetTodos()
	if err != nil {
		return nil, err
	}
	todos = append(todos[:id-1], todos[id:]...)
	WriteTodos(todos)
	return todos, err
}

func ModifyTodo(id int, title, description string) ([]models.Todo, error) {
	title, description = strings.TrimSpace(title), strings.TrimSpace(description)
	if title == "" || description == "" {
		return nil, fmt.Errorf("title and description cannot be empty")
	}
	todos, err := GetTodos()
	if err != nil {
		return nil, err
	}
	if id < 0 || id >= len(todos) {
		return nil, fmt.Errorf("invalid id")
	}
	todos[id] = models.Todo{
		ID:              id,
		TitleText:       title,
		DescriptionText: description,
		Done:            false,
	}
	if err = WriteTodos(todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func ToggleDone(id int) {
	todos, err := GetTodos()
	if err != nil {
		return
	}
	todos[id].Done = !todos[id].Done
	WriteTodos(todos)
}
