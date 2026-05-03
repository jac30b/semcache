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
	model  string
}

func NewOllamaEmbeder(logger *zap.Logger, model string) (*ollamaEmbeder, error) {
	ollamaClient, err := ollama.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	if model == "" {
		model = "nomic-embed-text"
	}

	c := &ollamaEmbeder{
		logger: logger.Named("embeding"),
		client: ollamaClient,
		model:  model,
	}
	c.logger.Debug("created embeder",
		zap.String("embeder", "ollama"),
		zap.String("model", model))
	return c, nil
}

func (e *ollamaEmbeder) Embed(query string) ([]float32, error) {
	var (
		dimensions int

		start = time.Now()
	)

	defer func() {
		e.logger.Debug("embeder.Embed",
			zap.String("embeder", "ollama"),
			zap.Int("dimensions", dimensions),
			zap.Duration("elapsed", time.Since(start)))
	}()

	res, err := e.client.Embed(context.TODO(), &ollama.EmbedRequest{
		Model: e.model,
		Input: []string{query},
		// todo: dimensions
	})

	if err != nil {
		return nil, err
	}

	if len(res.Embeddings) == 0 {
		return nil, ErrEmptyEmbedings
	}

	dimensions = len(res.Embeddings[0])

	return res.Embeddings[0], nil
}
