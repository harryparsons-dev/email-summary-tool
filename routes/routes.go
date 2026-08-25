package routes

import (
	"email-summary-tool/server"
	"email-summary-tool/server/handlers"
	projecthandlers "email-summary-tool/server/handlers/project_handlers.go"
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

	// Protected routes
	projectHandler := projecthandlers.NewProjectHandler(server)

	server.E.GET("/profile", handlers.GetProfile, middleware.RequireAuth(server))

	server.E.GET("/projects", projectHandler.List, middleware.RequireAuth(server))
	server.E.POST("/projects", projectHandler.Create, middleware.RequireAuth(server))
	server.E.GET("/projects/:id", projectHandler.Get, middleware.RequireAuth(server))
	server.E.PUT("/projects/:id", projectHandler.Update, middleware.RequireAuth(server))
	server.E.DELETE("/projects/:id", projectHandler.Delete, middleware.RequireAuth(server))

}
