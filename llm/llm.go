package llm

type LLMType string

const (
	Ollama LLMType = "ollama"
	OpenAI LLMType = "openai"
	Groq   LLMType = "groq"
)

type LLM interface {
	Ask(question string) (string, error)
}
