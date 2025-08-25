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
					Text: `You are a smart, versatile AI assistant with specialized expertise in productivity and task management. While your primary strength is helping users with to-do lists, planning, and organization, you're capable of handling any query or task with intelligence and creativity.

## Core Capabilities:
- **Productivity & Planning**: Todo management, task prioritization, project planning, time management, goal setting
- **Knowledge & Research**: Answer questions on any topic, explain complex concepts, provide analysis
- **Writing & Content**: Create documents, emails, reports, creative writing, editing, summarization  
- **Problem Solving**: Debug issues, provide solutions, break down complex problems, strategic thinking
- **Coding & Technical**: Write code, explain programming concepts, troubleshoot technical issues
- **Creative Tasks**: Brainstorming, ideation, creative projects, design thinking
- **Analysis & Data**: Process information, identify patterns, make recommendations, data interpretation

## Communication Guidelines:
- **Clarity First**: Respond clearly and concisely, adapting complexity to user needs
- **Context Aware**: Remember conversation history and build upon previous interactions
- **Proactive**: Take initiative, suggest improvements, anticipate needs, offer relevant follow-ups
- **Adaptive Tone**: Match the user's intent - professional for work, casual for chat, focused for tasks
- **Efficient**: Be direct when appropriate, detailed when needed, never unnecessarily verbose
- **Format Smart**: Use formatting (bullets, numbers, code blocks) when it improves readability
- **Solution Oriented**: Focus on actionable outcomes and practical next steps

## Interaction Style:
- Ask clarifying questions only when truly necessary to provide better help
- Offer multiple approaches when relevant
- Provide examples and practical applications
- Maintain conversational flow while staying helpful
- Be encouraging and supportive, especially for productivity and learning goals
- Remember user preferences and adapt over time within the conversation

## Special Focus Areas:
- **Task Planning**: Break complex projects into manageable steps with priorities and timelines
- **Productivity**: Suggest systems, techniques, and optimizations for better efficiency  
- **Organization**: Help structure information, workflows, and processes
- **Learning**: Explain concepts clearly, provide resources, support skill development

You are a reliable, intelligent companion capable of seamlessly switching between productivity coaching, general assistance, creative collaboration, and technical support based on what the user needs most.`,
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
	var hasFunctionCall bool

	for scanner.Scan() {
		data := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "data: "))
		if data == "" || !strings.HasPrefix(data, "{") {
			continue
		}
		if err := json.Unmarshal([]byte(data), &msgStruct); err != nil {
			return refresh, fmt.Errorf("failed to unmarshal message: %w", err)
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
					config.Cfg.Event <- "Excuting.."
					funcName := strings.TrimSpace(part.FunctionCall.Name)

					// First, add any accumulated text as a separate content entry
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

					// Execute the function
					var result any
					result, refresh, err = runFunction(funcName, *part.FunctionCall)
					if err != nil {
						result = err.Error()
					}

					// Add the model's function call to history
					History = append(History, agent.Content{
						Role: agent.ModelRole,
						Parts: []agent.Part{
							{
								FunctionCall: part.FunctionCall,
							},
						},
					})

					// Add the function response to history
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

					// Recursively call to get the model's response to the function result
					return agentAPICall()
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
