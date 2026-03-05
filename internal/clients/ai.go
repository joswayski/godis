package ai

import (
	"github.com/joswayski/godis/internal/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func CreateClient(cfg config.GodisConfig) openai.Client {
	client := openai.NewClient(
		option.WithAPIKey(cfg.AIApiKey),
		option.WithBaseURL(cfg.AIApiBaseUrl),
		option.WithHeader("HTTP-Referer", "https://github.com/joswayski/godis"),
		option.WithHeader("X-OpenRouter-Title", "Godis"),
	)

	return client

}
