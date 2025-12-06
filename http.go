package main

import (
	"crypto/sha256"
	"net/http"
	"strings"
	"time"

	"github.com/jakub-galecki/semcache/storage"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type server struct {
	e *echo.Echo

	s *semcache
}

func newEchoServer(s *semcache) *server {
	e := echo.New()

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	e.GET("/query", func(c echo.Context) error {
		query := strings.TrimSpace(c.QueryParam("query"))
		if query == "" {
			return c.String(http.StatusBadRequest, "query is empty")
		}
		logger.Debug("received query from user",
			zap.String("query", query))

		// exact match
		id := genId(query)
		e, err := s.storage.Get(id)
		if err == nil && e.Answer != "" {
			return c.String(http.StatusOK, e.Answer)
		}

		// embed
		vec, err := s.embeder.Embed(query)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// nearest match
		entries, err := s.storage.FindNearest(vec, 1)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		if len(entries) > 0 {
			return c.String(http.StatusOK, entries[0].Answer)
		}

		// ask llm
		res, err := s.llm.Ask(query)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		id = genId(query)
		e = storage.Entry{
			Id:        id,
			Prompt:    query,
			Answer:    res,
			Embeding:  vec,
			CreatedAt: time.Now(),
		}

		err = s.storage.Set(e)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, e.Answer)
	})

	return &server{
		e: e,
		s: s,
	}
}

func genId(query string) []byte {
	sha256Hasher := sha256.New()
	sha256Hasher.Write([]byte(query))
	return sha256Hasher.Sum(nil)
}

func (s *server) start(port string) {
	s.e.Logger.Fatal(s.e.Start(port))
}

func (s *server) stop() {
	s.e.Close()
}
