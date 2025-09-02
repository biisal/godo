package main

import (
	"fmt"
	"os"

	"github.com/biisal/todo-cli/config"
	"github.com/biisal/todo-cli/logger"
	"github.com/biisal/todo-cli/todos/actions/agent"
	"github.com/biisal/todo-cli/todos/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	Flogger := logger.NewLogger("logs.log", "[GODO-APP]", logger.Debug)
	defer Flogger.Close()
	if err := config.MustLoad(); err != nil {
		fmt.Printf("Failed To Load Config %v: ", err)
		os.Exit(1)
	}
	defer config.Cfg.DB.Close()
	m := ui.InitialModel(Flogger)
	p := tea.NewProgram(m, tea.WithAltScreen())

	history, err := agent.GetChatHistoryFromDB()
	if err != nil {
		m.FLogger.Error("Error getting chat history from DB: ", err)
		fmt.Println("Exiting")
		os.Exit(1)
	}
	agent.History = *history
	tea.ClearScreen()
	go func() {
		for msg := range config.StreamResponse {
			m.FLogger.Debug("Got message: ", msg)
			p.Send(config.StreamMsg{Text: msg.Text, IsUser: msg.IsUser})
			m.FLogger.Debug("Send message: ", msg)
		}
	}()
	if _, err := p.Run(); err != nil {
		m.FLogger.Error("Error running program:", err)
		fmt.Printf("Alas, there's been an error: %v", err)
	}

	fmt.Println("Goodbye!")
}
