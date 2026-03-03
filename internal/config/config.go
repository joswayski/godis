package config

import (
	"log/slog"
	"os"

	"github.com/joswayski/godis/internal/env"
)

type GodisConfig struct {
	DiscordToken string
}

func GetConfig() GodisConfig {
	env.GetEnv()

	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		slog.Error("DISCORD_TOKEN not set!")
		os.Exit(1)
	}

	slog.Debug("Config loaded!")
	return GodisConfig{
		DiscordToken: discordToken,
	}
}
