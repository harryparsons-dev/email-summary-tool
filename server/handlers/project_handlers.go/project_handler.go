package projecthandlers

import (
	"email-summary-tool/server"
	"net/http"

	"github.com/labstack/echo/v5"
)

type ProjectHandler struct {
	Server *server.Server
}

func NewProjectHandler(s *server.Server) *ProjectHandler {
	return &ProjectHandler{
		Server: s,
	}
}

func (h *ProjectHandler) List(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": "List of projects",
	})
}


