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
			},
			"required": []string{"prompt"},
		},
	}),
	openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
		Name:        "no_response",
		Description: openai.String("Call this when the message doesn't warrant a response from you"),
	}),
}
