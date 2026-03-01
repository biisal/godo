package main

import (
	"context"
	"os"
	"syscall"

	"github.com/biisal/godo/internal/logger"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/fatih/color"
)

func runAutoUpdate(currentVersion string, disableAutoUpdate bool) error {
	if disableAutoUpdate {
		logger.Info("Auto-update disabled via config")
		return nil
	}

	if currentVersion == "" || currentVersion == "dev" || currentVersion == "latest" {
		logger.Info("Development build: skipping auto-update")
		return nil
	}

	logger.Info("Checking for updates... You can disable auto-update by setting DISABLE_AUTO_UPDATE=true in config")

	repo := selfupdate.ParseSlug("biisal/godo")

	latest, err := selfupdate.UpdateSelf(context.Background(), currentVersion, repo)
	if err != nil {
		logger.Error("Error checking for updates: %v", err)
		return nil
	}

	if latest.Version() == currentVersion {
		logger.Success("Already on the latest version (%s)", latest.Version())
		return nil
	}

	logger.Success("Updated from %s to %s", currentVersion, latest.Version())
	color.Cyan("Restarting with the new version...")

	// Re-exec the current process with the updated binary
	execPath, err := os.Executable()
	if err != nil {
		logger.Error("Could not determine executable path: %v", err)
		color.Cyan("Please restart the application manually to use the new version")
		return nil
	}

	return syscall.Exec(execPath, os.Args, os.Environ())
}
