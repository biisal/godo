package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/biisal/godo/internal/config"
	"github.com/biisal/godo/internal/logger"
	"github.com/biisal/godo/internal/tui/actions/agent"
	"github.com/muesli/termenv"
)

func initLogger() func() {
	closeLog, err := logger.Init("logs.log", slog.LevelDebug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
		os.Exit(1)
	}
	slog.Info("Starting GODO-AGENT")
	slog.Info("isDarkTheme", "darkmode", termenv.HasDarkBackground())
	return closeLog
}

func initConfig() {
	if err := config.MustLoad(); err != nil {
		slog.Error("Error loading config", "err", err)
		fmt.Printf("Failed To Load Config: %v\n", err)
		os.Exit(1)
	}
}

func initBot() *agent.Bot {
	bot := agent.NewBot()
	history, err := bot.GetChatHistoryFromDB()
	if err != nil {
		slog.Error("Error getting chat history from DB", "err", err)
		fmt.Println("Exiting")
		os.Exit(1)
	}
	bot.History = *history
	return bot
}
