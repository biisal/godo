package agent

import (
	"testing"

	agentModel "github.com/biisal/godo/internal/tui/models/agent"
)

func TestAppendMessageUpdatesOAIMessages(t *testing.T) {
	b := &Bot{
		systemPrompt: "You are a test agent.",
	}

	// 1. Initial manual initOAIMessages (what agentAPICall does)
	b.initOAIMessages()

	if len(b.oaiMessages) != 1 {
		t.Fatalf("expected 1 oaiMessage (system prompt), got %d", len(b.oaiMessages))
	}

	// 2. Simulate what AgentResponse does when user sends a prompt
	userMsg := agentModel.Message{
		Role:    agentModel.UserRole,
		Content: "Hello World",
	}
	b.appendMessage(userMsg)

	// 3. Verify both History and oaiMessages grew
	if len(b.History) != 1 {
		t.Errorf("expected 1 History item, got %d", len(b.History))
	}
	if len(b.oaiMessages) != 2 {
		t.Fatalf("expected 2 oaiMessages (system + user), got %d", len(b.oaiMessages))
	}

	// 4. Simulate assistant responding
	asstMsg := agentModel.Message{
		Role:    agentModel.AssistantRole,
		Content: "Hello Human",
	}
	b.appendMessage(asstMsg)

	if len(b.History) != 2 {
		t.Errorf("expected 2 History items, got %d", len(b.History))
	}
	if len(b.oaiMessages) != 3 {
		t.Fatalf("expected 3 oaiMessages (system + user + assistant), got %d", len(b.oaiMessages))
	}
}

func TestInitOAIMessagesRebuildsCorrectly(t *testing.T) {
	b := &Bot{
		systemPrompt: "System",
		History: []agentModel.Message{
			{Role: agentModel.UserRole, Content: "1"},
			{Role: agentModel.AssistantRole, Content: "2"},
		},
	}

	// It should build an array sized len(History) + 1 (the system prompt)
	b.initOAIMessages()

	if len(b.oaiMessages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(b.oaiMessages))
	}
}
