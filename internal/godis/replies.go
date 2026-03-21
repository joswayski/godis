package godis

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/messages"
	"github.com/openai/openai-go/v3"
)

const maxBufferedMessages = 5

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

	g.mutex.Lock()
	defer g.mutex.Unlock()
	// Check if we have a pending batch of messages to send
	buf, exists := g.MessageBuffer[m.ChannelID]
	if !exists {
		buf = MessageBuffer{
			BufferedMessages: []*discordgo.Message{m.Message},
		}
		buf.Timer = time.AfterFunc(1500*time.Millisecond, func() {
			g.processBufferedMessages(s, m.ChannelID)
		})
	} else if len(buf.BufferedMessages) < maxBufferedMessages {
		buf.BufferedMessages = append(buf.BufferedMessages, m.Message)
		buf.Timer.Reset(1500 * time.Millisecond)
	} else {
		// buffer is full
		buf.Timer.Stop()
		// Add the one that just came in
		buf.BufferedMessages = append(buf.BufferedMessages, m.Message)
		g.MessageBuffer[m.ChannelID] = buf
		go g.processBufferedMessages(s, m.ChannelID)
		return
	}

	// Write the appended message back to the global buffer
	g.MessageBuffer[m.ChannelID] = buf
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

		} else if strings.HasPrefix(att.ContentType, "video/") {
			// ignore for now, most models dont support it
			continue
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

func (g *Godis) processBufferedMessages(s *discordgo.Session, channelId string) {
	// Check if we need to refetch for any embeds on any bufferred messages
	// Sometimes messages don't have embeds populated right away
	g.mutex.Lock()
	bufferedMessages := g.MessageBuffer[channelId].BufferedMessages
	// Delete all the buffered messages so the timer starts again
	delete(g.MessageBuffer, channelId)
	g.mutex.Unlock()

	// if any messages have links in them with no embeds
	needsRefetch := false
	if slices.ContainsFunc(bufferedMessages, shouldRefetchForEmbeds) {
		needsRefetch = true
	}

	if needsRefetch {
		// Wait once for all to populate embeds
		time.Sleep(1500 * time.Millisecond)
		for i, msg := range bufferedMessages {
			if !shouldRefetchForEmbeds(msg) {
				continue
			}

			refetchedMessage, err := s.ChannelMessage(channelId, msg.ID)
			if err != nil {
				// Keep the message as is without the embeds
				slog.Error("Error refetching the channel message while retrieving embeds", "error", err.Error())
			} else {
				bufferedMessages[i] = refetchedMessage
			}
		}
	}

	aiParams := openai.ChatCompletionNewParams{
		// https://openrouter.ai/docs/quickstart
		Model:               g.Config.AIApiModels[0], // TODO allow fallbacks / retries,
		MaxCompletionTokens: openai.Int(int64(g.Config.AIMaxOutputTokens)),
	}

	var aiMessagesToSend []openai.ChatCompletionMessageParamUnion

	// Now we start adding the messages, first the system prompt
	aiMessagesToSend = append(aiMessagesToSend, openai.SystemMessage(g.Config.AISystemPrompt))

	// Now we should add some context of message history
	// We need to set the correct message on the history look back, these should already have embeds
	// as they are older than our buffer
	historyCount := max(0, g.Config.AINumberOfMessagesInHistory-len(bufferedMessages))
	if historyCount > 0 {
		// Give it context of the messages before our first bufferred one (oldest), capped at what the user put in
		history, err := s.ChannelMessages(channelId, historyCount, bufferedMessages[0].ID, "", "")

		if err != nil {
			slog.Error("Error fetching channel history", "error", err.Error())
			// This is fine, we just wont add any history
		}

		// Messages in discord history come newest first, so reverse it and add it to the AI call
		for i := len(history) - 1; i >= 0; i-- {
			msg := history[i]
			isAssistant := msg.Author.ID == s.State.User.ID // our own messages
			aiMessagesToSend = append(aiMessagesToSend, buildMessages(msg, isAssistant))
		}
	}

	// Add our bufferred messages at the end, this includes our current message
	for _, msg := range bufferedMessages {
		isAssistant := msg.Author.ID == s.State.User.ID // our own messages
		aiMessagesToSend = append(aiMessagesToSend, buildMessages(msg, isAssistant))
	}

	// Set the messages and send it
	aiParams.Messages = aiMessagesToSend
	response, err := g.AIClient.Chat.Completions.New(context.TODO(), aiParams)

	if err != nil {
		slog.Error("Error ocurred generating response", "error", err.Error(), "message", bufferedMessages[len(bufferedMessages)-1])
		return
		// TODO dependng on if the model failed or whatever it might be good to send
		// an appropriate response. right now just log
	}

	if len(response.Choices) == 0 || strings.Contains(strings.ToLower(response.Choices[0].Message.Content), "no_response") {
		// TODO use tools instead
		return
	}

	// Add channel typing indicator
	s.ChannelTyping(channelId)
	jitter := time.Duration(rand.IntN(2000)+100) * time.Millisecond
	time.Sleep(jitter)

	_, err = s.ChannelMessageSend(channelId, response.Choices[0].Message.Content)

	if err != nil {
		slog.Error("Error sending discord message", "response", response, "error", err.Error())
	}
}
