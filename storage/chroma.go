package storage

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/amikos-tech/chroma-go/pkg/embeddings"
	"go.uber.org/zap"
)

var _ Storage = (*chromaStorage)(nil)

type chromaStorage struct {
	logger     *zap.Logger
	client     chroma.Client
	collection chroma.Collection
}

func NewChromaStorage(logger *zap.Logger, path string) (*chromaStorage, error) {
	client, err := chroma.NewHTTPClient(
		chroma.WithBaseURL(path),
	)
	if err != nil {
		return nil, err
	}

	col, err := client.GetOrCreateCollection(context.TODO(), "semcache")
	if err != nil {
		return nil, err
	}

	return &chromaStorage{logger: logger, client: client, collection: col}, nil
}

func (cs *chromaStorage) Set(e Entry) error {
	start := time.Now()
	defer func() {
		cs.logger.Debug("storage.Set",
			zap.String("storage", "chroma"),
			zap.Duration("elapsed", time.Since(start)))
	}()

	raw, err := e.Bytes()
	if err != nil {
		return err
	}

	meta := chroma.NewDocumentMetadata(
		chroma.NewStringAttribute("data", string(raw)),
	)

	emb := embeddings.NewEmbeddingFromFloat32(e.Embeding)

	return cs.collection.Upsert(context.TODO(),
		chroma.WithIDs(chroma.DocumentID(e.Id)),
		chroma.WithTexts(e.Prompt),
		chroma.WithEmbeddings(emb),
		chroma.WithMetadatas(meta),
	)
}

func (cs *chromaStorage) Get(id []byte) (Entry, error) {
	var (
		ent   Entry
		start = time.Now()
	)
	defer func() {
		cs.logger.Debug("storage.Get",
			zap.String("storage", "chroma"),
			zap.Duration("elapsed", time.Since(start)))
	}()

	res, err := cs.collection.Get(context.TODO(), chroma.WithIDs(chroma.DocumentID(id)))
	if err != nil {
		return ent, err
	}

	if res.Count() == 0 {
		return ent, nil
	}

	metas := res.GetMetadatas()
	if len(metas) > 0 {
		if s, ok := metas[0].GetString("data"); ok {
			err = json.Unmarshal([]byte(s), &ent)
			if err != nil {
				return ent, err
			}
		}
	}

	return ent, nil
}

func (cs *chromaStorage) FindNearest(vector []float32, k int) ([]Entry, error) {
	var (
		entries []Entry
		start   = time.Now()
	)
	defer func() {
		cs.logger.Debug("storage.FindNearest",
			zap.String("storage", "chroma"),
			zap.Int("found", len(entries)),
			zap.Duration("elapsed", time.Since(start)))
	}()

	emb := embeddings.NewEmbeddingFromFloat32(vector)

	res, err := cs.collection.Query(context.TODO(),
		chroma.WithQueryEmbeddings(emb),
		chroma.WithNResults(k),
		chroma.WithInclude(chroma.IncludeMetadatas, chroma.IncludeDistances),
	)
	if err != nil {
		return nil, err
	}

	if res.CountGroups() == 0 {
		return entries, nil
	}

	metas := res.GetMetadatasGroups()[0]
	dists := res.GetDistancesGroups()[0]

	for i := range metas {
		if s, ok := metas[i].GetString("data"); ok {
			var ent Entry
			err = json.Unmarshal([]byte(s), &ent)
			if err != nil {
				return nil, err
			}
			if i < len(dists) {
				ent.similarity = float32(1.0 - dists[i])
			}
			entries = append(entries, ent)
		}
	}

	return entries, nil
}

func (cs *chromaStorage) Close() error {
	var err error
	err = errors.Join(err, cs.collection.Close())
	err = errors.Join(err, cs.client.Close())
	return err
}
