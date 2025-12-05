package embeding

type EmbederType string

const (
	HttpEmbeder EmbederType = "http"
)

type Embeder interface {
	Embed(string) ([]float32, error)
}
