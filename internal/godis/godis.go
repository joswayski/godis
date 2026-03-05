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
		if g.Config.AIAllowedServers[m.GuildID] && g.Config.AIAllowedChannels[m.ChannelID] {
			go g.HandleReplies(s, m)
		}
	}

}
