package messages

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

func GetContent(m *discordgo.Message) string {
	return GetUsername(m) + ": " + m.Content + " - Timestamp: " + m.Timestamp.Format(time.RFC3339)
}
