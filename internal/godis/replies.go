package godis

import (
	"context"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/messages"
	"github.com/joswayski/godis/internal/webhooks"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

func (g *Godis) HandleReplies(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.WebhookID != "" && webhooks.IsGodisWebhook(m.WebhookID) {
		// if it was a self poast from the embed replacements, ignore it
		return
	}

	if m.Author.ID == s.State.User.ID {
		// Ignore own messages
		return
	}

	params := responses.ResponseNewParams{
		// https://openrouter.ai/docs/api/reference/responses/basic-usage
		Model:           g.Config.AIApiModels[0], // TODO allow fallbacks
		Instructions:    openai.String(g.Config.AISystemPrompt),
		MaxOutputTokens: openai.Int(1500), // TODO param?

	}

	var inputItems responses.ResponseInputParam

	// TODO streaming responses, chat indicator
	// Give it context
	history, err := s.ChannelMessages(m.ChannelID, 20, "", "", "")

	if err != nil {
		slog.Error("Error fetching channel history", "error", err.Error())
	}

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

	// Add the latest message to the end
	inputItems = append(inputItems, responses.ResponseInputItemParamOfMessage(messages.GetContent(m.Message), responses.EasyInputMessageRoleUser))

	params.Input = responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems}

	response, err := g.AIClient.Responses.New(context.TODO(), params)

	if err != nil {
		slog.Error("Error ocurred generating response", "error", err.Error(), "message", m.Content)
		return
	}

	slog.Info("response", "response", response)

	if len(response.Output) == 0 || strings.Contains(response.OutputText(), "NO_RESPONSE") {
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, response.OutputText())

	if err != nil {
		slog.Error("Error sending discord message", "response", response)
	}
}
