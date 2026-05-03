package embeding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

type voyageEmbeder struct {
	logger *zap.Logger
	client *http.Client
	apiKey string
	model  string
}

func NewVoyageEmbeder(logger *zap.Logger, model string) (*voyageEmbeder, error) {
	apiKey := os.Getenv("VOYAGE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("VOYAGE_API_KEY not set")
	}

	if model == "" {
		model = "voyage-3"
	}

	return &voyageEmbeder{
		logger: logger.Named("voyage"),
		client: &http.Client{Timeout: 30 * time.Second},
		apiKey: apiKey,
		model:  model,
	}, nil
}

type voyageRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type voyageResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (e *voyageEmbeder) Embed(query string) ([]float32, error) {
	start := time.Now()
	defer func() {
		e.logger.Debug("embeder.Embed",
			zap.String("embeder", "voyage"),
			zap.Duration("elapsed", time.Since(start)))
	}()

	reqBody, _ := json.Marshal(voyageRequest{
		Input: []string{query},
		Model: e.model,
	})

	req, err := http.NewRequest("POST", "https://api.voyageai.com/v1/embeddings", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("voyage api error: %s", resp.Status)
	}

	var vResp voyageResponse
	if err := json.NewDecoder(resp.Body).Decode(&vResp); err != nil {
		return nil, err
	}

	if len(vResp.Data) == 0 {
		return nil, ErrEmptyEmbedings
	}

	return vResp.Data[0].Embedding, nil
}
