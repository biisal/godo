package main

import (
	"fmt"
	"log"
	"os"

	"github.com/biisal/todo-cli/config"
	"github.com/biisal/todo-cli/todos/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func initLogging() (*os.File, error) {
	f, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to set up logging: %w", err)
	}
	return f, nil
}
func main() {
	logfile, _ := initLogging()
	defer logfile.Close()
	defer fmt.Println("config are set up")
	if err := config.MustLoad(); err != nil {
		fmt.Printf("Failed To Load Config %v: ", err)
		os.Exit(1)
	}
	defer config.Cfg.DB.Close()
	fLOgger := log.New(logfile, "[GODO-AGENT]", log.Ldate|log.Ltime|log.Lshortfile)
	m := ui.InitialModel(fLOgger)
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
