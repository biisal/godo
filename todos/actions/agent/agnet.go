package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/biisal/todo-cli/config"
	"github.com/biisal/todo-cli/todos/actions/todo"
	"github.com/biisal/todo-cli/todos/models/agent"
)

var (
	AddTodoFunc      = "AddTodo"
	GetTodosInfoFunc = "GetTodosInfo"
	GetTodosFunc     = "GetTodos"
	GetTodoBYIDFunc  = "GetTodoByID"
	DeleteTodoFunc   = "DeleteTodo"
	ModifyTodoFunc   = "ModifyTodo"
	ToggleDoneFunc   = "ToggleDone"
	History          = make([]agent.Message, 0)
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

func agentAPICall() ([]agent.ToolCall, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	stream := true
	url := "https://api.groq.com/openai/v1/chat/completions"
	body := agent.AgentReq{
		Messages: append([]agent.Message{
			{
				Role: "system",
				Content: `You are a helpful and goal-oriented agent. Your primary role is to assist users in staying consistent with their goals.

You have access to tools and may use them when necessary — but you must use them wisely and only when they are genuinely needed to provide accurate information.
You must summarize any tool outputs clearly and return only the relevant results to the user. Do not mention tool usage or reveal that a tool was invoked.
Focus on providing concise, actionable, and helpful responses that guide the user effectively toward their goals.`,
			},
		}, History...),
		Model:  config.Cfg.GROQ_MODEL,
		TopP:   1,
		Stream: stream,
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
			getFuncFormatted("function", GetTodosFunc,
				"Get all todos with ID, title, description and Completed status",
				map[string]agent.PropertyType{},
				make([]string, 0)),
			getFuncFormatted("function", ToggleDoneFunc, "Use to update the done staus of the todo with fixed value or toggle it", map[string]agent.PropertyType{
				"id": {
					Type:        "number",
					Description: "The id of the todo",
				},
				"done": {
					Type:        "boolean",
					Description: "The new done status of the todo",
				},
			}, []string{"id"}),
			getFuncFormatted("function", DeleteTodoFunc, "Delete todo by ID", map[string]agent.PropertyType{
				"id": {
					Type:        "number",
					Description: "The id of the todo",
				},
			}, []string{"id"}),
			getFuncFormatted("function", GetTodoBYIDFunc, "Get todo by ID", map[string]agent.PropertyType{
				"id": {
					Type:        "number",
					Description: "The id of the todo",
				}}, []string{"id"}),
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
	if resp.StatusCode != 200 {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		return nil, fmt.Errorf("unexpected status code: %d, reason: %s", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()
	var toolCalls []agent.ToolCall
	var msgStruct agent.AgentRes
	var fullMsg string
	var currentMessageIndex int = -1

	if stream {
		scanner := bufio.NewScanner(resp.Body)
		buf := make([]byte, 1024*1024)
		scanner.Buffer(buf, len(buf))
		for scanner.Scan() {
			data := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "data: "))
			if data == "[DATA]" {
				break
			}
			if data == "" || !strings.HasPrefix(data, "{") {
				continue
			}
			if err := json.Unmarshal([]byte(data), &msgStruct); err != nil {
				return nil, err

			}
			if len(msgStruct.Choices) == 0 {
				continue
			}
			delta := msgStruct.Choices[0].Delta
			if len(delta.ToolCalls) > 0 {
				toolCalls = append(toolCalls, delta.ToolCalls...)
				break
			}
			content := delta.Content
			if content == "" {
				content = delta.Reasoning
			}
			if content != "" {
				fullMsg += content
				if currentMessageIndex == -1 {
					History = append(History, agent.Message{
						Role:    agent.AssistantRole,
						Content: content,
					})
					currentMessageIndex = len(History) - 1
				} else {
					History[currentMessageIndex].Content += content
				}
				config.Ping <- ""
			}
		}
	}
	if len(toolCalls) > 0 && currentMessageIndex != -1 {
		History[currentMessageIndex].ToolCalls = toolCalls
		return toolCalls, nil
	}

	return nil, nil
}

func AgentResponse(prompt string) ([]agent.Message, bool, error) {
	var refresh = false
	History = append(History, agent.Message{
		Role:    agent.UserRole,
		Content: prompt,
	})
	config.Ping <- ""
	tools, err := agentAPICall()
	if err != nil {
		return nil, refresh, err
	}

	ev := config.Cfg.Event
	ev <- "ᯓ➤ Thinking"
	defer func() {
		ev <- ":)"
	}()

	for _, tool := range tools {
		config.Cfg.Event <- ":*"
		funcName := strings.TrimSpace(tool.Function.Name)
		var resultStr string
		resultStr, refresh, err = runFunction(funcName, tool)
		if err != nil {
			return nil, refresh, nil
		}
		History = append(History,
			agent.Message{
				Role:       agent.ToolRole,
				Content:    resultStr,
				ToolCallId: tool.Id,
				Name:       AddTodoFunc,
			},
		)
		tools, err = agentAPICall()
		if err != nil {
			return nil, refresh, err
		}
	}
	return History, refresh, nil

}

func runFunction(funcName string, tool agent.ToolCall) (string, bool, error) {
	var resultStr string
	var refresh bool
	switch funcName {
	case AddTodoFunc:
		var args struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal([]byte(tool.Function.Arguments), &args); err != nil {
			return "", refresh, fmt.Errorf("invalid JSON in tool call arguments: %w\nraw: %s", err, tool.Function.Arguments)
		}
		_, err := todo.AddTodo(args.Title, args.Description)
		if err != nil {
			return "", refresh, err
		}
		refresh = true
		resultStr = fmt.Sprintf("Added todo with title: %s and description: %s", args.Title, args.Description)

	case GetTodosInfoFunc:
		total, completed, remains, err := todo.GetTodosInfo()
		resultStr = fmt.Sprintf("Total : %d, Completed : %d, Remaining : %d", total, completed, remains)
		if err != nil {
			return "", refresh, err
		}
	case GetTodosFunc:
		todos, err := todo.GetTodos()
		if err != nil {
			return "", refresh, err
		}
		for _, todo := range todos {
			resultStr += fmt.Sprintf("%d {Title : %s\nDescription %s\n Done :%t}\n\n", todo.ID+1, todo.TitleText, todo.DescriptionText, todo.Done)
		}
	case ToggleDoneFunc:
		var args struct {
			Id   int  `json:"id"`
			Done bool `json:"done"`
		}
		if err := json.Unmarshal([]byte(tool.Function.Arguments), &args); err != nil {
			return "", refresh, fmt.Errorf("invalid JSON in tool call arguments: %w\nraw: %s", err, tool.Function.Arguments)
		}
		isDone, err := todo.ToggleDone(args.Id, args.Done)
		if err != nil {
			return "", refresh, err
		}
		refresh = true
		resultStr = fmt.Sprintf("Completed Status set to : %t", isDone)
	case GetTodoBYIDFunc:
		var arg struct {
			Id int `json:"id"`
		}
		if err := json.Unmarshal([]byte(tool.Function.Arguments), &arg); err != nil {
			return "", refresh, fmt.Errorf("invalid JSON in tool call arguments: %w\nraw: %s", err, tool.Function.Arguments)
		}
		todo, err := todo.GetTodoById(arg.Id + 1)
		if err != nil {
			return "", refresh, err
		}
		resultStr = fmt.Sprintf("%d {Title : %s\nDescription %s\n Done :%t}\n\n", todo.ID, todo.TitleText, todo.DescriptionText, todo.Done)
	default:
		return "", refresh, fmt.Errorf("Unknown function %s", funcName)

	}
	return resultStr, refresh, nil

}

// func writeHistory(data any) error {
// 	file, err := os.OpenFile("history.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()
// 	// dataMap := map[string]any{}
// 	// err = json.Unmarshal([]byte(data.(string)), &dataMap)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	err = json.NewEncoder(file).Encode(data)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
