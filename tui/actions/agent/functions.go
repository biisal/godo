package agent

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
	"github.com/openai/openai-go/shared/constant"
)

const (
	PerformSqlFunc      = "PerformSql"
	RunShellCommandFunc = "RunShellCommand"
	ReadSkillFunc       = "ReadSkill"
)

func FormattedFunctions() []openai.ChatCompletionToolParam {
	return []openai.ChatCompletionToolParam{
		{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name: PerformSqlFunc,
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
	}
}
