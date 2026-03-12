package messages

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func GetContent(m *discordgo.Message) string {
	var b strings.Builder
	b.WriteString(GetUsername(m))
	b.WriteString(": ")
	b.WriteString(m.Content)
	b.WriteString(" - Timestamp: ")
	b.WriteString(m.Timestamp.Format(time.RFC3339))

	for _, embed := range m.Embeds {
		if embed.Title != "" {
			b.WriteString(" - [Link: Title: ")
			b.WriteString(embed.Title)
		}

		if embed.Description != "" {
			b.WriteString(" - Description: ")
			b.WriteString(embed.Description)
		}

		if embed.Title != "" {
			b.WriteString("]")
		}
	}
	return b.String()
}
