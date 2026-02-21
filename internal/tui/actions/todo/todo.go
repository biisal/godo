package todo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/biisal/godo/internal/config"
	"github.com/biisal/godo/internal/tui/models/todo"
)

var (
	ErrorEmpty     = errors.New("Title or Description can't be empty")
	ErrorInvalidId = errors.New("Invalid ID")
)

func GetTodos() ([]todo.Todo, error) {
	sqlStmt := `
	SELECT Id , Title, Description, Done
	FROM todos
	ORDER BY Id DESC
	`
	rows, err := config.Cfg.DB.Query(sqlStmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	todos := []todo.Todo{}
	for rows.Next() {
		var t todo.Todo
		if err := rows.Scan(&t.ID, &t.TitleText, &t.DescriptionText, &t.Done); err != nil {
			return nil, err
		}
		todos = append(todos, t)
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

func AddTodo(title, description string) ([]todo.Todo, error) {
	title, description = strings.TrimSpace(title), strings.TrimSpace(description)
	if title == "" || description == "" {
		return nil, ErrorEmpty
	}
	sqlStmt := `
	INSERT INTO todos (Title, Description, Done)
	VALUES (?, ?, ?)`
	if _, err := config.Cfg.DB.Exec(sqlStmt, title, description, false); err != nil {
		return nil, err
	}
	return GetTodos()

}

func DeleteTodo(id int) ([]todo.Todo, error) {
	sqlStmt := `
	DELETE FROM todos WHERE Id = ?`
	if _, err := config.Cfg.DB.Exec(sqlStmt, id); err != nil {
		return nil, err
	}
	return GetTodos()
}

func ModifyTodo(id int, title, description string) ([]todo.Todo, error) {
	title, description = strings.TrimSpace(title), strings.TrimSpace(description)
	if title == "" || description == "" {
		return nil, ErrorEmpty
	}
	sqlStmt := `
	UPDATE todos SET Title = ?, Description = ? WHERE Id = ?`
	if _, err := config.Cfg.DB.Exec(sqlStmt, title, description, id); err != nil {
		return nil, err
	}
	todos, err := GetTodos()
	if err != nil {
		return nil, err
	}
	return todos, nil
}

func ToggleDone(id int, doneStatus ...bool) (bool, error) {
	sqlStmt := `
	UPDATE todos SET Done = NOT Done WHERE Id = ?`
	if _, err := config.Cfg.DB.Exec(sqlStmt, id); err != nil {
		return false, err
	}
	var isDone bool
	err := config.Cfg.DB.QueryRow(`SELECT Done FROM todos WHERE Id = ?`, id).Scan(&isDone)
	if err != nil {
		return false, err
	}
	return isDone, nil
}

func GetTodosInfo() (int, int, int, error) {
	todos, err := GetTodos()
	if err != nil {
		return 0, 0, 0, err
	}
	completed, total := 0, len(todos)
	for _, t := range todos {
		if t.Done {
			completed++
		}
	}
	return total, completed, total - completed, nil
}

func GetTodoById(id int) (*todo.Todo, error) {

	sqlStmt := `
	SELECT Id , Title, Description, Done
	FROM todos
	WHERE Id = ?
	`
	row := config.Cfg.DB.QueryRow(sqlStmt, id)
	t := &todo.Todo{}
	if err := row.Scan(&t.ID, &t.TitleText, &t.DescriptionText, &t.Done); err != nil {
		return nil, err
	}
	return t, nil
}

func PerformSqlQuery(sqlStmt string) (string, error) {
	rows, err := config.Cfg.DB.Query(sqlStmt)
	if err != nil {
		return "", err
	}
	if rows == nil {
		return "[]", nil
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}

	// use sql.RawBytes — zero-copy, reuses the driver buffer
	ptrs := make([]any, len(cols))
	rawBytes := make([]sql.RawBytes, len(cols))
	for i := range rawBytes {
		ptrs[i] = &rawBytes[i]
	}

	var results []map[string]any

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return "", err
		}

		row := make(map[string]any, len(cols))
		for i, col := range cols {
			if rawBytes[i] == nil {
				row[col] = nil
			} else {
				row[col] = string(rawBytes[i]) // copy before next iteration
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	out, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}
	slog.Info("\n\nQuery Result", "query", sqlStmt, "result", string(out))
	return string(out), nil
}
