package main

import (
	"context"
	"fmt"

	"github.com/biisal/godo/internal/logger"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/fatih/color"
)

func runAutoUpdate(cmd string, currentVersion string, disableAutoUpdate bool) error {
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
		return nil // Don't fail the app if update check fails
	}

	if latest.Version() == currentVersion {
		logger.Success("Already on the latest version (%s)", latest.Version())
		return nil
	}

	logger.Success("Updated from %s to %s", currentVersion, latest.Version())
	_ = cmd // cmd parameter reserved for future use
	color.Cyan("Please restart the application to use the new version")
	return nil
}

func CheckForUpdate(currentVersion string) error {
	if currentVersion == "" || currentVersion == "dev" || currentVersion == "latest" {
		return fmt.Errorf("cannot check for updates in development build")
	}

	repo := selfupdate.ParseSlug("biisal/godo")

	latest, err := selfupdate.UpdateSelf(context.Background(), currentVersion, repo)
	if err != nil {
		return fmt.Errorf("error checking for updates: %w", err)
	}

	if latest.Version() == currentVersion {
		fmt.Printf("You're already on the latest version: %s\n", currentVersion)
		return nil
	}

	fmt.Printf("New version available: %s -> %s\n", currentVersion, latest.Version())
	fmt.Println(latest.ReleaseNotes)

	return nil
}
