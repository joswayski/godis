package godis

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/config"
	"github.com/joswayski/godis/internal/embeds"
	"github.com/openai/openai-go/v3"
)

type Godis struct {
	Config   config.GodisConfig
	AIClient openai.Client
}

func (g *Godis) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	slog.Info("Received message", "content", m.Message.Content, "author", m.Message.Author)

	go embeds.HandleEmbeds(s, m)

}
