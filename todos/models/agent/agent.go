package agent

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

type AgentModel struct {
	PromptInput  textinput.Model
	StreamChunk  string
	ChatViewport viewport.Model
}

const (
	UserRole      = "user"
	AssistantRole = "assistant"
	ModelRole     = "model"
	SystemRole    = "system"
	ToolRole      = "tool"
	FunctionRole  = "function"
)

//
// type Message struct {
// 	Name       string     `json:"name,omitempty"`
// 	ToolCallId string     `json:"tool_call_id,omitempty"`
// 	Role       string     `json:"role"`
// 	Content    string     `json:"content"`
// 	Reasoning  string     `json:"reasoning,omitempty"`
// 	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
// }
//
// type PropertyType struct {
// 	Type        string   `json:"type"`
// 	Description string   `json:"description"`
// 	Enum        []string `json:"enum,omitempty"` // ✅ optional field
// }
// type Parameters struct {
// 	Type       string                  `json:"type"`
// 	Properties map[string]PropertyType `json:"properties"`
// 	Required   []string                `json:"required"`
// }
// type FunctionReq struct {
// 	Name        string     `json:"name"`
// 	Description string     `json:"description"`
// 	Parameters  Parameters `json:"parameters"`
// }
// type Tool struct {
// 	Type     string      `json:"type"`
// 	Function FunctionReq `json:"function"`
// }
//
// type AgentReq struct {
// 	Messages            []Message `json:"messages"`
// 	Model               string    `json:"model"`
// 	Temperature         float64   `json:"temperature"`
// 	MaxCompletionTokens int       `json:"max_completion_tokens"`
// 	TopP                float64   `json:"top_p"`
// 	Stream              bool      `json:"stream"`
// 	Tools               []Tool    `json:"tools"`
// 	ToolChoice          string    `json:"tool_choice"`
// }
//
// type ToolCall struct {
// 	Id       string `json:"id"`
// 	Type     string `json:"type"`
// 	Function struct {
// 		Name      string `json:"name"`
// 		Arguments string `json:"arguments"`
// 	} `json:"function"`
// }
//
// type AgentRes struct {
// 	ID      string `json:"id"`
// 	Object  string `json:"object"`
// 	Created int    `json:"created"`
// 	Model   string `json:"model"`
// 	Choices []struct {
// 		Delta        Message `json:"delta"`
// 		Message      Message `json:"message"`
// 		Logprobs     any     `json:"logprobs,omitempty"`
// 		FinishReason any     `json:"finish_reason,omitempty"`
// 	} `json:"choices"`
// 	Error struct {
// 		Message string `json:"message"`
// 	} `json:"error"`
// }
//
// type AgentResTeaMsg struct {
// 	History []Message
// 	Error   error
// }
//
//

// ===Gemini===
type FunctionResponse struct {
	WillContinue *bool          `json:"willContinue,omitempty"`
	ID           string         `json:"id,omitempty"`
	Name         string         `json:"name,omitempty"`
	Response     map[string]any `json:"response,omitempty"`
}
type FunctionCall struct {
	ID   string         `json:"id,omitempty"`
	Args map[string]any `json:"args,omitempty"`
	Name string         `json:"name,omitempty"`
}
type Part struct {
	Text             string            `json:"text,omitempty"`
	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
}

type Content struct {
	Role      string `json:"role"`
	Parts     []Part `json:"parts"`
	IsToolReq bool   `json:"-"`
}

type Items struct {
	Type string `json:"type"`
}
type Property struct {
	Type        string `json:"type"`
	Items       *Items `json:"items,omitempty"`
	Description string `json:"description"`
}
type Parameter struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}
type FunctionDeclaration struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Parameters  Parameter `json:"parameters"`
}
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations"`
}

type AgentReq struct {
	Contents          []Content `json:"contents"`
	SystemInstruction *Content  `json:"systemInstruction,omitempty"`
	Tools             []Tool    `json:"tools"`
}

type Candidate struct {
	Content      Content `json:"content"`
	Index        int     `json:"index"`
	FinishReason string  `json:"finishReason,omitempty"`
}

type PromptTokensDetail struct {
	Modality   string `json:"modality"`
	TokenCount int    `json:"tokenCount"`
}

type UsageMetadata struct {
	PromptTokenCount     int                  `json:"promptTokenCount"`
	CandidatesTokenCount int                  `json:"candidatesTokenCount"`
	TotalTokenCount      int                  `json:"totalTokenCount"`
	PromptTokensDetails  []PromptTokensDetail `json:"promptTokensDetails,omitempty"`
	ThoughtsTokenCount   int                  `json:"thoughtsTokenCount,omitempty"`
}
type AgentRes struct {
	Candidates    []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usageMetadata"`
	ModelVersion  string        `json:"modelVersion,omitempty"`
	ResponseId    string        `json:"responseId,omitempty"`
}
