package main

import (
	"errors"

	"github.com/jakub-galecki/semcache/embeding"
	"github.com/jakub-galecki/semcache/llm"
	"github.com/jakub-galecki/semcache/storage"
)

type semcache struct {
	embeder   embeding.Embeder
	storage   storage.Storage
	llm       llm.LLM
	threshold float32
}

func newSemcache(config *config) *semcache {
	em, err := newEmbeder(config)
	if err != nil {
		panic(err)
	}

	llm, err := newLLM(config)
	if err != nil {
		panic(err)
	}

	storage, err := storage.NewChromaStorage(logger.Named("chroma"), "http://localhost:8900")
	if err != nil {
		panic(err)
	}

	return &semcache{
		embeder:   em,
		llm:       llm,
		storage:   storage,
		threshold: config.SimilarityThreshold,
	}
}

func newEmbeder(config *config) (embeding.Embeder, error) {
	switch config.Embeder {
	case embeding.HttpEmbeder:
		return embeding.NewHttpEmbeder(logger, config.EmbederPath)
	case embeding.OllamaEmbeder:
		return embeding.NewOllamaEmbeder(logger, config.EmbederModel)
	case embeding.OpenAIEmbeder:
		return embeding.NewOpenAIEmbeder(logger, config.EmbederModel)
	case embeding.VoyageEmbeder:
		return embeding.NewVoyageEmbeder(logger, config.EmbederModel)
	default:
		return nil, errors.New("embeder not exist")
	}
}

func newLLM(config *config) (llm.LLM, error) {
	switch config.Llm {
	case llm.Ollama:
		return llm.NewOllamaClient(logger, config.LlmModel)
	case llm.OpenAI:
		return llm.NewOpenAIClient(logger, config.LlmModel)
	case llm.Groq:
		return llm.NewGroqClient(logger, config.LlmModel)
	default:
		return nil, errors.New("llm not exist")
	}
}
