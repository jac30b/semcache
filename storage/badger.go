package storage

import (
	"encoding/json"
	"sort"

	badger "github.com/dgraph-io/badger/v4"
	"go.uber.org/zap"
)

var (
	_ Storage = (*badgerStorage)(nil)
)

type badgerStorage struct {
	logger *zap.Logger
	db     *badger.DB
}

func NewBadgerStorage(logger *zap.Logger, path string) (*badgerStorage, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	return &badgerStorage{
		db:     db,
		logger: logger,
	}, nil
}

func (bs *badgerStorage) Set(e Entry) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		raw, err := e.Bytes()
		if err != nil {
			return err
		}
		e := badger.NewEntry([]byte(e.Id), raw)
		return txn.SetEntry(e)
	})
}

func (bs *badgerStorage) Get(id []byte) (Entry, error) {
	var ent Entry
	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(id)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &ent)
		})
	})
	if err != nil {
		return ent, err
	}
	return ent, nil
}

func (bs *badgerStorage) FindNearest(vector []float32, k int) ([]Entry, error) {
	var (
		e       Entry
		similar []Entry
	)

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				err := json.Unmarshal(v, &e)
				if err != nil {
					return err
				}
				sim, err := cosineSimilarity(vector, e.Embeding)
				if err != nil {
					return err
				}
				// bs.logger.Debug("checking entry in database",
				// 	zap.String("prompt", e.Prompt),
				// 	zap.String("answer", e.Answer),
				// 	zap.Float32("similarity", sim))
				if sim >= 0.9 {
					e.similarity = sim
					similar = append(similar, e)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort by similarity descending
	sort.Slice(similar, func(i, j int) bool {
		return similar[i].similarity > similar[j].similarity
	})

	// Trim to K results
	if len(similar) > k {
		similar = similar[:k]
	}

	return similar, nil
}

func (bs *badgerStorage) Close() error {
	return bs.db.Close()
}
