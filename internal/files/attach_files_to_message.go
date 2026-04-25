package files

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/openai/openai-go/v3"
)

func AttachFilesToMessage(aiMessageParts []openai.ChatCompletionContentPartUnionParam, msg *discordgo.Message) []openai.ChatCompletionContentPartUnionParam {
	// Attach our own attachments
	for _, att := range msg.Attachments {
		aiMessageParts = attachFile(aiMessageParts, att, "Current message")
	}

	// Attach any reply attachments
	if msg.ReferencedMessage != nil {
		for _, att := range msg.ReferencedMessage.Attachments {
			aiMessageParts = attachFile(aiMessageParts, att, "Referenced message")
		}
	}

	return aiMessageParts

}

// Source is if the attachment is for the current message or a reply
func attachmentLabel(att *discordgo.MessageAttachment, source string) string {
	return fmt.Sprintf("%s attachment: %s (%s)", source, att.Filename, att.ContentType)
}

func attachFile(aiMessageParts []openai.ChatCompletionContentPartUnionParam, att *discordgo.MessageAttachment, source string) []openai.ChatCompletionContentPartUnionParam {
	slog.Info("Attachment", "filename", att.Filename, "content_type", att.ContentType, "url", att.URL)

	if strings.HasPrefix(att.ContentType, "image/") {
		aiMessageParts = append(aiMessageParts, openai.TextContentPart(attachmentLabel(att, source)))
		aiMessageParts = append(aiMessageParts, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL:    att.URL,
			Detail: "auto",
		}))

		return aiMessageParts
	}

	if strings.HasPrefix(att.ContentType, "audio/") {
		data, err := DownloadFile(att.URL)
		if err != nil {
			slog.Error("Error downloading file", "name", att.Filename, "url", att.URL, "error", err.Error())
			return aiMessageParts
		}

		b64 := base64.StdEncoding.EncodeToString(data)
		aiMessageParts = append(aiMessageParts, openai.TextContentPart(attachmentLabel(att, source)))
		aiMessageParts = append(aiMessageParts, openai.InputAudioContentPart(openai.ChatCompletionContentPartInputAudioInputAudioParam{
			Data:   b64,
			Format: "ogg", // not supported by openai SDK, but gemini supports it
		}))

		return aiMessageParts

	}

	if strings.HasPrefix(att.ContentType, "video/") {
		// ignore for now, most models dont support it
		return aiMessageParts
	}

	// All other files
	aiMessageParts = append(aiMessageParts, openai.TextContentPart(attachmentLabel(att, source)))
	aiMessageParts = append(aiMessageParts, openai.FileContentPart(openai.ChatCompletionContentPartFileFileParam{
		FileData: openai.String(att.URL),
		Filename: openai.String(att.Filename),
	}))

	return aiMessageParts

}
