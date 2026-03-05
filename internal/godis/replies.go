package godis

import (
	"context"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
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
		Model:        g.Config.AIApiModels[0], // TODO allow fallbacks
		Input:        responses.ResponseNewParamsInputUnion{OfString: openai.String(m.Author.Username + ": " + m.Content)},
		Instructions: openai.String(g.Config.AISystemPrompt),
	}

	// Give it context
	s.ChannelMessages(m.ChannelID, 20, "", "", "")

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
