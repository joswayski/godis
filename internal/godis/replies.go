package godis

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/messages"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

func (g *Godis) HandleReplies(s *discordgo.Session, m *discordgo.MessageCreate) {
	slog.Info("Message handling in replies", "message", m)

	if !g.Config.AIEnabled {
		return
	}

	if g.HasReplaceableLink(m.Content) {
		return // Wait for the other goroutine to replace the message
	}

	if !g.Config.AIAllowedServers[m.GuildID] && !g.Config.AIAllowedChannels[m.ChannelID] {
		return // If we're not in an allowed server / channel combo, return
		// TODO improve this
	}

	if m.Author.ID == s.State.User.ID {
		// Ignore own messages
		return
	}

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

	// TODO streaming responses, chat indicator
	// Give it context of the messages before our current one
	history, err := s.ChannelMessages(m.ChannelID, g.Config.AINumberOfMessagesInHistory, m.ID, "", "")

	if err != nil {
		slog.Error("Error fetching channel history", "error", err.Error())
	}

	// for i, v := range history {
	// 	slog.Info("New message", "num", i+1, "content", v.Content)
	// 	return
	// }

	// Messages come newest first, so reverse it
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		role := responses.EasyInputMessageRoleUser
		content := messages.GetContent(msg)
		// TODO get files

		// Our own messages
		if msg.Author.ID == s.State.User.ID {
			role = responses.EasyInputMessageRoleAssistant
			content = msg.Content
		}

		inputItems = append(inputItems, responses.ResponseInputItemParamOfMessage(content, role))
	}

	// Add our newest message to the end
	inputItems = append(inputItems, responses.ResponseInputItemParamOfMessage(messages.GetContent(currentMsg), responses.EasyInputMessageRoleUser))

	for i, inputitem := range inputItems {
		slog.Info("Input item message", "num", i+1, "content", inputitem)
	}

	params.Input = responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems}

	response, err := g.AIClient.Responses.New(context.TODO(), params)

	if err != nil {
		slog.Error("Error ocurred generating response", "error", err.Error(), "message", m.Content)
		return
	}

	slog.Info("response", "response", response.ToolChoice)

	if len(response.Output) == 0 || strings.Contains(strings.ToLower(response.OutputText()), "no_response") {
		return
	}

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
