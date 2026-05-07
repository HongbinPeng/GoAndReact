package models

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
)

func NewChatModel(ctx context.Context) (*openai.ChatModel, error) {
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		APIKey:  "sk-e4dcae9060904558a2f64d6eb12249e0",
		Model:   "qwen3.6-plus",
		ExtraFields: map[string]any{
			"enable_thinking": false,
		},
	})
	if err != nil {
		return nil, err
	}
	return chatModel, nil
}
