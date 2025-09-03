package main

import (
	"fmt"
	"os"

	"github.com/biisal/godo/config"
	"github.com/biisal/godo/logger"
	"github.com/biisal/godo/tui/actions/agent"
	"github.com/biisal/godo/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	Flogger := logger.NewLogger("logs.log", "[GODO-APP]", logger.Error)
	Flogger.Info("Starting GODO-AGENT")
	defer Flogger.Close()
	if err := config.MustLoad(); err != nil {
		Flogger.Error("Error loading config:", err)
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
			m.FLogger.Info("Got message: ", msg)
			p.Send(config.StreamMsg{Text: msg.Text, IsUser: msg.IsUser})
			m.FLogger.Info("Send message: ", msg)
		}
	}()
	if _, err := p.Run(); err != nil {
		m.FLogger.Error("Error running program:", err)
		fmt.Printf("Alas, there's been an error: %v", err)
	}

	fmt.Println("Goodbye!")
}
