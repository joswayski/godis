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
		aiMessageParts = addAttachment(aiMessageParts, att, "Current message")
	}

	// Attach any reply attachments
	if msg.ReferencedMessage != nil {
		for _, att := range msg.ReferencedMessage.Attachments {
			aiMessageParts = addAttachment(aiMessageParts, att, "Referenced message")
		}
	}

	// Attach our own embeds
	for _, emb := range msg.Embeds {
		aiMessageParts = addEmbed(aiMessageParts, emb, "Current message")
	}

	// Attach any reply embeds
	if msg.ReferencedMessage != nil {
		for _, emb := range msg.ReferencedMessage.Embeds {
			aiMessageParts = addEmbed(aiMessageParts, emb, "Referenced message")
		}
	}

	return aiMessageParts

}

// Source is if the file is for the current message or a reply
func getAttachmentLabel(att *discordgo.MessageAttachment, source string) string {
	return fmt.Sprintf("%s file: %s (%s)", source, att.Filename, att.ContentType)
}

func addAttachment(aiMessageParts []openai.ChatCompletionContentPartUnionParam, att *discordgo.MessageAttachment, source string) []openai.ChatCompletionContentPartUnionParam {
	slog.Info("Attachment", "filename", att.Filename, "content_type", att.ContentType, "url", att.URL)
	contentType := normalizeContentType(att.ContentType)

	if strings.HasPrefix(contentType, "image/") {
		if !isSupportedImageContentType(contentType) {
			slog.Warn("Unsupported image content type", "name", att.Filename, "content_type", att.ContentType)
			return aiMessageParts
		}

		aiMessageParts = append(aiMessageParts, openai.TextContentPart(getAttachmentLabel(att, source)))
		aiMessageParts = append(aiMessageParts, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL:    att.URL,
			Detail: "auto",
		}))

		return aiMessageParts
	}

	if strings.HasPrefix(contentType, "audio/") {
		audioFormat, ok := getAudioFormat(contentType)
		if !ok {
			slog.Warn("Unsupported audio content type", "name", att.Filename, "content_type", att.ContentType)
			return aiMessageParts
		}

		data, err := DownloadFile(att.URL)
		if err != nil {
			slog.Error("Error downloading file", "name", att.Filename, "url", att.URL, "error", err.Error())
			return aiMessageParts
		}

		b64 := base64.StdEncoding.EncodeToString(data)
		aiMessageParts = append(aiMessageParts, openai.TextContentPart(getAttachmentLabel(att, source)))
		aiMessageParts = append(aiMessageParts, openai.InputAudioContentPart(openai.ChatCompletionContentPartInputAudioInputAudioParam{
			Data:   b64,
			Format: audioFormat,
		}))

		return aiMessageParts

	}

	if strings.HasPrefix(contentType, "video/") {
		// ignore for now, most models dont support it
		return aiMessageParts
	}

	if contentType != "application/pdf" {
		slog.Warn("Unsupported file content type", "name", att.Filename, "content_type", att.ContentType)
		return aiMessageParts
	}

	aiMessageParts = append(aiMessageParts, openai.TextContentPart(getAttachmentLabel(att, source)))
	aiMessageParts = append(aiMessageParts, openai.FileContentPart(openai.ChatCompletionContentPartFileFileParam{
		FileData: openai.String(att.URL),
		Filename: openai.String(att.Filename),
	}))

	return aiMessageParts

}

func normalizeContentType(contentType string) string {
	return strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
}

func isSupportedImageContentType(contentType string) bool {
	switch normalizeContentType(contentType) {
	case "image/png", "image/jpeg", "image/webp", "image/gif":
		return true
	default:
		return false
	}
}

// Returns the audio format expected by the model provider, if we can infer one.
func getAudioFormat(contentType string) (string, bool) {
	contentType = normalizeContentType(contentType)

	switch contentType {
	case "audio/mpeg", "audio/mp3", "audio/mpeg3", "audio/x-mpeg-3":
		return "mp3", true
	case "audio/wav", "audio/wave", "audio/x-wav", "audio/vnd.wave":
		return "wav", true
	case "audio/aiff", "audio/x-aiff":
		return "aiff", true
	case "audio/aac", "audio/x-aac":
		return "aac", true
	case "audio/ogg", "audio/opus", "application/ogg":
		return "ogg", true
	case "audio/flac", "audio/x-flac":
		return "flac", true
	case "audio/mp4", "audio/x-m4a", "audio/m4a":
		return "m4a", true
	case "audio/l16", "audio/pcm", "audio/x-pcm":
		return "pcm16", true
	default:
		return "", false
	}
}

func addEmbed(aiMessageParts []openai.ChatCompletionContentPartUnionParam, emb *discordgo.MessageEmbed, source string) []openai.ChatCompletionContentPartUnionParam {

	if emb.Image != nil && emb.Image.URL != "" {
		aiMessageParts = append(aiMessageParts,
			openai.TextContentPart(fmt.Sprintf("%s embed image", source)),
			openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
				URL:    emb.Image.URL,
				Detail: "auto",
			}),
		)
	}

	if emb.Thumbnail != nil && emb.Thumbnail.URL != "" {
		aiMessageParts = append(aiMessageParts,
			openai.TextContentPart(fmt.Sprintf("%s embed thumbnail", source)),
			openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
				URL:    emb.Thumbnail.URL,
				Detail: "auto",
			}),
		)
	}

	return aiMessageParts
}
