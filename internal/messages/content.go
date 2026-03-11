package messages

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

func GetContent(m *discordgo.Message) string {
	content := GetUsername(m) + ": " + m.Content + " - Timestamp: " + m.Timestamp.Format(time.RFC3339)


	// TODO this may be empty
	for _, embed := range m.Embeds {
		if embed.Title != "" {
			content += " - [Link: Title: " + embed.Title
		}

		if embed.Description != "" {
			content += " - Description: " + embed.Description
		}

		if embed.Title != "" {
			content += "]"
		}
	}
	return content
}
