package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/biisal/todo-cli/config"
	"github.com/biisal/todo-cli/todos/models/agent"
	"github.com/biisal/todo-cli/todos/models/todo"
)

var (
	ErrorEmpty     = errors.New("Title or Description can't be empty")
	ErrorInvalidId = errors.New("Invalid ID")
)

const (
	TodoFilePath = "./todos.json"
)

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

var (
	GetTodoCountFunc = "GetTodosCount"
	AddTodoFunc      = "AddTodo"
	DeleteTodoFunc   = "DeleteTodo"
	ModifyTodoFunc   = "ModifyTodo"
	ToggleDoneFunc   = "ToggleDone"
)

func aiAPICall(history []agent.Message) (*agent.AgentRes, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	history = append([]agent.Message{
		{
			Role:    "system",
			Content: "You are a very helpfull agent that can help me with my tasks",
		},
	}, history...)
	url := "https://api.groq.com/openai/v1/chat/completions"
	body := agent.AgentReq{
		Messages:            history,
		Model:               "llama-3.1-8b-instant",
		Temperature:         1,
		MaxCompletionTokens: 100,
		TopP:                1,
		Stream:              false,
		Tools: []agent.Tool{
			{
				Type: "function",

				Function: agent.FunctionReq{
					Name:        GetTodoCountFunc,
					Description: "Returns the number of todos",
					Parameters: struct {
						Type       string                        `json:"type"`
						Properties map[string]agent.PropertyType `json:"properties"`
						Required   []string                      `json:"required"`
					}{
						Type:       "object",
						Properties: map[string]agent.PropertyType{}, // ✅ Empty map
						Required:   []string{},
					},
				},
			},
			{
				Type: "function",
				Function: agent.FunctionReq{
					Name:        AddTodoFunc,
					Description: "Add a todo to the list",
					Parameters: struct {
						Type       string                        `json:"type"`
						Properties map[string]agent.PropertyType `json:"properties"`
						Required   []string                      `json:"required"`
					}{
						Type: "object",
						Properties: map[string]agent.PropertyType{
							"title": {
								Type:        "string",
								Description: "The title of the todo",
							},
							"description": {
								Type:        "string",
								Description: "The description of the todo",
							},
						},
						Required: []string{"title", "description"},
					},
				},
			},
		},
		ToolChoice: "auto",
	}

	bodyJson, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Cfg.GROQ_API_KEY)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got status code %d\nResponse: %s", resp.StatusCode, resp.Body)
	}
	var msgStruct agent.AgentRes
	data, _ := io.ReadAll(resp.Body)
	if err = json.Unmarshal(data, &msgStruct); err != nil {
		return nil, err
	}

	return &msgStruct, nil
}

func AiResponse(history []agent.Message) ([]agent.Message, bool, error) {
	agentRes, err := aiAPICall(history)
	refrsh := false
	if err != nil {
		return nil, refrsh, err
	}
	if len(agentRes.Choices) == 0 {
		return history, refrsh, nil
	}
	tools := agentRes.Choices[0].Message.ToolCalls
	if len(tools) > 0 {
		funcName := strings.TrimSpace(tools[0].Function.Name)
		switch funcName {
		case GetTodoCountFunc:
			result := GetTodosCount()
			history = append(history,
				agent.Message{
					Role:       "tool",
					Content:    result,
					ToolCallId: tools[0].Id,
					Name:       GetTodoCountFunc,
				},
			)
			agentRes, err = aiAPICall(history)
			if err != nil {
				return nil, refrsh, err
			}
			history = append(history, agentRes.Choices[0].Message)
		case AddTodoFunc:
			var args struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			}
			if err = json.Unmarshal([]byte(tools[0].Function.Arguments), &args); err != nil {
				return nil, refrsh, fmt.Errorf("invalid JSON in tool call arguments: %w\nraw: %s", err, tools[0].Function.Arguments)
			}
			_, err := AddTodo(args.Title, args.Description)
			if err != nil {
				return nil, refrsh, err
			}
			refrsh = true
			resultStr := fmt.Sprintf("Added todo with title: %s and description: %s", args.Title, args.Description)
			history = append(history,
				agent.Message{
					Role:       "tool",
					Content:    resultStr,
					ToolCallId: tools[0].Id,
					Name:       AddTodoFunc,
				},
			)
			agentRes, err = aiAPICall(history)
			if err != nil {
				return nil, refrsh, err
			}
			history = append(history, agentRes.Choices[0].Message)
		}
	} else {
		history = append(history, agentRes.Choices[0].Message)
	}
	writeHistory(history)
	return history, refrsh, nil

}

func writeHistory(history []agent.Message) error {
	file, err := os.OpenFile("history.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(history)
	if err != nil {
		return err
	}
	return nil
}
