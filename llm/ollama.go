package llm

import (
	"context"

	ollama "github.com/ollama/ollama/api"
	"go.uber.org/zap"
)

type ollamaClient struct {
	logger *zap.Logger
	client *ollama.Client
}

func NewOllamaClient(logger *zap.Logger) (*ollamaClient, error) {
	oc, err := ollama.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}
	c := &ollamaClient{
		logger: logger.Named("embeding"),
		client: oc,
	}
	c.logger.Debug("created llm",
		zap.String("llm", "ollama"))

	return c, nil
}

func (c *ollamaClient) Ask(question string) (string, error) {
	var response string

	c.logger.Debug("generating response to query",
		zap.String("query", question))

	// todo: we can store whole context?
	err := c.client.Generate(context.TODO(), &ollama.GenerateRequest{
		Model:  "qwen3",
		Prompt: question,
	}, func(res ollama.GenerateResponse) error {
		response += res.Response
		return nil
	})
	if err != nil {
		return "", err
	}
	return response, nil
}
