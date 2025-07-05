package agent

import "github.com/biisal/todo-cli/todos/models/agent"

func FormattedFunctions() []agent.Tool {
	return []agent.Tool{
		{
			FunctionDeclarations: []agent.FunctionDeclaration{
				getFuncFormatted(
					AddTodoFunc,
					"Use to add todo in the todos list",
					map[string]agent.Property{
						"title": {
							Type:        "string",
							Description: "The title of the todo",
						},
						"description": {
							Type:        "string",
							Description: "The description of the todo",
						},
					},
					[]string{"title", "description"},
				),
				getFuncFormatted(
					GetTodosInfoFunc,
					"Returns the todos info [total , completed , remains]",
					map[string]agent.Property{},
					[]string{},
				),
				getFuncFormatted(
					GetTodosFunc,
					"Get all todos with ID, title, description and Completed status",
					map[string]agent.Property{},
					[]string{},
				),
				getFuncFormatted(
					ToggleDoneFunc,
					"Use to update the done status of the todo with fixed value or toggle it",
					map[string]agent.Property{
						"id": {
							Type:        "number",
							Description: "The id of the todo",
						},
						"done": {
							Type:        "boolean",
							Description: "The new done status of the todo",
						},
					},
					[]string{"id"},
				),
				getFuncFormatted(
					DeleteTodoFunc,
					"Delete todo by ID",
					map[string]agent.Property{
						"id": {
							Type:        "number",
							Description: "The id of the todo",
						},
					},
					[]string{"id"},
				),
				getFuncFormatted(
					GetTodoBYIDFunc,
					"Get todo by ID",
					map[string]agent.Property{
						"id": {
							Type:        "number",
							Description: "The id of the todo",
						},
					},
					[]string{"id"},
				),
			},
		},
	}
}
