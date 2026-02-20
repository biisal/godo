package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/biisal/godo/bus"
	"github.com/biisal/godo/config"
	"github.com/biisal/godo/identity"
	"github.com/biisal/godo/tui/actions/todo"
	"github.com/biisal/godo/tui/models/agent"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Bot struct {
	History      []agent.Message
	oaiMessages  []openai.ChatCompletionMessageParamUnion
	client       *openai.Client
	tools        []openai.ChatCompletionToolParam
	systemPrompt string
}

func NewBot() *Bot {
	builder := identity.NewContextBuilder()
	c := openai.NewClient(
		option.WithAPIKey(config.Cfg.OPENAI_API_KEY),
		option.WithBaseURL(config.Cfg.OPENAI_BASE_URL),
	)
	return &Bot{
		tools:        FormattedFunctions(),
		systemPrompt: builder.BuildSystemPrompt(),
		client:       &c,
	}
}

func (b *Bot) GetChatHistoryFromDB() (*[]agent.Message, error) {
	sqlStmt := "SELECT chat FROM chats"
	rows, err := config.Cfg.DB.Query(sqlStmt)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	var history []agent.Message
	for rows.Next() {
		var chatContent []byte
		if err := rows.Scan(&chatContent); err != nil {
			fmt.Println(err)
		}
		msg := agent.Message{}
		if err := json.Unmarshal(chatContent, &msg); err != nil {
			fmt.Println(err)
		}
		history = append(history, msg)
	}
	return &history, nil
}

func (b *Bot) AddChatToDB(msg agent.Message) error {
	sqlStmt := "INSERT INTO chats (chat) VALUES (?)"
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = config.Cfg.DB.Exec(sqlStmt, string(msgJson))
	return err
}

func (b *Bot) TruncateChats() error {
	sqlStmt := "DELETE FROM chats"
	_, err := config.Cfg.DB.Exec(sqlStmt)
	if err == nil {
		b.History = nil
		b.oaiMessages = nil
	}
	return err
}

func toOAIMessage(m agent.Message) openai.ChatCompletionMessageParamUnion {
	switch m.Role {
	case agent.UserRole:
		return openai.UserMessage(m.Content)
	case agent.AssistantRole:
		if len(m.ToolCalls) > 0 {
			calls := make([]openai.ChatCompletionMessageToolCallParam, 0, len(m.ToolCalls))
			for _, tc := range m.ToolCalls {
				calls = append(calls, openai.ChatCompletionMessageToolCallParam{
					ID:   tc.ID,
					Type: "function",
					Function: openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				})
			}
			asst := openai.ChatCompletionAssistantMessageParam{ToolCalls: calls}
			return openai.ChatCompletionMessageParamUnion{OfAssistant: &asst}
		}
		return openai.AssistantMessage(m.Content)
	case agent.ToolRole:
		return openai.ToolMessage(m.Content, m.ToolCallID)
	default:
		return openai.UserMessage(m.Content)
	}
}

func (b *Bot) appendMessage(msg agent.Message) {
	b.History = append(b.History, msg)
	b.oaiMessages = append(b.oaiMessages, toOAIMessage(msg))
}

func (b *Bot) initOAIMessages() {
	b.oaiMessages = make([]openai.ChatCompletionMessageParamUnion, 0, len(b.History)+1)
	b.oaiMessages = append(b.oaiMessages, openai.SystemMessage(b.systemPrompt))
	for _, m := range b.History {
		b.oaiMessages = append(b.oaiMessages, toOAIMessage(m))
	}
}

const maxToolSteps = 5

func deltaReasoning(delta openai.ChatCompletionChunkChoiceDelta) string {
	if delta.JSON.ExtraFields == nil {
		return ""
	}
	f, ok := delta.JSON.ExtraFields["reasoning"]
	if !ok {
		return ""
	}
	var r string
	json.Unmarshal([]byte(f.Raw()), &r)
	return r
}

