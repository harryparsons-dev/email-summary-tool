package routes

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func InitializeRoutes(e *echo.Echo) {
	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "email summary server is running")
	})
}
