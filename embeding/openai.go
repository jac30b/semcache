package embeding

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type openAIEmbeder struct {
	logger *zap.Logger
	client *openai.Client
	model  openai.EmbeddingModel
}

func NewOpenAIEmbeder(logger *zap.Logger, model string) (*openAIEmbeder, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	m := openai.SmallEmbedding3
	if model != "" {
		m = openai.EmbeddingModel(model)
	}

	return &openAIEmbeder{
		logger: logger.Named("openai"),
		client: openai.NewClient(apiKey),
		model:  m,
	}, nil
}

func (e *openAIEmbeder) Embed(query string) ([]float32, error) {
	start := time.Now()
	defer func() {
		e.logger.Debug("embeder.Embed",
			zap.String("embeder", "openai"),
			zap.Duration("elapsed", time.Since(start)))
	}()

	resp, err := e.client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{query},
			Model: e.model,
		},
	)

	if err != nil {
		return nil, err
	}

	return resp.Data[0].Embedding, nil
}
