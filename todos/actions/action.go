package actions

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/biisal/todo-cli/todos/models"
)

type Action struct {
}

const (
	TodoFilePath = "./todos.json"
)

func GetTodos(reverse ...bool) ([]models.Todo, error) {
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
	if len(reverse) > 0 && reverse[0] {
		slices.Reverse(todos)
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
	todos, err := GetTodos()
	if err != nil {
		return nil, err
	}
	todos = append(todos, models.Todo{
		ID:              len(todos) + 1,
		TitleText:       title,
		DescriptionText: description,
		Done:            false,
	})
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

func ModifyTodo(id int, title, description string) {
	todos, err := GetTodos()
	if err != nil {
		return
	}
	todos[id] = models.Todo{
		ID:              id,
		TitleText:       title,
		DescriptionText: description,
		Done:            false,
	}
	WriteTodos(todos)
}

func ToggleDone(id int) {
	todos, err := GetTodos()
	if err != nil {
		return
	}
	todos[id].Done = !todos[id].Done
	WriteTodos(todos)
}
