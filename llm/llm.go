package llm

type LLMType string

const (
	Ollama LLMType = "ollama"
)

type LLM interface {
	Ask(question string) (string, error)
}
