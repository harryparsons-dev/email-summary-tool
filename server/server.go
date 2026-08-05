package server

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type Server struct {
	e *echo.Echo
}

func NewServer() *Server {
	e := echo.New()
	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "email summary server is running")
	})

	return &Server{e: e}
}

func (s *Server) Start(address string) error {
	return s.e.Start(address)
}
