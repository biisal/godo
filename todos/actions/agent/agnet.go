package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
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

func sendDummyResponse() {
	text := `Hello there! I’m an AI assistant designed to help you with a wide variety of tasks. Whether you need assistance with writing, coding, learning new concepts, brainstorming ideas, or just having a conversation, I’m here to support you.

I can generate text, explain complex topics in simple terms, provide step-by-step guidance, and even help you plan projects or solve problems. For example, if you’re working on a programming project, I can suggest code snippets, debug errors, or explain how certain algorithms work. If you’re learning a new subject, I can break it down into manageable parts and give examples to make it easier to understand.

I’m also capable of creative tasks—like drafting stories, creating dialogues, or imagining scenarios—and analytical tasks, like summarizing information, comparing options, or generating structured plans. My main goal is to be helpful, informative, and easy to understand, while adapting to your style and preferences.

So, whether you have a specific question, want to explore ideas, or just want to experiment with AI, feel free to ask me anything. I’ll provide responses that are clear, thorough, and tailored to your needs. Let’s make your tasks easier and more fun together!`
	for word := range strings.SplitSeq(text, " ") {
		config.StreamResponse <- config.StreamMsg{Text: word + " "}
	}

}
func GetChatHistoryFromDB() (*[]agent.Content, error) {
	sqlStmt := "SELECT chat FROM chats"
	rows, err := config.Cfg.DB.Query(sqlStmt)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	var history []agent.Content
	for rows.Next() {
		var chatContent []byte
		if err := rows.Scan(&chatContent); err != nil {
			fmt.Println(err)
		}
		chat := agent.Content{}
		if err := json.Unmarshal(chatContent, &chat); err != nil {
			fmt.Println(err)
		}
		history = append(history, chat)
	}
	return &history, nil

}
func AddChatToDB(content agent.Content) error {
	sqlStmt := "INSERT INTO chats (chat) VALUES (?)"
	contentJson, err := json.Marshal(content)
	if err != nil {
		return err
	}
	_, err = config.Cfg.DB.Exec(sqlStmt, string(contentJson))
	if err != nil {
		return err
	}
	return nil
}
func TruncateChats() error {
	// sqlStmt := "TRUNCATE TABLE chats"
	sqlStmt := "DELETE FROM chats"
	_, err := config.Cfg.DB.Exec(sqlStmt)
	if err == nil {
		History = nil
	}
	return err

}
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

// AI generated helper function
func stripMarkdown(md string) string {
	// 1. Remove links: [text](url) → keep "text"
	re := regexp.MustCompile(`\[(.*?)\]\(.*?\)`)
	text := re.ReplaceAllString(md, "$1")

	// 2. Remove images: ![alt](url) → keep "alt"
	re = regexp.MustCompile(`!\[(.*?)\]\(.*?\)`)
	text = re.ReplaceAllString(text, "$1")

	// 3. Remove bold/italic (**bold**, *italic*, ***both***)
	re = regexp.MustCompile(`\*{1,3}([^\*]+)\*{1,3}`)
	text = re.ReplaceAllString(text, "$1")

	// 4. Remove inline code `code`
	re = regexp.MustCompile("`([^`]+)`")
	text = re.ReplaceAllString(text, "$1")

	// 5. Remove headings (### Title → Title)
	re = regexp.MustCompile(`(?m)^#{1,6}\s*`)
	text = re.ReplaceAllString(text, "")

	// 6. Remove list markers (-, *, 1.)
	re = regexp.MustCompile(`(?m)^\s*[-*+]\s+`)
	text = re.ReplaceAllString(text, "")
	re = regexp.MustCompile(`(?m)^\s*\d+\.\s+`)
	text = re.ReplaceAllString(text, "")

	return text
}

func agentAPICall(logger *logger.Logger, refresh ...bool) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	isRefresh := false
	if len(refresh) > 0 {
		isRefresh = refresh[0]
	}
	currentDateTime := time.Now().Format("2006-01-02 15:04:05")
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", config.Cfg.GEMINI_MODEL, config.Cfg.GEMINI_API_KEY)
	body := agent.AgentReq{
		Contents: History,
		Tools:    FormattedFunctions(),
		SystemInstruction: &agent.Content{
			Role: agent.SystemRole,
			Parts: []agent.Part{
				{
					Text: `
You are the AI assistant inside the GoDo CLI app.  
You help with productivity, task management, coding, troubleshooting, writing, and learning.

Abilities:
- Manage and organize tasks
- Break down projects into steps
- Explain concepts and answer questions
- Debug and solve technical problems
- Summarize, write, and edit text
- Suggest ways to boost efficiency

Style:
- Be clear, concise, and practical
- Keep context across inputs
- Match tone: professional for work, casual for chat

You always have access to the current time: ` + currentDateTime,
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
					for word := range strings.SplitSeq(part.Text, " ") {
						config.StreamResponse <- config.StreamMsg{Text: word + " "}
						time.Sleep(20 * time.Millisecond)
					}

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
						config.StreamResponse <- config.StreamMsg{Text: ""}
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

					return agentAPICall(logger, isRefresh)
				}
			}
		}
	}

	// Add any remaining text content after processing all chunks
	if fullMsg != "" && !hasFunctionCall {
		History = append(History, agent.Content{
			Role: agent.ModelRole,
			Parts: []agent.Part{
				{
					Text: stripMarkdown(fullMsg),
				},
			},
		})
	}
	// config.StreamResponse <- fullMsg
	AddChatToDB(History[len(History)-1])
	return isRefresh, nil
}

func AgentResponse(prompt string, logger *logger.Logger) ([]agent.Content, bool, error) {
	var refresh = false
	var userInput = agent.Content{
		Role: agent.UserRole,
		Parts: []agent.Part{
			{
				Text: prompt,
			},
		},
	}

	History = append(History, userInput)
	AddChatToDB(userInput)
	var err error
	config.StreamResponse <- config.StreamMsg{Text: "START"}
	logger.Debug("START SHOULD BE IN UI")
	refresh, err = agentAPICall(logger)
	// sendDummyResponse()
	if err != nil {
		return nil, refresh, err
	}
	config.StreamResponse <- config.StreamMsg{Text: "DONE"}
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
