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
	model  string
}

func NewOllamaClient(logger *zap.Logger, model string) (*ollamaClient, error) {
	oc, err := ollama.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	if model == "" {
		model = "phi3:mini"
	}

	c := &ollamaClient{
		logger: logger.Named("llm"),
		client: oc,
		model:  model,
	}
	c.logger.Debug("created llm",
		zap.String("llm", "ollama"),
		zap.String("model", model))

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
		Model:  c.model,
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
