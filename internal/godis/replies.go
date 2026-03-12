package godis

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/messages"
	"github.com/openai/openai-go/v3"
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

	params := openai.ChatCompletionNewParams{
		// https://openrouter.ai/docs/quickstart
		Model:               g.Config.AIApiModels[0], // TODO allow fallbacks / retries,
		MaxCompletionTokens: openai.Int(int64(g.Config.AIMaxOutputTokens)),
		Modalities:          []string{"text", "image"},
	}

	var messages []openai.ChatCompletionMessageParamUnion

	// Give it context of the messages before our current one
	history, err := s.ChannelMessages(m.ChannelID, g.Config.AINumberOfMessagesInHistory, m.ID, "", "")

	if err != nil {
		slog.Error("Error fetching channel history", "error", err.Error())
	}
	messages = append(messages, openai.SystemMessage(g.Config.AISystemPrompt))

	// Messages come newest first, so reverse it
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		isAssistant := msg.Author.ID == s.State.User.ID // our own messages

		messages = append(messages, buildMessages(msg, isAssistant))
	}

	// Add our newest message to the end
	messages = append(messages, buildMessages(currentMsg, false))

	// Set the messages and send it
	params.Messages = messages
	response, err := g.AIClient.Chat.Completions.New(context.TODO(), params)

	if err != nil {
		slog.Error("Error ocurred generating response", "error", err.Error(), "message", m.Content)
		return
	}

	if len(response.Choices) == 0 || strings.Contains(strings.ToLower(response.Choices[0].Message.Content), "no_response") {
		// TODO use tools instead
		return
	}

	// Add channel typing indicator
	s.ChannelTyping(m.ChannelID)
	jitter := time.Duration(rand.IntN(2000)+100) * time.Millisecond
	time.Sleep(jitter)

	imageFiles := extractImages(response.Choices[0].Message.RawJSON())
	if len(imageFiles) > 0 {
		_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content: response.Choices[0].Message.Content,
			Files:   imageFiles,
		})
	} else {
		_, err = s.ChannelMessageSend(m.ChannelID, response.Choices[0].Message.Content)

	}

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

// TODO when we generate files, we need to handle assistnat messages falling through as user messages
func buildMessages(msg *discordgo.Message, isAssistant bool) openai.ChatCompletionMessageParamUnion {
	content := msg.Content

	// Don't format in this manner for our own bot messages otherwise replies
	// tend to include the timestamp and user name when poasting
	if !isAssistant {
		content = messages.GetContent(msg)
	}

	// Handle the no attachment scenario
	if len(msg.Attachments) == 0 {
		if isAssistant {
			return openai.AssistantMessage(content)
		}

		return openai.UserMessage(content)
	}

	// Handle messages with attachments
	// First add the text / user / timestamp
	parts := []openai.ChatCompletionContentPartUnionParam{
		openai.TextContentPart(content),
	}

	for _, att := range msg.Attachments {
		slog.Info("Attachment", "filename", att.Filename, "content_type", att.ContentType, "url", att.URL)

		if strings.HasPrefix(att.ContentType, "image/") {
			parts = append(parts, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
				URL:    att.URL,
				Detail: "auto",
			}))
		} else if strings.HasPrefix(att.ContentType, "audio/") {
			data, err := downloadFile(att.URL)
			if err != nil {
				slog.Error("Error downloading file", "name", att.Filename, "url", att.URL, "error", err.Error())
				continue
			}

			b64 := base64.StdEncoding.EncodeToString(data)
			parts = append(parts, openai.InputAudioContentPart(openai.ChatCompletionContentPartInputAudioInputAudioParam{
				Data:   b64,
				Format: "ogg", // not supported by openai SDK, but gemini supports it
			}))

		} else {
			// All other files
			parts = append(parts, openai.FileContentPart(openai.ChatCompletionContentPartFileFileParam{
				FileData: openai.String(att.URL),
				Filename: openai.String(att.Filename),
			}))
		}

	}

	return openai.UserMessage(parts)

}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

type openRouterImage struct {
	Type     string `json:"type"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

func extractImages(rawJSON string) []*discordgo.File {
	var parsed struct {
		Images []openRouterImage `json:"images"`
	}

	err := json.Unmarshal([]byte(rawJSON), &parsed)

	if err != nil {
		return nil
	}

	var files []*discordgo.File
	for i, img := range parsed.Images {
		parts := strings.SplitN(img.ImageURL.URL, ",", 2)
		if len(parts) != 2 {
			continue
		}

		data, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			slog.Error("Error decoding image", "error", err.Error())
			continue
		}

		ext := "png"
		if strings.Contains(parts[0], "image/jpeg") {
			ext = "jpeg"
		}

		files = append(files, &discordgo.File{
			Name:        fmt.Sprintf("image_%d.%s", i, ext),
			ContentType: fmt.Sprintf("image/%s", ext),
			Reader:      bytes.NewReader(data),
		})
	}

	return files
}
