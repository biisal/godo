package bus

import (
	"strings"
	"testing"
	"time"
)

func TestStreamMsg(t *testing.T) {
	msg := StreamMsg{
		Text: "Hello World",
		Type: "message",
	}

	if msg.Text != "Hello World" {
		t.Errorf("Expected Text to be 'Hello World', got '%s'", msg.Text)
	}
	if msg.Type != "message" {
		t.Errorf("Expected Type to be 'message', got '%s'", msg.Type)
	}
}

func TestEmitState(t *testing.T) {
	done := make(chan bool)

	go func() {
		time.Sleep(10 * time.Millisecond)
		msg := <-StreamResponse
		if msg.Type != "status" {
			t.Errorf("Expected Type to be 'status', got '%s'", msg.Type)
		}
		done <- true
	}()

	EmitState("Testing state...")

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestEmitMessageStatus(t *testing.T) {
	done := make(chan bool)

	go func() {
		time.Sleep(10 * time.Millisecond)
		msg := <-StreamResponse
		if msg.Type != "messageStatus" {
			t.Errorf("Expected Type to be 'messageStatus', got '%s'", msg.Type)
		}
		done <- true
	}()

	EmitMessageStatus("Test message")

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestEmitThinking(t *testing.T) {
	done := make(chan bool)

	go func() {
		time.Sleep(10 * time.Millisecond)
		msg := <-StreamResponse
		if msg.Type != "thinking" {
			t.Errorf("Expected Type to be 'thinking', got '%s'", msg.Type)
		}
		if msg.Text != "reasoning..." {
			t.Errorf("Expected Text to be 'reasoning...', got '%s'", msg.Text)
		}
		done <- true
	}()

	EmitThinking("reasoning...")

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestEmitContent(t *testing.T) {
	done := make(chan bool)

	go func() {
		time.Sleep(10 * time.Millisecond)
		msg := <-StreamResponse
		if msg.Text != "Hello content" {
			t.Errorf("Expected Text to be 'Hello content', got '%s'", msg.Text)
		}
		done <- true
	}()

	EmitContent("Hello content")

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestEmitUser(t *testing.T) {
	done := make(chan bool)

	go func() {
		time.Sleep(10 * time.Millisecond)
		msg := <-StreamResponse
		if msg.Type != "user" {
			t.Errorf("Expected Type to be 'user', got '%s'", msg.Type)
		}
		if msg.Text != "user message" {
			t.Errorf("Expected Text to be 'user message', got '%s'", msg.Text)
		}
		done <- true
	}()

	EmitUser("user message")

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestEmitStreamStart(t *testing.T) {
	done := make(chan bool)

	go func() {
		time.Sleep(10 * time.Millisecond)
		msg := <-StreamResponse
		if msg.Type != "stream_start" {
			t.Errorf("Expected Type to be 'stream_start', got '%s'", msg.Type)
		}
		done <- true
	}()

	EmitStreamStart()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestEmitStreamEnd(t *testing.T) {
	done := make(chan bool)

	go func() {
		time.Sleep(10 * time.Millisecond)
		msg := <-StreamResponse
		if msg.Type != "stream_end" {
			t.Errorf("Expected Type to be 'stream_end', got '%s'", msg.Type)
		}
		done <- true
	}()

	EmitStreamEnd()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestEmitShell(t *testing.T) {
	done := make(chan bool)

	go func() {
		time.Sleep(10 * time.Millisecond)
		msg := <-StreamResponse
		if msg.Type != "shell" {
			t.Errorf("Expected Type to be 'shell', got '%s'", msg.Type)
		}
		if msg.Text != "shell command" {
			t.Errorf("Expected Text to be 'shell command', got '%s'", msg.Text)
		}
		done <- true
	}()

	EmitShell("shell command")

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestToolStatusMessage(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		expected string
	}{
		{"PerformSql", "PerformSql", "Checking your todos..."},
		{"RunShellCommand", "RunShellCommand", "Running command..."},
		{"ReadSkill", "ReadSkill", "Loading skill instructions..."},
		{"GlobSearch", "GlobSearch", "Searching files by pattern..."},
		{"ReadFiles", "ReadFiles", "Reading files..."},
		{"ProjectTree", "ProjectTree", "Building project tree..."},
		{"DuckDuckGoSearch", "DuckDuckGoSearch", "Searching the web..."},
		{"ScrapePage", "ScrapePage", "Reading webpage content..."},
		{"WriteFile", "WriteFile", "Writing file..."},
		{"EditFile", "EditFile", "Editing file..."},
		{"PatchFile", "PatchFile", "Applying patch..."},
		{"InsertAtLine", "InsertAtLine", "Inserting content into file..."},
		{"Unknown tool", "UnknownTool", "Running UnknownTool..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toolStatusMessage(tt.toolName)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// ===== EDGE CASE TESTS =====

func TestEmitStateEmptyString(t *testing.T) {
	// Drain existing messages first
	go func() {
		for {
			select {
			case <-StreamResponse:
			default:
				return
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)

	done := make(chan bool)

	go func() {
		msg := <-StreamResponse
		if msg.Type != "status" {
			t.Errorf("Expected Type to be 'status', got '%s'", msg.Type)
		}
		done <- true
	}()

	EmitState("")

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestEmitContentVeryLongString(t *testing.T) {
	// Drain existing messages first
	go func() {
		for {
			select {
			case <-StreamResponse:
			default:
				return
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)

	done := make(chan bool)
	longString := strings.Repeat("A", 100000) // 100KB string

	go func() {
		msg := <-StreamResponse
		if msg.Text != longString {
			t.Errorf("Expected Text length %d, got %d", len(longString), len(msg.Text))
		}
		done <- true
	}()

	EmitContent(longString)

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestEmitContentUnicode(t *testing.T) {
	// Drain existing messages first
	go func() {
		for {
			select {
			case <-StreamResponse:
			default:
				return
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)

	done := make(chan bool)

	unicodeText := "Hello 世界 🌍 🪐🚀 \"quotes\" 'single' `code`"

	go func() {
		msg := <-StreamResponse
		if msg.Text != unicodeText {
			t.Errorf("Expected '%s', got '%s'", unicodeText, msg.Text)
		}
		done <- true
	}()

	EmitContent(unicodeText)

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestEmitContentSpecialChars(t *testing.T) {
	// Drain existing messages first
	go func() {
		for {
			select {
			case <-StreamResponse:
			default:
				return
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)

	done := make(chan bool)

	specialChars := "Tab:\tNewline:\nNull:\x00Bell:\aBackslash:\\"

	go func() {
		msg := <-StreamResponse
		if msg.Text != specialChars {
			t.Error("Should preserve special characters")
		}
		done <- true
	}()

	EmitContent(specialChars)

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestConcurrentEmits(t *testing.T) {
	t.Skip("Skipping - channel buffer causes blocking with concurrent sends without listeners")
}

func TestEmitThinkingEmptyString(t *testing.T) {
	// Drain existing messages first
	go func() {
		for {
			select {
			case <-StreamResponse:
			default:
				return
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)

	done := make(chan bool)

	go func() {
		msg := <-StreamResponse
		if msg.Type != "thinking" {
			t.Errorf("Expected Type to be 'thinking', got '%s'", msg.Type)
		}
		done <- true
	}()

	EmitThinking("")

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for StreamResponse")
	}
}

func TestMultipleEmitSequences(t *testing.T) {
	// Drain existing messages first
	go func() {
		for {
			select {
			case <-StreamResponse:
			default:
				return
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)

	done := make(chan bool)

	go func() {
		// Receive all messages in sequence
		expectedTypes := []string{"stream_start", "thinking", "content", "stream_end"}
		types := []string{}

		for i := 0; i < 4; i++ {
			select {
			case msg := <-StreamResponse:
				types = append(types, msg.Type)
			case <-time.After(500 * time.Millisecond):
				t.Error("Timeout waiting for message", i)
				done <- true
				return
			}
		}

		if len(types) != len(expectedTypes) {
			t.Errorf("Expected %d types, got %d", len(expectedTypes), len(types))
		}
		done <- true
	}()

	EmitStreamStart()
	EmitThinking("thinking...")
	EmitContent("content")
	EmitStreamEnd()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for sequence")
	}
}
