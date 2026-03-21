package godis

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/config"
	"github.com/openai/openai-go/v3"
)

type MessageBuffer struct {
	BufferedMessages []*discordgo.Message
	Timer            *time.Timer
}
type Godis struct {
	Config        config.GodisConfig
	AIClient      openai.Client
	mutex         sync.Mutex
	MessageBuffer map[string]MessageBuffer // String is the channel ID
}

func (g *Godis) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	go g.HandleEmbeds(s, m)

	go g.HandleReplies(s, m)

}

func (g *Godis) IsAIAllowed(serverId string, channelId string) bool {
	return g.Config.AIAllowedServers[serverId] || g.Config.AIAllowedChannels[channelId]
}
