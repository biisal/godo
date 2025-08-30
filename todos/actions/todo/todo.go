package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/biisal/todo-cli/config"
	"github.com/biisal/todo-cli/todos/models/todo"
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
		var todo todo.Todo
		if err := rows.Scan(&todo.ID, &todo.TitleText, &todo.DescriptionText, &todo.Done); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
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
	path := config.TodoFilePath
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
	isDone := false
	config.Cfg.DB.QueryRow(`SELECT Done FROM todos WHERE Id = ?`, id).Scan(&isDone)
	return isDone, nil

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

func GetTodoById(id int) (*todo.Todo, error) {

	sqlStmt := `
	SELECT Id , Title, Description, Done
	FROM todos
	WHERE Id = ?
	`
	row := config.Cfg.DB.QueryRow(sqlStmt, id)
	todo := &todo.Todo{}
	if err := row.Scan(&todo.ID, &todo.TitleText, &todo.DescriptionText, &todo.Done); err != nil {
		return nil, err
	}
	return todo, nil
}

type QueryResult struct {
	Status        string           `json:"status"`
	QueryType     string           `json:"query_type"`
	Data          []map[string]any `json:"data,omitempty"`
	RowsAffected  int64            `json:"rows_affected,omitempty"`
	RowCount      int              `json:"row_count,omitempty"`
	ExecutionTime string           `json:"execution_time"`
	Columns       []string         `json:"columns,omitempty"`
	Error         *QueryError      `json:"error,omitempty"`
}

type QueryError struct {
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

func PerformSqlQuery(sqlStmt string) (*QueryResult, error) {
	startTime := time.Now()

	result := &QueryResult{
		Status:        "success",
		QueryType:     getQueryType(sqlStmt),
		ExecutionTime: "",
	}

	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sqlStmt)), "SELECT") {
		return handleSelectQuery(sqlStmt, result, startTime)
	} else {
		return handleModificationQuery(sqlStmt, result, startTime)
	}
}

func handleSelectQuery(sqlStmt string, result *QueryResult, startTime time.Time) (*QueryResult, error) {
	rows, err := config.Cfg.DB.Query(sqlStmt)
	if err != nil {
		result.Status = "error"
		result.Error = &QueryError{
			Message:    err.Error(),
			Suggestion: getSuggestionForError(err.Error()),
		}
		result.ExecutionTime = time.Since(startTime).String()
		return result, nil // Return result with error info, not Go error
	}
	defer rows.Close()

	// Get column information
	columns, err := rows.Columns()
	if err != nil {
		result.Status = "error"
		result.Error = &QueryError{Message: fmt.Sprintf("Failed to get columns: %v", err)}
		result.ExecutionTime = time.Since(startTime).String()
		return result, nil
	}
	result.Columns = columns

	// Prepare slice to hold column values
	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	// Read all rows
	var data []map[string]any
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			result.Status = "error"
			result.Error = &QueryError{
				Message:    fmt.Sprintf("Failed to scan row: %v", err),
				Suggestion: "Check if query columns match expected data types",
			}
			result.ExecutionTime = time.Since(startTime).String()
			return result, nil
		}

		// Convert row to map
		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for better JSON serialization
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		data = append(data, row)
	}

	result.Data = data
	result.RowCount = len(data)
	result.ExecutionTime = time.Since(startTime).String()
	return result, nil
}

func handleModificationQuery(sqlStmt string, result *QueryResult, startTime time.Time) (*QueryResult, error) {
	res, err := config.Cfg.DB.Exec(sqlStmt)
	if err != nil {
		result.Status = "error"
		result.Error = &QueryError{
			Message:    err.Error(),
			Suggestion: getSuggestionForError(err.Error()),
		}
		result.ExecutionTime = time.Since(startTime).String()
		return result, nil
	}

	rowsAffected, err := res.RowsAffected()
	if err == nil {
		result.RowsAffected = rowsAffected
	}

	result.ExecutionTime = time.Since(startTime).String()
	return result, nil
}

func getQueryType(sqlStmt string) string {
	stmt := strings.ToUpper(strings.TrimSpace(sqlStmt))
	if strings.HasPrefix(stmt, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(stmt, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(stmt, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(stmt, "DELETE") {
		return "DELETE"
	} else if strings.HasPrefix(stmt, "CREATE") {
		return "CREATE"
	} else if strings.HasPrefix(stmt, "ALTER") {
		return "ALTER"
	} else if strings.HasPrefix(stmt, "DROP") {
		return "DROP"
	}
	return "UNKNOWN"
}

func getSuggestionForError(errorMsg string) string {
	errorMsg = strings.ToLower(errorMsg)

	if strings.Contains(errorMsg, "no such table") {
		return "Check if the table name exists and is spelled correctly"
	} else if strings.Contains(errorMsg, "no such column") {
		return "Verify the column name exists in the table schema"
	} else if strings.Contains(errorMsg, "syntax error") {
		return "Review SQL syntax - check for missing commas, quotes, or keywords"
	} else if strings.Contains(errorMsg, "expected") && strings.Contains(errorMsg, "destination arguments") {
		return "The number of columns in SELECT doesn't match the Scan destination"
	} else if strings.Contains(errorMsg, "constraint") {
		return "Check foreign key constraints or unique constraints"
	}
	return "Review the query for common SQL errors"
}
