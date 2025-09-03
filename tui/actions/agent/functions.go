package agent

import "github.com/biisal/godo/tui/models/agent"

func FormattedFunctions() []agent.Tool {
	return []agent.Tool{
		{
			FunctionDeclarations: []agent.FunctionDeclaration{
				getFuncFormatted(
					PerformSqlFunc,
					`Execute SQL queries on the SQLite database for the 'todos' table. 
The table has columns: Id (INTEGER PRIMARY KEY), Title (TEXT), Description (TEXT), Done (BOOLEAN).
Use SELECT queries to fetch data, INSERT to add todos, UPDATE to modify them, 
and DELETE to remove them. Always write valid SQLite queries. 
Only interact with the 'todos' table.`,
					map[string]agent.Property{
						"query": {
							Type:        "string",
							Description: "The sqlite3 query to execute",
						},
					},
					[]string{"query"},
				),
			},
		},
	}
}
