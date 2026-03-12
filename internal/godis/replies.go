package godis

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/messages"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// https://developers.openai.com/api/reference/go/

func (g *Godis) HandleReplies(s *discordgo.Session, m *discordgo.MessageCreate) {
	slog.Info("Message handling in replies", "message", m)

	if !g.Config.AIEnabled {
		return
	}

	if !g.IsAIAllowed(m.GuildID, m.ChannelID) {
		return
	}

	if g.HasReplaceableLink(m.Content) {
		return // Wait for the other goroutine to replace the message
	}

	if m.Author.ID == s.State.User.ID {
		// Ignore own messages
		return
	}

	// Sometimes messages don't have embeds populated right away
	shouldRefetch := shouldRefetchForEmbeds(m.Message)
	currentMsg := m.Message
	if shouldRefetch {
		time.Sleep(1500 * time.Millisecond)
		refetchedMessage, err := s.ChannelMessage(m.ChannelID, m.ID)
		if err != nil {
			slog.Error("Error refetching the channel message while retrieving embeds", "error", err.Error())
		} else {
			currentMsg = refetchedMessage
		}
	}

	params := responses.ResponseNewParams{
		// https://openrouter.ai/docs/api/reference/responses/basic-usage
		Model:           g.Config.AIApiModels[0], // TODO allow fallbacks / retries
		Instructions:    openai.String(g.Config.AISystemPrompt),
		MaxOutputTokens: openai.Int(int64(g.Config.AIMaxOutputTokens)),
	}

	var inputItems responses.ResponseInputParam

	// Give it context of the messages before our current one
	history, err := s.ChannelMessages(m.ChannelID, g.Config.AINumberOfMessagesInHistory, m.ID, "", "")

	if err != nil {
		slog.Error("Error fetching channel history", "error", err.Error())
	}

	// Messages come newest first, so reverse it
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		role := responses.EasyInputMessageRoleUser

		// Our own messages
		if msg.Author.ID == s.State.User.ID {
			role = responses.EasyInputMessageRoleAssistant
		}

		inputItems = append(inputItems, buildInputItem(msg, role))
	}

	// Add our newest message to the end
	inputItems = append(inputItems, buildInputItem(currentMsg, responses.EasyInputMessageRoleUser))

	params.Input = responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems}

	response, err := g.AIClient.Responses.New(context.TODO(), params)

	if err != nil {
		slog.Error("Error ocurred generating response", "error", err.Error(), "message", m.Content)
		return
	}

	if len(response.Output) == 0 || strings.Contains(strings.ToLower(response.OutputText()), "no_response") {
		// TODO use tools instead
		return
	}

	// Add channel typing indicator
	s.ChannelTyping(m.ChannelID)
	jitter := time.Duration(rand.IntN(2000)+100) * time.Millisecond
	time.Sleep(jitter)

	_, err = s.ChannelMessageSend(m.ChannelID, response.OutputText())

	if err != nil {
		slog.Error("Error sending discord message", "response", response)
	}
}

func shouldRefetchForEmbeds(msg *discordgo.Message) bool {
	if msg == nil {
		return false
	}

	if len(msg.Embeds) > 0 {
		// We already have them
		return false
	}

	// If it has a link, we should wait and get the message again
	return strings.Contains(msg.Content, "http://") || strings.Contains(msg.Content, "https://")
}

func buildInputItem(msg *discordgo.Message, role responses.EasyInputMessageRole) responses.ResponseInputItemUnionParam {
	content := msg.Content

	// Don't format in this manner for our own bot messages otherwise replies
	// tend to include the timestamp and user name when posting
	if role != responses.EasyInputMessageRoleAssistant {
		content = messages.GetContent(msg)
	}

	if len(msg.Attachments) == 0 {
		return responses.ResponseInputItemParamOfMessage(content, role)
	}

	// Handle messages with attachments
	// First add the text / user / timestamp
	parts := responses.ResponseInputMessageContentListParam{
		responses.ResponseInputContentParamOfInputText(content),
	}

	for _, att := range msg.Attachments {
		slog.Info("Attachment", "filename", att.Filename, "content_type", att.ContentType, "url", att.URL)

		if strings.HasPrefix(att.ContentType, "image/") {
			parts = append(parts, responses.ResponseInputContentUnionParam{
				OfInputImage: &responses.ResponseInputImageParam{
					ImageURL: openai.String(att.URL),
					Detail:   responses.ResponseInputImageDetailAuto,
				},
			})
		} else if strings.HasPrefix(att.ContentType, "audio/") {
			// Skip for now
			// TODO
			// Need to convert from .ogg to wav or mp3
			continue
		} else {
			// All other files
			parts = append(parts, responses.ResponseInputContentUnionParam{
				OfInputFile: &responses.ResponseInputFileParam{
					FileURL:  openai.String(att.URL),
					Filename: openai.String(att.Filename),
				},
			})
		}

	}

	return responses.ResponseInputItemParamOfMessage(parts, role)

}
