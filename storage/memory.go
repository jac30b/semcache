package storage

import (
	"container/list"
	"errors"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Ensure memoryStorage implements Storage interface
var _ Storage = (*memoryStorage)(nil)

type lruItem struct {
	key   string
	entry Entry
}

type memoryStorage struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	data    map[string]*list.Element
	lruList *list.List
	maxSize int
}

// NewMemoryStorage creates a new in-memory storage with LRU eviction
func NewMemoryStorage(logger *zap.Logger) *memoryStorage {
	ms := &memoryStorage{
		logger:  logger.Named("memory"),
		data:    make(map[string]*list.Element),
		lruList: list.New(),
		maxSize: 100, // Maximum capacity of 100 entries
	}

	ms.logger.Debug("created storage",
		zap.String("storage", "memory"),
		zap.Int("maxSize", ms.maxSize))

	return ms
}

// Set stores an entry in memory with LRU tracking
func (ms *memoryStorage) Set(e Entry) error {
	start := time.Now()
	defer func() {
		ms.logger.Debug("storage.Set",
			zap.String("storage", "memory"),
			zap.Duration("elapsed", time.Since(start)))
	}()

	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := string(e.Id)

	// If entry already exists, remove it from current position in LRU list
	if elem, exists := ms.data[key]; exists {
		ms.lruList.Remove(elem)
		delete(ms.data, key)
	} else if ms.lruList.Len() >= ms.maxSize {
		// If we're at capacity, remove the least recently used item
		oldest := ms.lruList.Back()
		if oldest != nil {
			oldestItem := oldest.Value.(*lruItem)
			ms.logger.Debug("evicting LRU entry",
				zap.String("key", oldestItem.key))
			ms.lruList.Remove(oldest)
			delete(ms.data, oldestItem.key)
		}
	}

	// Add new entry to the front of the list (most recently used)
	item := &lruItem{key: key, entry: e}
	elem := ms.lruList.PushFront(item)
	ms.data[key] = elem

	return nil
}

// Get retrieves an entry by ID and updates LRU status
func (ms *memoryStorage) Get(id []byte) (Entry, error) {
	var ent Entry

	start := time.Now()
	defer func() {
		ms.logger.Debug("storage.Get",
			zap.String("storage", "memory"),
			zap.Duration("elapsed", time.Since(start)))
	}()

	key := string(id)

	ms.mu.Lock() // Need write lock to update LRU order
	defer ms.mu.Unlock()

	elem, exists := ms.data[key]
	if !exists {
		return ent, errors.New("not found")
	}

	// Move to front of LRU list (mark as most recently used)
	ms.lruList.MoveToFront(elem)
	item := elem.Value.(*lruItem)

	return item.entry, nil
}

// FindNearest finds entries with embeddings similar to the given vector
func (ms *memoryStorage) FindNearest(vector []float32, k int) ([]Entry, error) {
	var similar []Entry

	start := time.Now()
	defer func() {
		ms.logger.Debug("storage.FindNearest",
			zap.String("storage", "memory"),
			zap.Duration("elapsed", time.Since(start)))
	}()

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Calculate similarity for all entries
	for _, elem := range ms.data {
		item := elem.Value.(*lruItem)
		e := item.entry

		sim, err := cosineSimilarity(vector, e.Embeding)
		if err != nil {
			continue
		}

		if sim >= 0.9 {
			e.Similarity = sim
			similar = append(similar, e)
		}
	}

	// Sort by similarity descending
	sort.Slice(similar, func(i, j int) bool {
		return similar[i].Similarity > similar[j].Similarity
	})

	// Return k results or all if less than k
	if len(similar) > k {
		return similar[:k], nil
	}
	return similar, nil
}

// Close cleans up resources
func (ms *memoryStorage) Close() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Clear the map and list to release memory
	ms.data = make(map[string]*list.Element)
	ms.lruList.Init() // Reset the list
	return nil
}

// Size returns the current number of entries in the cache
func (ms *memoryStorage) Size() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.data)
}
