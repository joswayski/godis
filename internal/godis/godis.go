package godis

import (
	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/config"
	"github.com/openai/openai-go/v3"
)

type Godis struct {
	Config   config.GodisConfig
	AIClient openai.Client
}

func (g *Godis) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {

	go g.HandleEmbeds(s, m)

	if g.Config.AIEnabled {
		go g.HandleReplies(s, m)
	}

}
