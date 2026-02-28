package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/biisal/godo/internal/builder"
	"github.com/biisal/godo/internal/bus"
	"github.com/biisal/godo/internal/config"
	agentModel "github.com/biisal/godo/internal/tui/models/agent"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Bot struct {
	History      []agentModel.Message
	oaiMessages  []openai.ChatCompletionMessageParamUnion
	client       *openai.Client
	tools        []openai.ChatCompletionToolParam
	systemPrompt string
	ModelName    string
}

func NewBot() *Bot {
	cb := builder.NewContextBuilder()
	c := openai.NewClient(
		option.WithAPIKey(config.Cfg.OPENAI_API_KEY),
		option.WithBaseURL(config.Cfg.OPENAI_BASE_URL),
	)
	return &Bot{
		tools:        FormattedFunctions(),
		systemPrompt: cb.BuildSystemPrompt(),
		client:       &c,
	}
}

func (b *Bot) GetChatHistoryFromDB() (*[]agentModel.Message, error) {
	sqlStmt := "SELECT chat FROM chats"
	rows, err := config.Cfg.DB.Query(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("failed to query chats: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("error closing rows", "err", err)
		}
	}()

	var history []agentModel.Message
	for rows.Next() {
		var chatContent []byte
		if err := rows.Scan(&chatContent); err != nil {
			return nil, fmt.Errorf("failed to scan chat: %w", err)
		}
		msg := agentModel.Message{}
		if err := json.Unmarshal(chatContent, &msg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal chat: %w", err)
		}
		history = append(history, msg)
	}
	return &history, nil
}

func (b *Bot) AddChatToDB(msg agentModel.Message) error {
	sqlStmt := "INSERT INTO chats (chat) VALUES (?)"
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = config.Cfg.DB.Exec(sqlStmt, string(msgJSON))
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

func toOAIMessage(m agentModel.Message) openai.ChatCompletionMessageParamUnion {
	switch m.Role {
	case agentModel.UserRole:
		return openai.UserMessage(m.Content)
	case agentModel.AssistantRole:
		if len(m.ToolCalls) > 0 {
			calls := make([]openai.ChatCompletionMessageToolCallParam, 0, len(m.ToolCalls))
			for _, tc := range m.ToolCalls {
				args := tc.Function.Arguments
				if args == "" || !json.Valid([]byte(args)) {
					args = "{}"
				}
				calls = append(calls, openai.ChatCompletionMessageToolCallParam{
					ID:   tc.ID,
					Type: "function",
					Function: openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      tc.Function.Name,
						Arguments: args,
					},
				})
			}
			asst := openai.ChatCompletionAssistantMessageParam{ToolCalls: calls}
			return openai.ChatCompletionMessageParamUnion{OfAssistant: &asst}
		}
		return openai.AssistantMessage(m.Content)
	case agentModel.ToolRole:
		return openai.ToolMessage(m.Content, m.ToolCallID)
	default:
		return openai.UserMessage(m.Content)
	}
}

func (b *Bot) appendMessage(msg agentModel.Message) {
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

const maxToolSteps = 200

func deltaReasoning(delta openai.ChatCompletionChunkChoiceDelta) string {
	if delta.JSON.ExtraFields == nil {
		return ""
	}
	f, ok := delta.JSON.ExtraFields["reasoning"]
	if !ok {
		return ""
	}
	var r string
	if err := json.Unmarshal([]byte(f.Raw()), &r); err != nil {
		slog.Error("failed to unmarshal reasoning", "err", err)
	}
	return r
}

func (b *Bot) agentAPICall(refresh ...bool) (bool, error) {
	isRefresh := false
	if len(refresh) > 0 {
		isRefresh = refresh[0]
	}
	b.initOAIMessages()

	for range maxToolSteps {
		ctx, cancel := context.WithCancel(context.Background())

		stream := b.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Model:    config.Cfg.OPENAI_MODEL,
			Messages: b.oaiMessages,
			Tools:    b.tools,
		}, option.WithJSONSet("think", true))

		acc := openai.ChatCompletionAccumulator{}
		bus.EmitState(agentModel.StateThinking)
		var reasoningBuilder strings.Builder
		thiking := true

		thinkStartTime := time.Now()
		for stream.Next() {
			chunk := stream.Current()
			raw, _ := json.Marshal(chunk)
			slog.Debug("Raw chunk received", "json", string(raw))
			acc.AddChunk(chunk)

			for _, choice := range chunk.Choices {
				if r := deltaReasoning(choice.Delta); r != "" {
					reasoningBuilder.WriteString(r)
					bus.EmitThinking(r)
					if choice.Delta.Content == "" {
						continue
					}
				}
				if thiking {
					bus.EmitState(agentModel.StateReady)
					bus.EmitMessageStatus(fmt.Sprintf("\nThought for %.1fs", time.Since(thinkStartTime).Seconds()))
					thiking = false
				}
				bus.EmitContent(choice.Delta.Content)
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
			assistantMsg := agentModel.Message{
				Role:      agentModel.AssistantRole,
				ToolCalls: toolCalls,
			}
			if txt := strings.TrimSpace(choice.Message.Content); txt != "" {
				assistantMsg.Content = txt
				bus.EmitContent("")
			}
			assistantMsg.Reasoning = reasoningBuilder.String()
			b.appendMessage(assistantMsg)

			for _, tc := range toolCalls {
				bus.EmitToolCall(tc.Function.Name)
				slog.Info("\n\nRunning tool----------------------------", "name", tc.Function.Name, "args", tc.Function.Arguments)
				result, shouldRefresh, err := runFunction(tc.Function.Name, tc)
				slog.Info("\n\nTool result----------------------------", "result", result, "shouldRefresh", shouldRefresh, "err", err)
				isRefresh = isRefresh || shouldRefresh
				var resultStr string
				if err != nil {
					resultStr = err.Error()
				} else {
					if text, ok := result.(string); ok {
						resultStr = text
					} else {
						b, _ := json.Marshal(result)
						resultStr = string(b)
					}
				}

				toolMsg := agentModel.Message{
					Role:       agentModel.ToolRole,
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
			msg := agentModel.Message{
				Role:      agentModel.AssistantRole,
				Reasoning: reasoningBuilder.String(),
				Content:   finalText,
			}
			b.appendMessage(msg)
			if err := b.AddChatToDB(msg); err != nil {
				slog.Error("error saving chat to db", "err", err)
			}
		}

		return isRefresh, nil
	}

	return isRefresh, fmt.Errorf("agent exceeded max tool steps (%d)", maxToolSteps)
}

func (b *Bot) AgentResponse(prompt string) ([]agentModel.Message, bool, error) {
	if prompt == "" {
		return nil, false, fmt.Errorf("empty message not allowd")
	}
	userMsg := agentModel.Message{
		Role:    agentModel.UserRole,
		Content: prompt,
	}
	b.History = append(b.History, userMsg)
	if err := b.AddChatToDB(userMsg); err != nil {
		slog.Error("error saving chat to db", "err", err)
	}

	bus.EmitStreamStart()
	bus.EmitState(agentModel.StateProcessing)

	defer bus.EmitState(agentModel.StateIdle)
	refresh, err := b.agentAPICall()
	if err != nil {
		return nil, refresh, err
	}
	bus.EmitStreamEnd()
	return b.History, refresh, nil
}

func runFunction(funcName string, tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	if fn, ok := tools[funcName]; ok {
		return fn(tc)
	}
	return nil, false, fmt.Errorf("unknown function: %s", funcName)
}
