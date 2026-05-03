# semcache
A minimal semantic cache prototype for LLMs

## Configuration
The application uses `default.yml` for configuration.

```yaml
embeder: "ollama" # ollama, http, openai, voyage
embederPath: "http://localhost:8080/embed"
similarityThreshold: 0.7
embederModel: "nomic-embed-text"
llm: "ollama" # ollama, openai, groq
llmModel: "phi3:mini"
```

### Fast APIs
For faster performance, you can use OpenAI or Groq.

Set the following environment variables:
- `OPENAI_API_KEY` for OpenAI
- `GROQ_API_KEY` for Groq
- `VOYAGE_API_KEY` for Voyage AI

Then update `default.yml`:
```yaml
llm: "openai"
llmModel: "gpt-4o-mini"
embeder: "openai"
embederModel: "text-embedding-3-small"
```
or
```yaml
llm: "groq"
llmModel: "llama-3.3-70b-versatile"
embeder: "voyage"
embederModel: "voyage-3"
```

### Faster/Better Local Models
If you prefer running locally with Ollama:
- **LLM:** `llmModel: "qwen2:0.5b"` (Faster), `llmModel: "tinyllama"` (Faster)
- **Embedder:** `embederModel: "nomic-embed-text"` (Better than all-minilm), `embederModel: "mxbai-embed-large"` (Top-tier accuracy)
