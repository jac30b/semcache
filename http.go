package main

import (
	"net/http"
	"strings"

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

		vec, err := s.embeder.Embed(query)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, vec)
	})

	return &server{
		e: e,
		s: s,
	}
}

func (s *server) start(port string) {
	s.e.Logger.Fatal(s.e.Start(port))
}

func (s *server) stop() {
	s.e.Close()
}
