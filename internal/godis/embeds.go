package godis

import (
	"log/slog"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/files"
	"github.com/joswayski/godis/internal/webhooks"
)

var (
	xitterRegex    = regexp.MustCompile(`https?://(?:[a-zA-Z0-9-]+\.)?(?:twitter|x)\.com(/[^\s]*)?`)
	facebookRegex  = regexp.MustCompile(`https?://(?:[a-zA-Z0-9-]+\.)?facebook\.com(/(?:share|reel)/[^\s]*)`) // Don't replace profile links
	instagramRegex = regexp.MustCompile(`https?://(?:[a-zA-Z0-9-]+\.)?instagram\.com(/(?:p|reel|reels|tv|stories)/[^\s]*)`)
)

func (g *Godis) HandleEmbeds(s *discordgo.Session, m *discordgo.MessageCreate) {

	newMessage := m.Content
	newMessage = xitterRegex.ReplaceAllString(newMessage, "https://vxtwitter.com$1")
	newMessage = facebookRegex.ReplaceAllString(newMessage, "https://facebed.com$1")
	newMessage = instagramRegex.ReplaceAllString(newMessage, "https://eeinstagram.com$1")

	if newMessage == m.Content {
		// No changes, ignore
		return
	}

	godisWebhook, err := webhooks.GetGodisWebhook(s, m)
	if err != nil {
		slog.Error("Godis webhook not found!", "error", err.Error())
		return
	}

	// Use the server nickname if it's there
	nameToUse := ""
	if m.Member != nil {
		nameToUse = m.Member.Nick

		if nameToUse == "" && m.Member.User != nil {
			nameToUse = m.Member.User.GlobalName
		}
	}

	if nameToUse == "" && m.Author != nil {
		// This is always there
		nameToUse = m.Author.Username
	}

	// Download all the files
	allFiles, closers := files.GetFilesInMessage(m)

	// Publish the updated message
	_, err = s.WebhookExecute(godisWebhook.ID, godisWebhook.Token, true, &discordgo.WebhookParams{
		Content:   newMessage,
		AvatarURL: m.Author.AvatarURL(""),
		// Don't re-ping if any tags
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{},
			Roles: []string{},
			Users: []string{},
		},
		Components: m.Components,
		Username:   nameToUse,
		Files:      allFiles,
	})

	// Close the file bodies from earlier
	for _, closer := range closers {
		closer.Close()
	}

	if err != nil {
		slog.Error("Error publishing new message", "error", err.Error(), "content", newMessage, "author", m.Author)
		return
	}

	// Delete the old one
	err = s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		slog.Error("Error deleting old message", "error", err.Error(), "content", m.Content, "author", m.Author)
	}

}
