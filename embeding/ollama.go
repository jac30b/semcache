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
		Model: "all-minilm",
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
