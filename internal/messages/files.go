package messages

import "github.com/bwmarrin/discordgo"

func HasAttachableMedia(msg *discordgo.Message) bool {
	// Check our own message
	if len(msg.Attachments) > 0 {
		return true
	}

	for _, emb := range msg.Embeds {
		if hasMedia(emb) {
			return true
		}
	}

	// Check any reply
	if msg.ReferencedMessage != nil {
		if len(msg.ReferencedMessage.Attachments) > 0 {
			return true
		}
		for _, emb := range msg.ReferencedMessage.Embeds {
			if hasMedia(emb) {
				return true
			}

		}
	}

	return false
}

func hasMedia(emb *discordgo.MessageEmbed) bool {
	if emb.Image != nil && emb.Image.URL != "" {
		return true
	}

	if emb.Thumbnail != nil && emb.Thumbnail.URL != "" {
		return true
	}

	return false
}
