package bus

import "fmt"

// StreamMsg represents a message sent through the event bus to the TUI.
type StreamMsg struct {
	Text   string
	IsUser bool
	Type   string
}

// StreamResponse is the channel used to communicate between the agent and the TUI.
var StreamResponse = make(chan StreamMsg)

// Emit sends a generic stream message.
func Emit(text string) {
	StreamResponse <- StreamMsg{Text: text}
}

// EmitStatus sends a status bar update (e.g. "Thinking...").
func EmitStatus(text string) {
	StreamResponse <- StreamMsg{Text: text, Type: "status"}
}

// EmitThinking sends streamed reasoning text.
func EmitThinking(text string) {
	StreamResponse <- StreamMsg{Text: text, Type: "thinking"}
}

// EmitContent sends streamed content text.
func EmitContent(text string) {
	StreamResponse <- StreamMsg{Text: text}
}

// EmitUser sends a user message for display.
func EmitUser(text string) {
	StreamResponse <- StreamMsg{Text: text, IsUser: true}
}

// EmitToolCall sends a status for an active tool call.
func EmitToolCall(name string) {
	EmitStatus(toolStatusMessage(name))
}

func toolStatusMessage(name string) string {
	switch name {
	case "PerformSql":
		return "Checking your todos..."
	case "RunShellCommand":
		return "Running command..."
	case "ReadSkill":
		return "Loading skill instructions..."
	case "GlobSearch":
		return "Searching files by pattern..."
	case "ReadFiles":
		return "Reading files..."
	case "ProjectTree":
		return "Building project tree..."
	case "DuckDuckGoSearch":
		return "Searching the web..."
	case "ScrapePage":
		return "Reading webpage content..."
	case "WriteFile":
		return "Writing file..."
	case "EditFile":
		return "Editing file..."
	case "PatchFile":
		return "Applying patch..."
	case "InsertAtLine":
		return "Inserting content into file..."
	default:
		return fmt.Sprintf("Running %s...", name)
	}
}

// EmitShell sends shell command or output for the side panel.
func EmitShell(text string) {
	StreamResponse <- StreamMsg{Text: text, Type: "shell"}
}
