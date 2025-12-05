package embeding

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type httpEmbeder struct {
	logger *zap.Logger

	client *http.Client
	path   string
}

func NewHttpEmbeder(logger *zap.Logger, path string) (*httpEmbeder, error) {
	em := &httpEmbeder{
		client: http.DefaultClient,
		path:   path,
		logger: logger.Named("embeding"),
	}

	em.logger.Debug("created embeder",
		zap.String("type", "http"),
		zap.String("path", path))

	return em, nil
}

func (he *httpEmbeder) Embed(query string) ([]float32, error) {
	now := time.Now()
	defer func() {
		he.logger.Debug("sending embeding request",
			zap.String("query", query),
			zap.Duration("elapsed", time.Since(now)))
	}()

	req := map[string][]string{
		"inputs": {query},
	}

	raw, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	res, err := he.client.Post(he.path, "application/json", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New("invalid response status: " + res.Status)
	}

	var vec [][]float32
	err = json.NewDecoder(res.Body).Decode(&vec)
	if err != nil {
		return nil, err
	}

	if len(vec) == 0 {
		return nil, errors.New("empty embed response")
	}

	return vec[0], nil
}
