package llm

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type openAIClient struct {
	logger *zap.Logger
	client *openai.Client
	model  string
}

func NewOpenAIClient(logger *zap.Logger, model string) (*openAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	if model == "" {
		model = openai.GPT4oMini
	}

	client := openai.NewClient(apiKey)

	return &openAIClient{
		logger: logger.Named("openai"),
		client: client,
		model:  model,
	}, nil
}

func (c *openAIClient) Ask(question string) (string, error) {
	start := time.Now()
	defer func() {
		c.logger.Debug("llm.Ask",
			zap.String("llm", "openai"),
			zap.Duration("elapsed", time.Since(start)))
	}()

	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: question,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
