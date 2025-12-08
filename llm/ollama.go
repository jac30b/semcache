package llm

import (
	"context"
	"time"

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

	start := time.Now()
	defer func() {
		c.logger.Debug("llm.Ask",
			zap.String("llm", "ollama"),
			zap.Duration("elapsed", time.Since(start)))
	}()

	c.logger.Debug("generating response to query",
		zap.String("query", question))

	// todo: we can store whole context?
	err := c.client.Generate(context.TODO(), &ollama.GenerateRequest{
		Model:  "phi3:mini",
		Prompt: question,
		Think:  &ollama.ThinkValue{Value: false},
	}, func(res ollama.GenerateResponse) error {
		response += res.Response
		return nil
	})
	if err != nil {
		return "", err
	}
	return response, nil
}
