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

	getEmbedContent(m.Embeds, &b)
	getAttachmentsContent(m.Attachments, &b)
	if m.ReferencedMessage != nil {
		// Only do one level
		b.WriteString(" - In a reply to: ")
		b.WriteString(GetUsername(m.ReferencedMessage))
		b.WriteString(": ")
		b.WriteString(m.ReferencedMessage.Content)
		getEmbedContent(m.ReferencedMessage.Embeds, &b)
		getAttachmentsContent(m.ReferencedMessage.Attachments, &b)
	}
	return b.String()
}

func getEmbedContent(embeds []*discordgo.MessageEmbed, b *strings.Builder) {
	for _, embed := range embeds {
		if embed.Title == "" && embed.Description == "" {
			continue
		}

		b.WriteString(" - [Link")

		if embed.Title != "" {
			b.WriteString(": Title: ")
			b.WriteString(embed.Title)
		}

		if embed.Description != "" {
			b.WriteString(": Description: ")
			b.WriteString(embed.Description)
		}

		b.WriteString("]")

	}
}

func getAttachmentsContent(attachments []*discordgo.MessageAttachment, b *strings.Builder) {
	for _, att := range attachments {

		b.WriteString(" - [Attachment: ")
		b.WriteString(att.Filename)

		if att.ContentType != "" {
			b.WriteString(" (")
			b.WriteString(att.ContentType)
			b.WriteString(" )")

		}

		b.WriteString("]")

	}
}
