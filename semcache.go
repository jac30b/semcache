package main

import (
	"errors"

	"github.com/jakub-galecki/semcache/embeding"
)

type semcache struct {
	embeder embeding.Embeder
}

func newSemcache(config *config) *semcache {
	em, err := newEmbeder(config)
	if err != nil {
		panic(err)
	}

	return &semcache{
		embeder: em,
	}
}

func newEmbeder(config *config) (embeding.Embeder, error) {
	switch config.Embeder {
	case embeding.HttpEmbeder:
		return embeding.NewHttpEmbeder(logger, config.EmbederPath)
	default:
		return nil, errors.New("embeder not exist")
	}
}
