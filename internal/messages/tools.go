package messages

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

var Tools = []openai.ChatCompletionToolUnionParam{
	openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
		Name:        "generate_image",
		Description: openai.String("Generate an image based on a text conversation"),
		Parameters: shared.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"prompt": map[string]any{
					"type":        "string",
					"description": "A detailed description of the image to generate",
				},
				"aspect_ratio": map[string]any{
					"type":        "string",
					"description": "Optional aspect ratio for the image",
					// TODO allow params
					// "enum":        []string{"1:1", "2:3", "3:2", "3:4", "4:3", "4:5", "5:4", "9:16", "16:9", "21:9", "1:4", "4:1", "1:8", "8:1"},
					"enum": []string{"1:1"},
				},
				"image_size": map[string]any{
					"type":        "string",
					"description": "Optional output size for the image",
					// TODO allow params
					// "enum":        []string{"0.5K", "1K", "2K", "4K"},
					"enum": []string{"4K"},
				},
			},
			"required": []string{"prompt", "aspect_ratio", "image_size"},
		},
	}),
	openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
		Name:        "no_response",
		Description: openai.String("Call this when the message is not directed at you, not interesting, or doesn't warrant a response from you"),
	}),
}
