package embeds

import (
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
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
	slog.Info("Received message", "content", m.Message.Content, "author", m.Message.Author.Username)
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
		// Publish the updated message

		// Delete the old one
	}

}
