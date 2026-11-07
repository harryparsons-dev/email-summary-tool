package handlers

import (
	"database/sql"
	"email-summary-tool/database/models"
	"email-summary-tool/pkg/responses"
	"email-summary-tool/server"
	userresponses "email-summary-tool/server/responses"
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

func NewAuthHandler(server *server.Server) *AuthHandler {
	return &AuthHandler{
		server:       server,
		tokenService: tokenservice.NewTokenService(server.JWT),
	}
}

func (h *AuthHandler) SignUp(c *echo.Context) error {
	ctx := c.Request().Context()
	var request struct {
		Email    string `json:"email" form:"email" validate:"required"`
		Password string `json:"password" form:"password" validate:"required"`
	}
	if err := c.Bind(&request); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Invalid request payload")
	}
	if err := c.Validate(&request); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Email and password are required")
	}

	existingUser := &models.User{}
	h.server.Db.NewSelect().Model(&models.User{}).Where("email = ?", request.Email).Scan(ctx, existingUser)
	if existingUser.ID != "" {
		return responses.Error(c, http.StatusBadRequest, "User with this email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Error signing up user")
	}

	user := &models.User{
		Email:        request.Email,
		PasswordHash: string(hash),
	}
	if _, err := h.server.Db.NewInsert().Model(user).Exec(ctx); err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Error signing up user")
	}

	return responses.Success(c, http.StatusCreated, "Signed up successfully")
}

func (h *AuthHandler) Login(c *echo.Context) error {
	ctx := c.Request().Context()
	var request struct {
		Email    string `json:"email" form:"email" validate:"required"`
		Password string `json:"password" form:"password" validate:"required"`
	}
	if err := c.Bind(&request); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Invalid request payload")
	}
	if err := c.Validate(&request); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Email and password are required")
	}

	user := &models.User{}
	h.server.Db.NewSelect().Model(&models.User{}).Where("email = ?", request.Email).Scan(ctx, user)
	if user.ID == "" {
		return responses.Error(c, http.StatusUnauthorized, "Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		return responses.Error(c, http.StatusUnauthorized, "Invalid email or password")
	}

	accessToken, err := h.server.JWT.SignAccessToken(user.ID)
	if err != nil {
		log.Println(err)
		return responses.Error(c, http.StatusInternalServerError, "Error logining in user")
	}
	refreshToken, err := h.server.JWT.SignRefreshToken(user.ID)
	if err != nil {
		log.Println(err)
		return responses.Error(c, http.StatusInternalServerError, "Error logining in user")
	}

	h.tokenService.SetAccessCookie(c, accessToken)
	h.tokenService.SetRefreshCookie(c, refreshToken)

	return responses.Success(c, http.StatusOK, "Logined in successfully")
}

func (h *AuthHandler) Refresh(c *echo.Context) error {
	ctx := c.Request().Context()
	cookie, err := c.Request().Cookie("RefreshToken")
	if err != nil {
		return responses.Error(c, http.StatusUnauthorized, "Refresh token required")
	}

	userID, err := h.server.JWT.VerifyRefreshToken(cookie.Value)
	if err != nil {
		return responses.Error(c, http.StatusUnauthorized, "Invalid or expired refresh token")
	}

	user := &models.User{}
	err = h.server.Db.NewSelect().Model(user).Where("id = ?", userID).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return responses.Error(c, http.StatusUnauthorized, "User not found")
	}
	if err != nil {
		log.Printf("Error looking up user while refreshing token: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Error refreshing token")
	}

	accessToken, err := h.server.JWT.SignAccessToken(user.ID)
	if err != nil {
		log.Printf("Error signing refreshed token: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Error refreshing token")
	}

	h.tokenService.SetAccessCookie(c, accessToken)

	return responses.Data(c, http.StatusOK, userresponses.NewUserResponse(user))
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	h.tokenService.ClearAuthCookies(c)

	return responses.Success(c, http.StatusOK, "Signed out successfully")
}
