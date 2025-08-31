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
	"github.com/biisal/todo-cli/logger"
	"github.com/biisal/todo-cli/todos/actions/todo"
	"github.com/biisal/todo-cli/todos/models/agent"
	"github.com/mitchellh/mapstructure"
)

var (
	PerformSqlFunc   = "PerformSql"
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

func agentAPICall(refresh ...bool) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	isRefresh := false
	if len(refresh) > 0 {
		isRefresh = refresh[0]
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", config.Cfg.GEMINI_MODEL, config.Cfg.GEMINI_API_KEY)
	body := agent.AgentReq{
		Contents: History,
		Tools:    FormattedFunctions(),
		SystemInstruction: &agent.Content{
			Role: agent.SystemRole,
			Parts: []agent.Part{
				{
					Text: `
You are a smart, versatile AI assistant with strong skills in productivity and task management. 
Your main focus is helping with to-do lists, planning, organization, and efficiency, but you can also 
handle coding, technical support, research, writing, and creative problem solving.

Core abilities:
- Task management, planning, prioritization, time management
- Answering questions and explaining concepts clearly
- Writing, summarizing, and editing text
- Debugging and troubleshooting technical issues
- Brainstorming and creative support
- Analyzing data and making recommendations

Communication style:
- Be clear and concise, adapting to user needs
- Remember context and build on previous interactions
- Be proactive and solution-focused
- Match tone: professional for work, casual for chat
- Use formatting only when it helps readability
- Always aim for practical and actionable responses

Interaction approach:
- Ask questions only if needed for clarity
- Provide examples and options when useful
- Keep conversation flowing and supportive
- Encourage productivity and learning

Special focus:
- Breaking projects into manageable steps
- Suggesting techniques to boost efficiency
- Structuring workflows and information
- Supporting learning and skill growth

You are a reliable companion for productivity, problem solving, and general assistance in both CLI and formal chat.
`,
				},
			},
		},
	}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return isRefresh, err
	}
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyJson))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return isRefresh, err
	}
	if resp.StatusCode != 200 {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		return isRefresh, fmt.Errorf("unexpected status code: %d, reason: %s", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()
	var msgStruct agent.AgentRes

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, len(buf))

	var fullMsg string
	var hasFunctionCall bool

	for scanner.Scan() {
		data := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "data: "))
		if data == "" || !strings.HasPrefix(data, "{") {
			continue
		}
		if err := json.Unmarshal([]byte(data), &msgStruct); err != nil {
			return isRefresh, fmt.Errorf("failed to unmarshal message: %w", err)
		}

		for _, candidate := range msgStruct.Candidates {
			for _, part := range candidate.Content.Parts {
				// Handle text content
				if part.Text != "" {
					fullMsg += part.Text
				}

				// Handle function calls
				if part.FunctionCall != nil {
					hasFunctionCall = true
					funcName := strings.TrimSpace(part.FunctionCall.Name)

					if fullMsg != "" {
						History = append(History, agent.Content{
							Role: agent.ModelRole,
							Parts: []agent.Part{
								{
									Text: fullMsg,
								},
							},
						})
						fullMsg = "" // Reset after adding
					}

					var result any
					result, isRefresh, err = runFunction(funcName, *part.FunctionCall)
					if err != nil {
						result = err.Error()
					}

					History = append(History, agent.Content{
						Role: agent.ModelRole,
						Parts: []agent.Part{
							{
								FunctionCall: part.FunctionCall,
							},
						},
					})

					History = append(History, agent.Content{
						Role: agent.FunctionRole,
						Parts: []agent.Part{
							{
								FunctionResponse: &agent.FunctionResponse{
									ID:       part.FunctionCall.ID,
									Name:     funcName,
									Response: map[string]any{"output": result},
								},
							},
						},
					})

					return agentAPICall(isRefresh)
				}
			}
		}
		config.Ping <- ""
	}

	// Add any remaining text content after processing all chunks
	if fullMsg != "" && !hasFunctionCall {
		History = append(History, agent.Content{
			Role: agent.ModelRole,
			Parts: []agent.Part{
				{
					Text: fullMsg,
				},
			},
		})
	}
	return isRefresh, nil
}

func AgentResponse(prompt string, logger *logger.Logger) ([]agent.Content, bool, error) {
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
	config.Ping <- ""
	if err != nil {
		return nil, refresh, err
	}

	logger.Info("REFRESH IS", refresh)
	return History, refresh, nil
}

func runFunction(funcName string, tool agent.FunctionCall) (any, bool, error) {
	var result any
	var err error
	var refresh bool

	switch funcName {
	case PerformSqlFunc:
		var args struct {
			Query string `json:"query"`
		}
		if err := mapstructure.Decode(tool.Args, &args); err != nil {
			return "", refresh, fmt.Errorf("invalid JSON in tool call arguments: %w\nraw: %s", err, tool.Args)
		}
		result, err = todo.PerformSqlQuery(args.Query)
		if err != nil {
			return "", refresh, err
		}
		refresh = true
	default:
		return "", refresh, fmt.Errorf("Unknown function %s", funcName)
	}
	return result, refresh, nil
}
