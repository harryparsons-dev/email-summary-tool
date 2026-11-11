package routes

import (
	"mailai/server"
	projecthandlers "mailai/server/handlers/project_handlers"
	"mailai/server/middleware"
)

// InitializeRestrictedRoutes registers routes that require an authenticated user.
func InitializeRestrictedRoutes(server *server.Server) {
	projectHandler := projecthandlers.NewProjectHandler(server)
	projectEmailHandler := projecthandlers.NewProjectEmailHandler(server)
	projectEmailSummaryHandler := projecthandlers.NewProjectEmailSummaryHandler(server)

	restricted := server.E.Group("", middleware.Authenticated(server))

	restricted.GET("/projects", projectHandler.List)
	restricted.POST("/projects", projectHandler.Create)
	restricted.GET("/projects/:id", projectHandler.Get)
	restricted.PUT("/projects/:id", projectHandler.Update)
	restricted.DELETE("/projects/:id", projectHandler.Delete)
	restricted.GET("/projects/:id/emails", projectEmailHandler.List)
	restricted.GET("/projects/:id/emails/:emailId/summarise", projectEmailSummaryHandler.List)
	restricted.POST("/projects/:id/emails/:emailId/summarise", projectEmailSummaryHandler.Create)
	restricted.GET("/projects/:id/emails/:emailId/summarise/:summaryId", projectEmailSummaryHandler.Get)
}
