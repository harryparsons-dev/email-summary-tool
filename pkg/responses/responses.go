// Package responses writes the API's standard JSON response bodies.
package responses

import "github.com/labstack/echo/v5"

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

// Error writes {"error": message} with the given status code.
func Error(c *echo.Context, status int, message string) error {
	return c.JSON(status, ErrorResponse{Error: message})
}

// Success writes {"message": message} with the given status code.
func Success(c *echo.Context, status int, message string) error {
	return c.JSON(status, SuccessResponse{Message: message})
}

// Data writes data as the JSON body with the given status code.
func Data(c *echo.Context, status int, data any) error {
	return c.JSON(status, data)
}
