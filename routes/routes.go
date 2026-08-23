package routes

import (
	"email-summary-tool/server"
	"email-summary-tool/server/handlers"
	"email-summary-tool/server/middleware"
	tokenservice "email-summary-tool/services/tokenService"
	"net/http"

	"github.com/labstack/echo/v5"
)

func InitializeRoutes(server *server.Server) {

	tokenService := tokenservice.NewTokenService(server.JWT)
	authHandler := handlers.NewAuthHandler(server, tokenService)

	server.E.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "email summary server is running")
	})
	server.E.POST("/signup", authHandler.SignUp)
	server.E.POST("/login", authHandler.Login)
	server.E.POST("/logout", authHandler.Logout)
	server.E.GET("/refresh", authHandler.Refresh)
	server.E.GET("/profile", handlers.GetProfile, middleware.RequireAuth(server))
}
