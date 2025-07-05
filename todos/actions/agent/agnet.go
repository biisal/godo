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
	"github.com/mitchellh/mapstructure"
)

var (
	AddTodoFunc      = "AddTodo"
	GetTodosInfoFunc = "GetTodosInfo"
	GetTodosFunc     = "GetTodos"
	GetTodoBYIDFunc  = "GetTodoByID"
	DeleteTodoFunc   = "DeleteTodo"
	ModifyTodoFunc   = "ModifyTodo"
	ToggleDoneFunc   = "ToggleDone"
	History          = make([]agent.Content, 0)
)

func getFuncFormatted(name, description string, properties map[string]agent.Property, required []string) agent.FunctionDeclaration {
	return agent.FunctionDeclaration{
		Name:        name,
		Description: description,
		Parameters: agent.Parameter{
			Type:       "object",
			Required:   required,
			Properties: properties,
		},
	}
}

func agentAPICall() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	refresh := false
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", config.Cfg.GEMINI_MODEL, config.Cfg.GEMINI_API_KEY)
	body := agent.AgentReq{
		Contents: History,
		Tools:    FormattedFunctions(),
		SystemInstruction: &agent.Content{
			Role: agent.SystemRole,
			Parts: []agent.Part{
				{
					Text: `You are a smart, autonomous productivity agent. Your primary role is to assist the user with to-do list management, general task planning, and any other queries they may have. You are not limited to to-dos—you can plan, prioritize, brainstorm, and take initiative as needed.

Guidelines:
- Respond clearly and concisely using plain text, without formatting.
- For to-do tasks, suggest categories, priorities, and step-by-step breakdowns when useful.
- Use bullet points or numbered steps only where appropriate for readability.
- Always adapt your tone based on the user's intent—be friendly, focused, or casual as needed.
- Take initiative: propose ideas, organize tasks, and follow up on earlier actions.
- If a request extends beyond task planning (e.g., writing, searching, calculating), handle it smoothly.
- Ask clarifying questions only when absolutely necessary.
- Do not repeat instructions or over-explain yourself.
- Maintain session context and help the user stay organized without being intrusive.

You are a reliable agent capable of adapting to user needs and driving task flow forward efficiently`,
				},
			},
		},
	}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return refresh, err
	}
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyJson))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return refresh, err
	}
	if resp.StatusCode != 200 {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		config.WriteLog(false, string(body))
		return refresh, fmt.Errorf("unexpected status code: %d, reason: %s", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()
	var msgStruct agent.AgentRes

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, len(buf))

	var fullMsg string
	for scanner.Scan() {
		data := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "data: "))
		if data == "" || !strings.HasPrefix(data, "{") {
			continue
		}
		if err := json.Unmarshal([]byte(data), &msgStruct); err != nil {
			return refresh, fmt.Errorf("failed to unmarshal message: %w", err)
		}
		lastIndex := max(len(History)-1, 0)
		for _, candidate := range msgStruct.Candidates {
			content := candidate.Content
			if lastIndex == 0 || History[lastIndex].Role != agent.ModelRole {
				History = append(History, content)
				lastIndex = len(History) - 1
			}
			for _, part := range content.Parts {
				fullMsg += part.Text
				if part.FunctionCall != nil {
					config.Cfg.Event <- "Excuting.."
					funcName := strings.TrimSpace(part.FunctionCall.Name)
					var result any
					result, refresh, err = runFunction(funcName, *part.FunctionCall)
					if err != nil {
						result = err.Error()
					}
					History = append(History, []agent.Content{
						{
							Role: "model",
							Parts: []agent.Part{
								{
									FunctionCall: part.FunctionCall,
								},
							},
						}, {
							Role: agent.FunctionRole,
							Parts: []agent.Part{
								{
									FunctionResponse: &agent.FunctionResponse{
										ID:       part.FunctionCall.ID,
										Name:     "add",
										Response: map[string]any{"output": result},
									},
								},
							},
						}}...,
					)
					agentAPICall()
				}
			}
		}
		if History[lastIndex].Role != agent.ModelRole {
			History = append(History, agent.Content{
				Role: agent.ModelRole,
				Parts: []agent.Part{
					{
						Text: fullMsg,
					},
				},
			})
			lastIndex = len(History) - 1
		} else {
			History[lastIndex].Parts[0].Text = fullMsg
		}
		config.Ping <- ""

	}
	return refresh, nil
}

func AgentResponse(prompt string) ([]agent.Content, bool, error) {
	var refresh = false
	History = append(History, agent.Content{
		Role: agent.UserRole,
		Parts: []agent.Part{
			{
				Text: prompt,
			},
		},
	})
	config.Ping <- ""
	var err error
	refresh, err = agentAPICall()
	if err != nil {
		return nil, refresh, err
	}

	ev := config.Cfg.Event
	ev <- "ᯓ➤ Thinking"
	defer func() {
		ev <- ":)"
	}()

	return History, refresh, nil

}

func runFunction(funcName string, tool agent.FunctionCall) (any, bool, error) {
	var result any
	var err error
	var refresh bool

	switch funcName {
	case AddTodoFunc:
		var args struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}
		if err := mapstructure.Decode(tool.Args, &args); err != nil {
			return "", refresh, fmt.Errorf("invalid JSON in tool call arguments: %w\nraw: %s", err, tool.Args)
		}
		_, err := todo.AddTodo(args.Title, args.Description)
		if err != nil {
			return "", refresh, err
		}
		refresh = true
		result = fmt.Sprintf("Added todo with title: %s and description: %s", args.Title, args.Description)

	case GetTodosInfoFunc:
		total, completed, remains, err := todo.GetTodosInfo()
		result = fmt.Sprintf("Total : %d, Completed : %d, Remaining : %d", total, completed, remains)
		if err != nil {
			return "", refresh, err
		}
	case GetTodosFunc:
		todos, err := todo.GetTodos()
		if err != nil {
			return "", refresh, err
		}
		return todos, refresh, nil
	case ToggleDoneFunc:
		var args struct {
			Id   int  `json:"id"`
			Done bool `json:"done"`
		}
		if err := mapstructure.Decode(tool.Args, &args); err != nil {
			return "", refresh, fmt.Errorf("invalid JSON in tool call arguments: %w\nraw: %s", err, tool.Args)
		}
		isDone, err := todo.ToggleDone(args.Id-1, args.Done)
		if err != nil {
			return "", refresh, err
		}
		refresh = true
		result = fmt.Sprintf("Completed Status set to : %t", isDone)
	case GetTodoBYIDFunc:
		var arg struct {
			Id int `json:"id"`
		}
		if err := mapstructure.Decode(tool.Args, &arg); err != nil {
			return "", refresh, fmt.Errorf("invalid JSON in tool call arguments: %w\nraw: %s", err, tool.Args)
		}
		result, err = todo.GetTodoById(arg.Id + 1)
		if err != nil {
			return "", refresh, err
		}

	case DeleteTodoFunc:
		var arg struct {
			Id int `json:"id"`
		}
		if err := mapstructure.Decode(tool.Args, &arg); err != nil {
			return "", refresh, fmt.Errorf("invalid JSON in tool call arguments: %w\nraw: %s", err, tool.Args)
		}
		_, err = todo.DeleteTodo(arg.Id + 1)
		if err != nil {
			return "", refresh, err
		}
		result = fmt.Sprintf("Deleted todo with id : %d", arg.Id)
		refresh = true
	default:
		return "", refresh, fmt.Errorf("Unknown function %s", funcName)

	}
	return result, refresh, nil

}
