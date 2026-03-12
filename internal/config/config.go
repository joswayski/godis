package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joswayski/godis/internal/env"
)

const defaultAINumberOfMessagesInHistory = 20
const defaultAIMaxOutputTokens = 1000

type GodisConfig struct {
	DiscordToken                string
	AIApiKey                    string
	AIApiBaseUrl                string
	AIApiModels                 []string
	AIAllowedServers            map[string]bool
	AIAllowedChannels           map[string]bool // These are globally unique
	AISystemPrompt              string
	AIEnabled                   bool
	AINumberOfMessagesInHistory int
	AIMaxOutputTokens           int
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

	var aiAllowedServers = make(map[string]bool)
	for server := range strings.SplitSeq(aiAllowedServersList, ",") {
		s := strings.TrimSpace(server)
		if s != "" {
			aiAllowedServers[s] = true
		}
	}

	aiAllowedChannelsList := os.Getenv("AI_ALLOWED_CHANNELS") // TODO in the future, allow "ALL" as an input
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

	if len(aiAllowedServers) == 0 && len(aiAllowedChannels) == 0 {
		slog.Warn("No server or channels have been specified. AI features will be disabled!")
		aiEnabled = false
	}

	aiNumberOfMessagesInHistory, err := strconv.Atoi(os.Getenv("AI_NUMBER_OF_MESSAGES_IN_HISTORY"))
	if err != nil {
		slog.Warn("Invalid AI_NUMBER_OF_MESSAGES_IN_HISTORY detected, will be using the default", "value", defaultAINumberOfMessagesInHistory)
		aiNumberOfMessagesInHistory = defaultAINumberOfMessagesInHistory
	}

	if aiNumberOfMessagesInHistory > 100 { // todo make param
		slog.Warn(fmt.Sprintf("AI_NUMBER_OF_MESSAGES_IN_HISTORY is greater than 100 (received %d), limiting to %d to keep costs low.", aiNumberOfMessagesInHistory, defaultAINumberOfMessagesInHistory))
		aiNumberOfMessagesInHistory = 100
	}

	aiMaxOutputTokens, err := strconv.Atoi(os.Getenv("AI_MAX_OUTPUT_TOKENS"))
	if err != nil {
		slog.Warn("Invalid AI_MAX_OUTPUT_TOKENS detected, will be using the default", "value", defaultAIMaxOutputTokens)
		aiMaxOutputTokens = defaultAIMaxOutputTokens
	}

	if aiMaxOutputTokens > 1500 {
		// TODO might need a rework with images
		slog.Warn(fmt.Sprintf("AI_MAX_OUTPUT_TOKENS is greater than 1500 (received %d) limiting to %d to make sure messages fit in the Discord limit. ", aiMaxOutputTokens, defaultAIMaxOutputTokens))
		aiMaxOutputTokens = defaultAIMaxOutputTokens
	}

	slog.Info("Config Loaded!", "ai_enabled", aiEnabled, "ai_allowed_servers", aiAllowedServers, "ai_allowed_channels", aiAllowedChannels, "ai_base_url", aiApiBaseUrl, "ai_models", aiApiModels, "ai_number_of_messages_in_history", aiNumberOfMessagesInHistory, "ai_max_output_tokens", aiMaxOutputTokens)

	return GodisConfig{
		DiscordToken:                discordToken,
		AIApiKey:                    aiApiKey,
		AIAllowedServers:            aiAllowedServers,
		AIAllowedChannels:           aiAllowedChannels,
		AISystemPrompt:              aiSystemPrompt,
		AIEnabled:                   aiEnabled,
		AIApiBaseUrl:                aiApiBaseUrl,
		AIApiModels:                 aiApiModels,
		AINumberOfMessagesInHistory: aiNumberOfMessagesInHistory,
		AIMaxOutputTokens:           aiMaxOutputTokens,
	}
}
