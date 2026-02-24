package main

import (
	"log/slog"

	"github.com/joswayski/godis/internal/config"
	"github.com/joswayski/godis/internal/logger"
)

func main() {
	logger.Init()
	slog.Info("Godis is starting...")

	config.GetConfig()

	slog.Info("Godis is ready!")

}
