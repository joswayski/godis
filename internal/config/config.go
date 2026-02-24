package config

import (
	"log"
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
		log.Fatal("DISCORD_TOKEN not set!")
	}

	return GodisConfig{
		DiscordToken: discordToken,
	}
}
