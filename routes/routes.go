package routes

import (
	"email-summary-tool/server"
	"email-summary-tool/server/handlers"
	"net/http"

	"github.com/labstack/echo/v5"
)

func InitializeRoutes(server *server.Server) {

	authHandler := handlers.NewAuthHandler(server)

	server.E.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "email summary server is running")
	})
	server.E.POST("/signup", authHandler.SignUp)
	server.E.POST("/login", authHandler.Login)
	server.E.GET("/refresh", authHandler.Refresh)
}
