package index

import (
	"time"

	"github.com/coder/hnsw"
	"github.com/jakub-galecki/semcache/storage"
	"go.uber.org/zap"
)

type Key = string // todo: change to int

type hnswIndex struct {
	logger *zap.Logger
	index  *hnsw.Graph[Key]
}

func NewHNSWStorage(logger *zap.Logger) (*hnswIndex, error) {
	return &hnswIndex{
		logger: logger.Named("hnsw"),
		index:  hnsw.NewGraph[Key](),
	}, nil
}

func (hs *hnswIndex) Add(e storage.Entry) error {
	start := time.Now()
	defer func() {
		hs.logger.Debug("index.Add",
			zap.String("type", "hnsw"),
			zap.Duration("elapsed", time.Since(start)))
	}()

	// Store the actual entry
	key := string(e.Id)
	n := hnsw.MakeNode(
		key,
		e.Embeding,
	)
	hs.index.Add(n)
	return nil
}

func (hs *hnswIndex) FindNearest(vector []float32, k int) ([]Key, error) {
	var (
		similar []Key

		start   = time.Now()
		targets = make([]hnsw.Node[Key], 0, k)
	)

	defer func() {
		hs.logger.Debug("index.FindNearest",
			zap.String("storage", "hnsw"),
			zap.Int("targets", len(targets)),
			zap.Duration("elapsed", time.Since(start)))
	}()

	targets = hs.index.Search(vector, k)
	for _, target := range targets {
		similar = append(similar, target.Key)
	}

	return similar, nil
}

func (hs *hnswIndex) Close() error {
	hs.index = nil
	return nil
}
