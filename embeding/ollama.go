package embeding

import (
	"context"
	"time"

	ollama "github.com/ollama/ollama/api"
	"go.uber.org/zap"
)

var (
	_ Embeder = (*ollamaEmbeder)(nil)
)

type ollamaEmbeder struct {
	logger *zap.Logger
	client *ollama.Client
}

func NewOllamaEmbeder(logger *zap.Logger) (*ollamaEmbeder, error) {
	ollamaClient, err := ollama.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}
	c := &ollamaEmbeder{
		logger: logger.Named("embeding"),
		client: ollamaClient,
	}
	c.logger.Debug("created embeder",
		zap.String("embeder", "ollama"))
	return c, nil
}

func (e *ollamaEmbeder) Embed(query string) ([]float32, error) {
	now := time.Now()

	defer func() {
		e.logger.Debug("sending embeding request",
			zap.String("embeder", "ollama"),
			zap.String("query", query),
			zap.Duration("elapsed", time.Since(now)))
	}()

	res, err := e.client.Embed(context.TODO(), &ollama.EmbedRequest{
		Model: "embeddinggemma",
		Input: []string{query},
		// todo: dimensions
	})

	if err != nil {
		return nil, err
	}

	if len(res.Embeddings) == 0 {
		return nil, ErrEmptyEmbedings
	}

	return res.Embeddings[0], nil
}
