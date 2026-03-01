package main

import (
	"fmt"
	"log/slog"

	"github.com/biisal/godo/internal/bus"
	"github.com/biisal/godo/internal/tui/actions/agent"
	"github.com/biisal/godo/internal/tui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func run(bot *agent.Bot) {
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
}
