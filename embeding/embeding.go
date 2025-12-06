package embeding

import "errors"

type EmbederType string

const (
	HttpEmbeder   EmbederType = "http"
	OllamaEmbeder EmbederType = "ollama"
)

var (
	ErrEmptyEmbedings = errors.New("empty embedings")
)

type Embeder interface {
	Embed(string) ([]float32, error)
}
