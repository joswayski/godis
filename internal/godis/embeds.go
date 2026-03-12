package godis

import (
	"log/slog"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/messages"
	"github.com/joswayski/godis/internal/webhooks"
)

var (
	xitterRegex    = regexp.MustCompile(`https?://(?:[a-zA-Z0-9-]+\.)?(?:twitter|x)\.com(/[^\s]*)?`)
	facebookRegex  = regexp.MustCompile(`https?://(?:[a-zA-Z0-9-]+\.)?facebook\.com(/(?:share|reel)/[^\s]*)`) // Don't replace profile links
	instagramRegex = regexp.MustCompile(`https?://(?:[a-zA-Z0-9-]+\.)?instagram\.com(/(?:p|reel|reels|tv|stories)/[^\s]*)`)
)

func (g *Godis) HandleEmbeds(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !g.HasReplaceableLink(m.Content) {
		return
	}

	newMessage := m.Content
	newMessage = xitterRegex.ReplaceAllString(newMessage, "https://vxtwitter.com$1")
	newMessage = facebookRegex.ReplaceAllString(newMessage, "https://facebed.com$1")
	newMessage = instagramRegex.ReplaceAllString(newMessage, "https://eeinstagram.com$1")

	godisWebhook, err := webhooks.GetGodisWebhook(s, m)
	if err != nil {
		slog.Error("Godis webhook not found!", "error", err.Error())
		return
	}

	nameToUse := messages.GetUsername(m.Message)

	// Download all the files
	allFiles, closers := messages.GetFilesInMessage(m)

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

func (g *Godis) HasReplaceableLink(content string) bool {
	return xitterRegex.MatchString(content) || facebookRegex.MatchString(content) || instagramRegex.MatchString(content)

}
