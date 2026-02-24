package agent

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/openai/openai-go"
)

type AgentModel struct {
	PromptInput      textinput.Model
	StreamChunk      strings.Builder
	ChatViewport     viewport.Model
	StatusText       string
	CurrentReasoning strings.Builder
	ShellViewport    viewport.Model
	ShellContent     strings.Builder
}

const (
	UserRole      = "user"
	AssistantRole = "assistant"
	SystemRole    = "system"
	ToolRole      = "tool"
)

type Message struct {
	Role       string                                 `json:"role"`
	Reasoning  string                                 `json:"reasoning,omitempty"`
	Content    string                                 `json:"content,omitempty"`
	ToolCallID string                                 `json:"tool_call_id,omitempty"`
	Name       string                                 `json:"name,omitempty"`
	ToolCalls  []openai.ChatCompletionMessageToolCall `json:"tool_calls,omitempty"`
}
