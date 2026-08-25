package handlers

import (
	"database/sql"
	"email-summary-tool/database/models"
	"email-summary-tool/server"
	tokenservice "email-summary-tool/services/tokenService"
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	server       *server.Server
	tokenService *tokenservice.TokenService
}

func NewAuthHandler(server *server.Server, tokenService *tokenservice.TokenService) *AuthHandler {
	return &AuthHandler{
		server:       server,
		tokenService: tokenService,
	}
}

func (h *AuthHandler) SignUp(c *echo.Context) error {
	var request struct {
		Email    string `json:"email" form:"email" validate:"required"`
		Password string `json:"password" form:"password" validate:"required"`
	}
	if err := c.Bind(&request); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request payload")
	}
	if err := c.Validate(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email and password are required",
		})
	}

	existingUser := &models.User{}
	h.server.Db.NewSelect().Model(&models.User{}).Where("email = ?", request.Email).Scan(c.Request().Context(), existingUser)
	if existingUser.ID != "" {
		return c.String(http.StatusBadRequest, "User with this email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error signing up user")
	}

	user := &models.User{
		Email:        request.Email,
		PasswordHash: string(hash),
	}
	if _, err := h.server.Db.NewInsert().Model(user).Exec(c.Request().Context()); err != nil {
		return c.String(http.StatusInternalServerError, "Error signing up user")
	}

	return c.String(http.StatusCreated, "Signed up successfully")
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var request struct {
		Email    string `json:"email" form:"email" validate:"required"`
		Password string `json:"password" form:"password" validate:"required"`
	}
	if err := c.Bind(&request); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request payload")
	}
	if err := c.Validate(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email and password are required",
		})
	}

	user := &models.User{}
	h.server.Db.NewSelect().Model(&models.User{}).Where("email = ?", request.Email).Scan(c.Request().Context(), user)
	if user.ID == "" {
		return c.String(http.StatusUnauthorized, "Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		return c.String(http.StatusUnauthorized, "Invalid email or password")
	}

	accessToken, err := h.server.JWT.SignAccessToken(user.ID)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "Error logining in user")
	}
	refreshToken, err := h.server.JWT.SignRefreshToken(user.ID)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "Error logining in user")
	}

	h.tokenService.SetAccessCookie(c, accessToken)
	h.tokenService.SetRefreshCookie(c, refreshToken)

	return c.String(http.StatusOK, "Logined in successfully")
}

func (h *AuthHandler) Refresh(c *echo.Context) error {
	cookie, err := c.Request().Cookie("RefreshToken")
	if err != nil {
		return c.String(http.StatusUnauthorized, "Refresh token required")
	}

	userID, err := h.server.JWT.VerifyRefreshToken(cookie.Value)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid or expired refresh token",
		})
	}

	user := &models.User{}
	err = h.server.Db.NewSelect().Model(user).Where("id = ?", userID).Scan(c.Request().Context())
	if errors.Is(err, sql.ErrNoRows) {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User not found",
		})
	}
	if err != nil {
		log.Printf("Error looking up user while refreshing token: %v", err)
		return c.String(http.StatusInternalServerError, "Error refreshing token")
	}

	accessToken, err := h.server.JWT.SignAccessToken(user.ID)
	if err != nil {
		log.Printf("Error signing refreshed token: %v", err)
		return c.String(http.StatusInternalServerError, "Error refreshing token")
	}

	h.tokenService.SetAccessCookie(c, accessToken)

	return c.String(http.StatusOK, "Token refreshed successfully")
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	h.tokenService.ClearAuthCookies(c)

	return c.String(http.StatusOK, "Signed out successfully")
}
