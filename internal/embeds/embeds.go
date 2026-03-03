package embeds

import (
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/webhooks"
)

const (

	// Domains to replace
	xDomain         = "https://x.com"
	twitterDomain   = "https://twitter.com"
	facebookDomain  = "https://facebook.com"
	instagramDomain = "https://instagram.com"

	// Embed solutions - in the future we'll do our own parsing but these work
	vxTwitterDomain = "https://vxtwitter.com"
	facebedDomain   = "https://facebed.com"
	instaBedDomain  = "https://eeinstagram.com"
)

var instagramPostTypes = []string{"/p/", "/reel/", "/reels/", "/tv/", "/stories/"}

func HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {

	slog.Info("Received message", "content", m.Message.Content, "author", m.Message.Author)

	newMessage := m.Content
	shouldRepublish := false
	if strings.Contains(xDomain, newMessage) {
		newMessage = strings.ReplaceAll(newMessage, xDomain, vxTwitterDomain)
		shouldRepublish = true
	}

	if strings.Contains(twitterDomain, newMessage) {
		newMessage = strings.ReplaceAll(newMessage, twitterDomain, vxTwitterDomain)
		shouldRepublish = true
	}

	if strings.Contains(facebookDomain, newMessage) {
		newMessage = strings.ReplaceAll(newMessage, facebookDomain, facebedDomain)
		shouldRepublish = true
	}

	if strings.Contains(instagramDomain, newMessage) {
		for _, postType := range instagramPostTypes {
			if strings.Contains(newMessage, postType) {
				newMessage = strings.ReplaceAll(newMessage, instagramDomain, instaBedDomain)
				shouldRepublish = true
				break
			}
		}
	}

	if shouldRepublish {
		godisWebhook, err := webhooks.GetGodisWebhook(s, m)
		if err != nil {
			slog.Error("Godis webhook not found!", "error", err.Error())
			return
		}

		// Use the server nickname if it's there
		nameToUse := m.Member.Nick
		if nameToUse == "" {
			nameToUse = m.Member.User.GlobalName
		}

		if nameToUse == "" {
			// This is always there
			nameToUse = m.Author.Username
		}

		// Publish the updated message
		s.WebhookExecute(godisWebhook.ID, godisWebhook.Token, true, &discordgo.WebhookParams{
			Content:   newMessage,
			AvatarURL: m.Author.AvatarURL(""),
			// Don't re-ping if any tags
			AllowedMentions: &discordgo.MessageAllowedMentions{},
			Username:        nameToUse,
			// TODO ?
			// Files: m.Attachments,
		})

		// Delete the old one
		s.ChannelMessageDelete(m.ChannelID, m.ID)
	}

}
