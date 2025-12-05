package storage

import (
	badger "github.com/dgraph-io/badger/v4"
)

type badgerStorage struct {
	db *badger.DB
}

func NewBadgerStorage(path string) (*badgerStorage, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	return &badgerStorage{
		db: db,
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

func (bs *badgerStorage) FindNearest(vector []float32, k int) ([]Entry, error) {
	return nil, nil
}

func (bs *badgerStorage) Close() error {
	return bs.db.Close()
}
