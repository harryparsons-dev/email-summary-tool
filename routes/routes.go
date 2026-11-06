package routes

import (
	"email-summary-tool/pkg/responses"
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
	projectHandler := projecthandlers.NewProjectHandler(server)
	projectEmailHandler := projecthandlers.NewProjectEmailHandler(server)
	requireAuth := middleware.RequireAuth(server)

	server.E.GET("/", func(c *echo.Context) error {
		return responses.Success(c, http.StatusOK, "email summary server is running")
	})
	server.E.POST("/signup", authHandler.SignUp)
	server.E.POST("/login", authHandler.Login)
	server.E.POST("/logout", authHandler.Logout)
	server.E.GET("/refresh", authHandler.Refresh)
	server.E.GET("/profile", handlers.GetProfile, requireAuth)

	server.E.POST("/projects", projectHandler.Create, requireAuth)
	server.E.GET("/projects", projectHandler.List, requireAuth)
	server.E.GET("/projects/:projectId", projectHandler.Get, requireAuth)
	server.E.GET("/projects/:projectId/emails", projectEmailHandler.List, requireAuth)
}
