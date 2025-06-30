package agent

import "github.com/charmbracelet/bubbles/textinput"

type AgentModel struct {
	PromptInput textinput.Model
	Response    string
	History     []Message
}

const (
	UserRole      = "user"
	AssistantRole = "assistant"
	SystemRole    = "system"
)

type Message struct {
	Name       string     `json:"name,omitempty"`
	ToolCallId string     `json:"tool_call_id,omitempty"`
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type PropertyType struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"` // ✅ optional field
}

type FunctionReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  struct {
		Type       string                  `json:"type"`
		Properties map[string]PropertyType `json:"properties"`
		Required   []string                `json:"required"`
	} `json:"parameters"`
}
type Tool struct {
	Type     string      `json:"type"`
	Function FunctionReq `json:"function"`
}

type AgentReq struct {
	Messages            []Message `json:"messages"`
	Model               string    `json:"model"`
	Temperature         float64   `json:"temperature"`
	MaxCompletionTokens int       `json:"max_completion_tokens"`
	TopP                float64   `json:"top_p"`
	Stream              bool      `json:"stream"`
	Tools               []Tool    `json:"tools"`
	ToolChoice          string    `json:"tool_choice"`
}

type ToolCall struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type AgentRes struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message      Message `json:"message"`
		Logprobs     float64 `json:"logprobs"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type AgentResTeaMsg struct {
	History []Message
	Error   error
}
