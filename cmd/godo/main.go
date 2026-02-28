package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/biisal/godo/internal/bus"
	"github.com/biisal/godo/internal/config"
	"github.com/biisal/godo/internal/logger"
	"github.com/biisal/godo/internal/tui/actions/agent"
	"github.com/biisal/godo/internal/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

func main() {
	closeLog, err := logger.Init("logs.log", slog.LevelDebug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer closeLog()

	slog.Info("Starting GODO-AGENT")
	slog.Info("isDarkTheme", "darkmode", termenv.HasDarkBackground())

	if err = config.MustLoad(); err != nil {
		slog.Error("Error loading config", "err", err)
		fmt.Printf("Failed To Load Config: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err = config.Cfg.DB.Close(); err != nil {
			slog.Error("error closing db", "err", err)
		}
	}()

	bot := agent.NewBot()

	history, err := bot.GetChatHistoryFromDB()
	if err != nil {
		slog.Error("Error getting chat history from DB", "err", err)
		fmt.Println("Exiting")
		os.Exit(1)
	}
	bot.History = *history

	m := ui.InitialModel(bot)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := tea.ClearScreen(); err != nil {
		slog.Error("Error clearing screen", "err", err)
	}

	go func() {
		for msg := range bus.StreamResponse {
			p.Send(msg)
		}
	}()

	if _, err := p.Run(); err != nil {
		slog.Error("Error running program", "err", err)
		fmt.Printf("Alas, there's been an error: %v\n", err)
	}

	fmt.Println("Goodbye!")
}
