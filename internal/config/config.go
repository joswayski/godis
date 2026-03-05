package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joswayski/godis/internal/env"
)

type GodisConfig struct {
	DiscordToken      string
	OpenRouterToken   string
	AIAllowedServers  map[string]bool
	AIAllowedChannels map[string]bool // These are globally unique
}

func GetConfig() GodisConfig {
	env.GetEnv()

	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		slog.Error("DISCORD_TOKEN not set!")
		os.Exit(1)
	}

	openRouterToken := os.Getenv("OPENROUTER_TOKEN")
	if openRouterToken == "" {
		slog.Warn("OPENROUTER_TOKEN not set! AI capabilities will be disabled.")
	}

	aiAllowedServersList := os.Getenv("AI_ALLOWED_SERVERS")
	if aiAllowedServersList == "" {
		slog.Warn("No servers have been specified. AI features will be disabled!")
	}

	var aiAllowedServers = make(map[string]bool)
	for server := range strings.SplitSeq(aiAllowedServersList, ",") {
		s := strings.TrimSpace(server)
		if s != "" {
			aiAllowedServers[s] = true
		}
	}

	aiAllowedChannelsList := os.Getenv("AI_ALLOWED_CHANNELS")
	if aiAllowedChannelsList == "" {
		slog.Warn("No channels have been specified. AI features will be disabled!")
	}

	var aiAllowedChannels = make(map[string]bool)
	for channel := range strings.SplitSeq(aiAllowedChannelsList, ",") {
		s := strings.TrimSpace(channel)
		if s != "" {
			aiAllowedChannels[s] = true
		}
	}

	// TODO in the future, allow "ALL" as an input

	slog.Debug("Config loaded!")
	return GodisConfig{
		DiscordToken:      discordToken,
		OpenRouterToken:   openRouterToken,
		AIAllowedServers:  aiAllowedServers,
		AIAllowedChannels: aiAllowedChannels,
	}
}
