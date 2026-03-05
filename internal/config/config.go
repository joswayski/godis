package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joswayski/godis/internal/env"
)

type GodisConfig struct {
	DiscordToken      string
	AIApiKey          string
	AIApiBaseUrl      string
	AIApiModels       []string
	AIAllowedServers  map[string]bool
	AIAllowedChannels map[string]bool // These are globally unique
	AISystemPrompt    string
	AIEnabled         bool
}

func GetConfig() GodisConfig {
	env.GetEnv()
	aiEnabled := true
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		slog.Error("DISCORD_TOKEN not set!")
		os.Exit(1)
	}

	aiApiKey := os.Getenv("AI_API_KEY")
	if aiApiKey == "" {
		slog.Warn("AI_API_KEY not set! AI capabilities will be disabled.")
		aiEnabled = false
	}

	aiApiBaseUrl := os.Getenv("AI_API_BASE_URL")
	if aiApiBaseUrl == "" {
		slog.Warn("AI_API_BASE_URL not set! AI capabilities will be disabled.")
		aiEnabled = false
	}

	var aiApiModels []string

	aiModelList := os.Getenv("AI_API_MODELS")
	for model := range strings.SplitSeq(aiModelList, ",") {
		m := strings.TrimSpace(model)
		if m != "" {
			aiApiModels = append(aiApiModels, m)
		}
	}

	if len(aiApiModels) == 0 {
		slog.Warn("AI_API_MODELS not set! AI capabilities will be disabled.")
		aiEnabled = false
	}

	aiAllowedServersList := os.Getenv("AI_ALLOWED_SERVERS") // TODO in the future, allow "ALL" as an input
	if aiAllowedServersList == "" {
		slog.Warn("No servers have been specified. AI features will be disabled!")
		aiEnabled = false
	}

	var aiAllowedServers = make(map[string]bool)
	for server := range strings.SplitSeq(aiAllowedServersList, ",") {
		s := strings.TrimSpace(server)
		if s != "" {
			aiAllowedServers[s] = true
		}
	}

	aiAllowedChannelsList := os.Getenv("AI_ALLOWED_CHANNELS") // TODO in the future, allow "ALL" as an input
	if aiAllowedChannelsList == "" {
		slog.Warn("No channels have been specified. AI features will be disabled!")
		aiEnabled = false
	}

	var aiAllowedChannels = make(map[string]bool)
	for channel := range strings.SplitSeq(aiAllowedChannelsList, ",") {
		s := strings.TrimSpace(channel)
		if s != "" {
			aiAllowedChannels[s] = true
		}
	}

	aiSystemPrompt := os.Getenv("AI_SYSTEM_PROMPT")
	if aiSystemPrompt == "" {
		slog.Warn("No AI_SYSTEM_PROMPT detected. AI features will be disabled!")
		aiEnabled = false
	}

	slog.Debug("Config loaded!")
	return GodisConfig{
		DiscordToken:      discordToken,
		AIApiKey:          aiApiKey,
		AIAllowedServers:  aiAllowedServers,
		AIAllowedChannels: aiAllowedChannels,
		AISystemPrompt:    aiSystemPrompt,
		AIEnabled:         aiEnabled,
		AIApiBaseUrl:      aiApiBaseUrl,
		AIApiModels:       aiApiModels,
	}
}
