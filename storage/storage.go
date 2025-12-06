package storage

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

type Storage interface {
	Set(e Entry) error
	FindNearest(vector []float32, k int) ([]Entry, error)
	Get(id []byte) (Entry, error)
	Close() error
}

type Entry struct {
	Id        []byte
	Prompt    string
	Answer    string
	Embeding  []float32
	CreatedAt time.Time

	similarity float32
}

func (e *Entry) Bytes() ([]byte, error) {
	return json.Marshal(e)
}

// cosineSimilarity calculates the cosine similarity between two vectors using float32
func cosineSimilarity(a, b []float32) (float32, error) {
	// Check if vectors have the same dimension
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors must have the same dimension: len(a)=%d, len(b)=%d", len(a), len(b))
	}

	// Handle empty vectors
	if len(a) == 0 {
		return 0, fmt.Errorf("vectors cannot be empty")
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	// Calculate magnitudes
	normA = float32(math.Sqrt(float64(normA)))
	normB = float32(math.Sqrt(float64(normB)))

	// Handle zero vectors
	if normA == 0 || normB == 0 {
		return 0, fmt.Errorf("cosine similarity is undefined for zero vectors")
	}

	// Calculate and return cosine similarity
	similarity := dotProduct / (normA * normB)

	// Due to floating point errors, the result might slightly exceed 1.0
	// Clamp it to the valid range [-1.0, 1.0]
	if similarity > 1.0 {
		similarity = 1.0
	} else if similarity < -1.0 {
		similarity = -1.0
	}

	return similarity, nil
}
