package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/biisal/godo/internal/config"
)

var version = "dev"

func main() {
	disableAutoUpdate := os.Getenv("DISABLE_AUTO_UPDATE") == "true"
	if err := runAutoUpdate(version, disableAutoUpdate); err != nil {
		slog.Error("Auto-update error", "err", err)
	}

	closeLog := initLogger()
	defer closeLog()

	initConfig()
	defer func() {
		if err := config.Cfg.DB.Close(); err != nil {
			slog.Error("error closing db", "err", err)
		}
	}()

	bot := initBot()
	run(bot)

	fmt.Println("Goodbye!")
}
