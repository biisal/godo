package agent

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
	"github.com/openai/openai-go/shared/constant"
)

const (
	PerformSQLFunc       = "PerformSql"
	RunShellCommandFunc  = "RunShellCommand"
	ReadSkillFunc        = "ReadSkill"
	GlobSearchFunc       = "GlobSearch"
	ReadFilesFunc        = "ReadFiles"
	ProjectTreeFunc      = "ProjectTree"
	DuckDuckGoSearchFunc = "DuckDuckGoSearch"
	ScrapePageFunc       = "ScrapePage"
	WriteFileFunc        = "WriteFile"
	EditFileFunc         = "EditFile"
	PatchFileFunc        = "PatchFile"
	InsertAtLineFunc     = "InsertAtLine"
)

var tools = map[string]func(openai.ChatCompletionMessageToolCall) (any, bool, error){
	PerformSQLFunc:       runPerformSql,
	RunShellCommandFunc:  runShellCommand,
	ReadSkillFunc:        runReadSkill,
	GlobSearchFunc:       runGlobSearch,
	ReadFilesFunc:        runReadFiles,
	ProjectTreeFunc:      runProjectTree,
	DuckDuckGoSearchFunc: runDuckDuckGoSearch,
	ScrapePageFunc:       runScrapePage,
	WriteFileFunc:        runWriteFile,
	EditFileFunc:         runEditFile,
	PatchFileFunc:        runPatchFile,
	InsertAtLineFunc:     runInsertAtLine,
}

func FormattedFunctions() []openai.ChatCompletionToolParam {
	return []openai.ChatCompletionToolParam{
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name: PerformSQLFunc,
				Description: openai.String(`Execute any SQLite query on the 'todos' database.
CRITICAL: You MUST use this tool for ALL todo-related operations (listing, adding, completing, editing, deleting, finding).
DO NOT use the RunShellCommand tool for todo management.
Table schema: todos (Id INTEGER PRIMARY KEY, Title TEXT, Description TEXT, Done BOOLEAN)
Always write valid SQLite syntax and return the raw output.`),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "The SQLite query to execute (SELECT, INSERT, UPDATE, DELETE, etc.)",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name:        RunShellCommandFunc,
				Description: openai.String(`Execute a shell command and return its output. Use with caution.`),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"command": map[string]any{
							"type":        "string",
							"description": "The shell command to execute.",
						},
					},
					"required": []string{"command"},
				},
			},
		},
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name: ReadSkillFunc,
				Description: openai.String(`Read the instructions for a specific skill from its markdown file.
Provide the name of the skill (without the .md extension). Use this when you need specific instructions provided in the system prompt's skills list.`),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"skillName": map[string]any{
							"type":        "string",
							"description": "The name of the skill to read.",
						},
					},
					"required": []string{"skillName"},
				},
			},
		},
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name: GlobSearchFunc,
				Description: openai.String(`Find files by glob pattern under a root directory.
Supports standard glob wildcards plus recursive ** (example: src/**/*.jsx).`),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"pattern": map[string]any{
							"type":        "string",
							"description": "Glob pattern to match (for example: src/**/*.jsx, **/*.go, *.md).",
						},
						"root": map[string]any{
							"type":        "string",
							"description": "Optional root directory to search from. Defaults to current working directory.",
						},
					},
					"required": []string{"pattern"},
				},
			},
		},
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name: ReadFilesFunc,
				Description: openai.String(`Read multiple files in one tool call.
Returns file content and metadata for each requested path.`),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"paths": map[string]any{
							"type":        "array",
							"description": "List of file paths to read.",
							"items": map[string]any{
								"type": "string",
							},
						},
						"maxBytesPerFile": map[string]any{
							"type":        "integer",
							"description": "Optional byte limit per file. Defaults to 65536.",
						},
					},
					"required": []string{"paths"},
				},
			},
		},
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name: ProjectTreeFunc,
				Description: openai.String(`Return an at-a-glance project directory tree.
Use this to quickly inspect folder/file structure.`),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"root": map[string]any{
							"type":        "string",
							"description": "Optional root directory for the tree. Defaults to current working directory.",
						},
						"maxDepth": map[string]any{
							"type":        "integer",
							"description": "Optional depth limit. Defaults to 4.",
						},
						"includeFiles": map[string]any{
							"type":        "boolean",
							"description": "Whether to include files in addition to directories. Defaults to true.",
						},
					},
				},
			},
		},
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name: DuckDuckGoSearchFunc,
				Description: openai.String(`Search DuckDuckGo by POSTing a query and return ranked results.
Returns a list of results with title, URL, and description/snippet.`),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "Search query to send to DuckDuckGo.",
						},
						"maxResults": map[string]any{
							"type":        "integer",
							"description": "Optional maximum number of results. Defaults to 10.",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name: ScrapePageFunc,
				Description: openai.String(`Fetch and scrape a webpage by URL.
Returns title, description, and extracted plain text content.`),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"url": map[string]any{
							"type":        "string",
							"description": "Page URL to fetch and scrape.",
						},
						"maxChars": map[string]any{
							"type":        "integer",
							"description": "Optional max number of characters of extracted text. Defaults to 8000.",
						},
					},
					"required": []string{"url"},
				},
			},
		},
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name: WriteFileFunc,
				Description: openai.String(`Create or overwrite a file on disk with the provided content.
Can optionally create parent directories and append instead of overwrite.`),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "File path to write.",
						},
						"content": map[string]any{
							"type":        "string",
							"description": "Text content to write.",
						},
						"createParents": map[string]any{
							"type":        "boolean",
							"description": "If true, create parent directories when missing.",
						},
						"append": map[string]any{
							"type":        "boolean",
							"description": "If true, append to existing file instead of overwriting.",
						},
					},
					"required": []string{"path", "content"},
				},
			},
		},
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name: EditFileFunc,
				Description: openai.String(`Edit part of a file without rewriting everything.
Use either oldString/newString replacement, or lineNumber/newContent replacement.`),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "File path to edit.",
						},
						"oldString": map[string]any{
							"type":        "string",
							"description": "Existing text to replace.",
						},
						"newString": map[string]any{
							"type":        "string",
							"description": "Replacement text for oldString.",
						},
						"lineNumber": map[string]any{
							"type":        "integer",
							"description": "1-based line number to replace.",
						},
						"newContent": map[string]any{
							"type":        "string",
							"description": "New content for the specified lineNumber.",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name: PatchFileFunc,
				Description: openai.String(`Apply a unified diff patch to a file.
The patch should target the same path and include proper hunk headers.`),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "Target file path for the patch.",
						},
						"patch": map[string]any{
							"type":        "string",
							"description": "Unified diff patch content.",
						},
					},
					"required": []string{"path", "patch"},
				},
			},
		},
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name:        InsertAtLineFunc,
				Description: openai.String(`Insert content at a specific line number in a file.`),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "File path to modify.",
						},
						"lineNumber": map[string]any{
							"type":        "integer",
							"description": "1-based line number where content will be inserted.",
						},
						"content": map[string]any{
							"type":        "string",
							"description": "Content to insert.",
						},
					},
					"required": []string{"path", "lineNumber", "content"},
				},
			},
		},
	}
}
