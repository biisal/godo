package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/biisal/todo-cli/config"
	"github.com/biisal/todo-cli/todos/actions/todo"
	"github.com/biisal/todo-cli/todos/models/agent"
)

var (
	AddTodoFunc      = "AddTodo"
	GetTodosInfoFunc = "GetTodosInfo"
	DeleteTodoFunc   = "DeleteTodo"
	ModifyTodoFunc   = "ModifyTodo"
	ToggleDoneFunc   = "ToggleDone"
)

func getFuncFormatted(toolType, name, description string, properties map[string]agent.PropertyType, required []string) agent.Tool {
	return agent.Tool{
		Type: toolType,
		Function: agent.FunctionReq{
			Name:        name,
			Description: description,
			Parameters: agent.Parameters{
				Type:       "object",
				Properties: properties,
				Required:   required,
			},
		},
	}

}

func agentAPICall(history []agent.Message) (*agent.AgentRes, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	history = append([]agent.Message{
		{
			Role: "system",
			Content: `You are a helpful and goal-oriented agent. Your primary role is to assist users in staying consistent with their goals.

You have access to tools and may use them when necessary — but you must use them wisely and only when they are genuinely needed to provide accurate information.
You must summarize any tool outputs clearly and return only the relevant results to the user. Do not mention tool usage or reveal that a tool was invoked.
Focus on providing concise, actionable, and helpful responses that guide the user effectively toward their goals.`,
		},
	}, history...)
	url := "https://api.groq.com/openai/v1/chat/completions"
	body := agent.AgentReq{
		Messages: history,
		Model:    config.Cfg.GROQ_MODEL,
		TopP:     1,
		Stream:   false,
		Tools: []agent.Tool{
			getFuncFormatted("function", GetTodosInfoFunc,
				"Returns the todos info [total , completed , remains]",
				map[string]agent.PropertyType{},
				make([]string,
					0)),
			getFuncFormatted("function", AddTodoFunc, "Use to add todo in the todos list", map[string]agent.PropertyType{
				"title": {
					Type:        "string",
					Description: "The title of the todo",
				},
				"description": {
					Type:        "string",
					Description: "The description of the todo",
				},
			}, []string{"title", "description"},
			),
		},
		ToolChoice:          "auto",
		MaxCompletionTokens: 900,
	}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Cfg.GROQ_API_KEY)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	writeHistory(string(data))
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got status code %d\nResponse: %s", resp.StatusCode, data)
	}
	var msgStruct agent.AgentRes
	if err = json.Unmarshal(data, &msgStruct); err != nil {
		return nil, err
	}
	return &msgStruct, nil
}

func AgentResponse(history []agent.Message) ([]agent.Message, bool, error) {
	agentRes, err := agentAPICall(history)
	ev := config.Cfg.Event
	ev <- "ᯓ➤ Thinking"
	defer func() {
		ev <- ":)"
	}()
	refrsh := false
	if err != nil {
		return nil, refrsh, err
	}
	if len(agentRes.Choices) == 0 {
		return history, refrsh, nil
	}
	history = append(history, agentRes.Choices[0].Message)
	tools := agentRes.Choices[0].Message.ToolCalls
	if len(tools) > 0 {
		config.Cfg.Event <- "Working..."
		funcName := strings.TrimSpace(tools[0].Function.Name)
		switch funcName {
		case AddTodoFunc:
			var args struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			}
			if err = json.Unmarshal([]byte(tools[0].Function.Arguments), &args); err != nil {
				return nil, refrsh, fmt.Errorf("invalid JSON in tool call arguments: %w\nraw: %s", err, tools[0].Function.Arguments)
			}
			_, err := todo.AddTodo(args.Title, args.Description)
			if err != nil {
				return nil, refrsh, err
			}
			refrsh = true
			resultStr := fmt.Sprintf("Added todo with title: %s and description: %s", args.Title, args.Description)
			history = append(history,
				agent.Message{
					Role:       agent.ToolRole,
					Content:    resultStr,
					ToolCallId: tools[0].Id,
					Name:       AddTodoFunc,
				},
			)
			agentRes, err = agentAPICall(history)
			if err != nil {
				return nil, refrsh, err
			}
			history = append(history, agentRes.Choices[0].Message)
		case GetTodosInfoFunc:
			result, err := todo.GetTodosInfo()
			if err != nil {
				return nil, refrsh, err
			}
			history = append(history,
				agent.Message{
					Role:       agent.ToolRole,
					Content:    result,
					ToolCallId: tools[0].Id,
					Name:       GetTodosInfoFunc,
				},
			)
			agentRes, err = agentAPICall(history)
			if err != nil {
				return nil, refrsh, err
			}
			history = append(history, agentRes.Choices[0].Message)
		}
	}
	return history, refrsh, nil

}

func writeHistory(data any) error {
	file, err := os.OpenFile("history.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	dataMap := map[string]any{}
	err = json.Unmarshal([]byte(data.(string)), &dataMap)
	if err != nil {
		return err
	}
	err = json.NewEncoder(file).Encode(dataMap)
	if err != nil {
		return err
	}
	return nil
}