func (b *Bot) agentAPICall(refresh ...bool) (bool, error) {
	isRefresh := false
	if len(refresh) > 0 {
		isRefresh = refresh[0]
	}

	slog.Info("Running initOAIMessages", "time", time.Now())
	b.initOAIMessages()
	slog.Info("Finished initOAIMessages", "time", time.Now(), "time taken", time.Since(time.Now()))

	for range maxToolSteps {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

		stream := b.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Model:    config.Cfg.OPENAI_MODEL,
			Messages: b.oaiMessages,
			Tools:    b.tools,
		}, option.WithJSONSet("think", true))

		acc := openai.ChatCompletionAccumulator{}
		bus.EmitStatus("Thinking...")
		var reasoningBuilder strings.Builder

		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			for _, choice := range chunk.Choices {
				if r := deltaReasoning(choice.Delta); r != "" {
					reasoningBuilder.WriteString(r)
					bus.EmitThinking(r)
				}
				if content := choice.Delta.Content; content != "" {
					bus.EmitContent(content)
				}
			}
		}
		cancel()

		if err := stream.Err(); err != nil {
			return isRefresh, fmt.Errorf("stream error: %w", err)
		}

		if len(acc.Choices) == 0 {
			return isRefresh, nil
		}

		choice := acc.Choices[0]
		toolCalls := choice.Message.ToolCalls

		if len(toolCalls) > 0 {
			assistantMsg := agent.Message{
				Role:      agent.AssistantRole,
				ToolCalls: toolCalls,
			}
			if txt := strings.TrimSpace(choice.Message.Content); txt != "" {
				assistantMsg.Content = txt
				bus.Emit("")
			}
			assistantMsg.Reasoning = reasoningBuilder.String()
			b.appendMessage(assistantMsg)

			for _, tc := range toolCalls {
				bus.EmitToolCall(tc.Function.Name)
				// slog.Info("\n\nRunning tool----------------------------", "name", tc.Function.Name, "args", tc.Function.Arguments)
				result, shouldRefresh, err := runFunction(tc.Function.Name, tc)
				// slog.Info("\n\nTool result----------------------------", "result", result, "shouldRefresh", shouldRefresh, "err", err)
				isRefresh = isRefresh || shouldRefresh
				var resultStr string
				if err != nil {
					resultStr = err.Error()
				} else {
					b, _ := json.Marshal(result)
					resultStr = string(b)
				}

				toolMsg := agent.Message{
					Role:       agent.ToolRole,
					ToolCallID: tc.ID,
					Name:       tc.Function.Name,
					Content:    resultStr,
				}
				b.appendMessage(toolMsg)
			}
			continue
		}

		finalText := choice.Message.Content
		if finalText != "" || reasoningBuilder.Len() > 0 {
			msg := agent.Message{
				Role:      agent.AssistantRole,
				Reasoning: reasoningBuilder.String(),
				Content:   finalText,
			}
			b.appendMessage(msg)
			b.AddChatToDB(msg)
		}

		return isRefresh, nil
	}

	return isRefresh, fmt.Errorf("agent exceeded max tool steps (%d)", maxToolSteps)
}

func (b *Bot) AgentResponse(prompt string) ([]agent.Message, bool, error) {
	userMsg := agent.Message{
		Role:    agent.UserRole,
		Content: prompt,
	}
	b.History = append(b.History, userMsg)
	b.AddChatToDB(userMsg)

	bus.Emit("START")
	refresh, err := b.agentAPICall()
	if err != nil {
		return nil, refresh, err
	}
	bus.Emit("DONE")
	return b.History, refresh, nil
}

func runFunction(funcName string, tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	switch funcName {
	case PerformSqlFunc:
		var args struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return "", false, fmt.Errorf("invalid tool arguments: %w", err)
		}
		result, err := todo.PerformSqlQuery(args.Query)
		if err != nil {
			return "", false, err
		}
		return result, true, nil

	case RunShellCommandFunc:
		var args struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return "", false, fmt.Errorf("invalid tool arguments: %w", err)
		}
		cmd := exec.Command("sh", "-c", args.Command)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return string(out) + "\nError: " + err.Error(), false, nil
		}
		slog.Debug("command output", "output", string(out))
		return string(out), false, nil

	case ReadSkillFunc:
		var args struct {
			SkillName string `json:"skillName"`
		}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return "", false, fmt.Errorf("invalid tool arguments: %w", err)
		}
		baseDir, _ := os.Getwd()
		skillPath := filepath.Join(baseDir, "skills", args.SkillName+".md")
		content, err := os.ReadFile(skillPath)
		if err != nil {
			return "", false, fmt.Errorf("failed to read skill %s: %w", args.SkillName, err)
		}
		return string(content), false, nil

	default:
		return "", false, fmt.Errorf("unknown function: %s", funcName)
	}
}
