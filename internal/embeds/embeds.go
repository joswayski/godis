package embeds

import (
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/webhooks"
)

var (
	xitterRegex    = regexp.MustCompile(`https?://(?:[a-zA-Z0-9-]+\.)?(?:twitter|x)\.com(/[^\s]*)?`)
	facebookRegex  = regexp.MustCompile(`https?://(?:[a-zA-Z0-9-]+\.)?facebook\.com(/(?:share|reel)/[^\s]*)`) // Don't replace profile links
	instagramRegex = regexp.MustCompile(`https?://(?:[a-zA-Z0-9-]+\.)?instagram\.com(/(?:p|reel|reels|tv|stories)/[^\s]*)`)
)

func HandleEmbeds(s *discordgo.Session, m *discordgo.MessageCreate) {

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
	var allFiles []*discordgo.File
	var closers []io.Closer
	var fileMu sync.Mutex
	var fileWg sync.WaitGroup

	for _, attachment := range m.Attachments {
		fileWg.Add(1)
		go func(atch *discordgo.MessageAttachment) {
			defer fileWg.Done()
			response, err := http.Get(atch.URL)
			if err != nil {
				slog.Error("Error downloading file", "attachment", atch)
				return
			}

			if response.StatusCode != http.StatusOK {
				response.Body.Close()
				slog.Error("Non 200 status code downloading file", "status", response.Status, "attachment", atch)
				return
			}

			fileMu.Lock()
			closers = append(closers, response.Body)
			allFiles = append(allFiles, &discordgo.File{
				Name:        atch.Filename,
				ContentType: atch.ContentType,
				Reader:      response.Body,
			})
			fileMu.Unlock()
		}(attachment)
	}

	fileWg.Wait()

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
