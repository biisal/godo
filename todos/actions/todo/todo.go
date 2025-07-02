package todo

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/biisal/todo-cli/todos/models/todo"
)

var (
	ErrorEmpty     = errors.New("Title or Description can't be empty")
	ErrorInvalidId = errors.New("Invalid ID")
	TodoFilePath   = "./todos.json"
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("Failed to get user home directory: " + err.Error())
	}
	TodoFilePath = home + "/.local/share/godo/agentTodos.json"
}

func GetTodos() ([]todo.Todo, error) {
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

	var todos []todo.Todo
	err = json.NewDecoder(file).Decode(&todos)
	if err != nil {
		if err == io.EOF {
			return []todo.Todo{}, nil
		}
		return nil, err
	}
	return todos, nil
}

func GetTodosCount() string {
	todos, err := GetTodos()
	if err != nil {
		return "Not Found"
	}
	return "Total Todos: " + strconv.Itoa(len(todos))
}

func WriteTodos(todos []todo.Todo) error {
	path := TodoFilePath
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	for i := range todos {
		todos[i].ID = i
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(todos)
}

func AddTodo(title, description string) ([]todo.Todo, error) {
	title, description = strings.TrimSpace(title), strings.TrimSpace(description)
	if title == "" || description == "" {
		return nil, ErrorEmpty
	}
	todos, err := GetTodos()
	if err != nil {
		return nil, err
	}
	todos = append([]todo.Todo{{
		TitleText:       title,
		DescriptionText: description,
		Done:            false,
	}}, todos...)
	WriteTodos(todos)
	return todos, err
}

func DeleteTodo(id int) ([]todo.Todo, error) {
	todos, err := GetTodos()
	if err != nil {
		return nil, err
	}
	if id < 0 || id >= len(todos) {
		return nil, ErrorInvalidId
	}
	todos = slices.Delete(todos, id, id+1)
	WriteTodos(todos)
	return todos, err
}

func ModifyTodo(id int, title, description string) ([]todo.Todo, error) {
	title, description = strings.TrimSpace(title), strings.TrimSpace(description)
	if title == "" || description == "" {
		return nil, ErrorEmpty
	}
	todos, err := GetTodos()
	if err != nil {
		return nil, err
	}
	if id < 0 || id >= len(todos) {
		return nil, ErrorInvalidId
	}
	todos[id] = todo.Todo{
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

func GetTodosInfo() (int, int, int, error) {
	todos, err := GetTodos()
	if err != nil {
		return 0, 0, 0, err
	}
	completd, total := 0, len(todos)
	for _, todo := range todos {
		if todo.Done {
			completd++
		}
	}
	return total, completd, total - completd, nil
}
