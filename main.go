package main

import (
	"fmt"
	"os"

	"github.com/biisal/todo-cli/config"
	"github.com/biisal/todo-cli/logger"
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

	go func() {
		for msg := range config.Ping {
			config.WriteLog(false, msg)
			p.Send(msg)
			m.AgentModel.ChatViewport.GotoBottom()
		}
	}()

	tea.ClearScreen()
	if _, err := p.Run(); err != nil {
		config.WriteLog(false, err)
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
