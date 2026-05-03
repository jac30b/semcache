package llm

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type groqClient struct {
	logger *zap.Logger
	client *openai.Client
	model  string
}

func NewGroqClient(logger *zap.Logger, model string) (*groqClient, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY not set")
	}

	if model == "" {
		model = "llama-3.3-70b-versatile"
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"
	client := openai.NewClientWithConfig(config)

	return &groqClient{
		logger: logger.Named("groq"),
		client: client,
		model:  model,
	}, nil
}

func (c *groqClient) Ask(question string) (string, error) {
	start := time.Now()
	defer func() {
		c.logger.Debug("llm.Ask",
			zap.String("llm", "groq"),
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
