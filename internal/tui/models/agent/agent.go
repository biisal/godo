package agent

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/openai/openai-go"
)

type AgentModel struct {
	PromptInput   textinput.Model
	ChatViewport  viewport.Model
	StateText     string
	ShellViewport viewport.Model
	ShellContent  strings.Builder
}

const (
	UserRole      = "user"
	AssistantRole = "assistant"
	SystemRole    = "system"
	ToolRole      = "tool"

	StateThinking   = "Thinking..."
	StateProcessing = "Preparing request..."
	StateWriting    = "Generating response..."
	StateReady      = "Responding..."
	StateIdle       = "Ask me anything"
)

type Message struct {
	Role       string                                 `json:"role"`
	Reasoning  string                                 `json:"reasoning,omitempty"`
	Content    string                                 `json:"content,omitempty"`
	ToolCallID string                                 `json:"tool_call_id,omitempty"`
	Name       string                                 `json:"name,omitempty"`
	ToolCalls  []openai.ChatCompletionMessageToolCall `json:"tool_calls,omitempty"`
}
