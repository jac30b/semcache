package main

import (
	"os"

	"github.com/goccy/go-yaml"
	"github.com/jakub-galecki/semcache/embeding"
	"github.com/jakub-galecki/semcache/llm"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var logger *zap.Logger

type config struct {
	Embeder     embeding.EmbederType
	EmbederPath string      `yaml:"embederPath"`
	Llm         llm.LLMType `yaml:"llm"`
}

func init() {
	logger = zap.Must(zap.NewDevelopment())
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	conf, err := os.ReadFile("./default.yml")
	if err != nil {
		panic(err)
	}

	var config config
	err = yaml.Unmarshal(conf, &config)
	if err != nil {
		panic(err)
	}

	logger.Info("starting application",
		zap.String("embeder", string(config.Embeder)),
		zap.String("embederPath", config.EmbederPath))

	sc := newSemcache(&config)

	srv := newEchoServer(sc)
	defer srv.stop()

	srv.start(":8000")
}
