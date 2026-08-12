package auth

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

const userIDContextKey = "auth.user_id"

type Handler struct {
	service      *Service
	tokens       *TokenManager
	cookieSecure bool
	refreshTTL   time.Duration
}

func NewHandler(service *Service, tokens *TokenManager, cookieSecure bool, refreshTTL time.Duration) *Handler {
	return &Handler{service: service, tokens: tokens, cookieSecure: cookieSecure, refreshTTL: refreshTTL}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	group := e.Group("/auth")
	sensitiveLimiter := middleware.RateLimiter(middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{Rate: 0.2, Burst: 10, ExpiresIn: 10 * time.Minute},
	))
	refreshLimiter := middleware.RateLimiter(middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{Rate: 1, Burst: 20, ExpiresIn: 10 * time.Minute},
	))
	group.POST("/register", h.register, sensitiveLimiter)
	group.POST("/login", h.login, sensitiveLimiter)
	group.POST("/refresh", h.refresh, refreshLimiter, requireJSON)
	group.POST("/logout", h.logout, requireJSON)
	group.POST("/password/forgot", h.forgotPassword, sensitiveLimiter)
	group.POST("/password/reset", h.resetPassword, sensitiveLimiter)
	group.GET("/me", h.me, h.authenticate)
}

type credentialRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	User        User   `json:"user"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (h *Handler) register(c *echo.Context) error {
	var request credentialRequest
	if err := decodeJSON(c, &request); err != nil {
		return errorJSON(c, http.StatusBadRequest, "invalid_request", err.Error())
	}
	session, err := h.service.Register(c.Request().Context(), request.Email, request.Password)
	if err != nil {
		return h.authError(c, err)
	}
	h.setRefreshCookie(c, session.RefreshToken)
	return c.JSON(http.StatusCreated, responseFor(session))
}

func (h *Handler) login(c *echo.Context) error {
	var request credentialRequest
	if err := decodeJSON(c, &request); err != nil {
		return errorJSON(c, http.StatusBadRequest, "invalid_request", err.Error())
	}
	session, err := h.service.Login(c.Request().Context(), request.Email, request.Password)
	if err != nil {
		return h.authError(c, err)
	}
	h.setRefreshCookie(c, session.RefreshToken)
	return c.JSON(http.StatusOK, responseFor(session))
}

func (h *Handler) refresh(c *echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		return errorJSON(c, http.StatusUnauthorized, "invalid_token", "refresh token is invalid or expired")
	}
	session, err := h.service.Refresh(c.Request().Context(), cookie.Value)
	if err != nil {
		h.clearRefreshCookie(c)
		return h.authError(c, err)
	}
	h.setRefreshCookie(c, session.RefreshToken)
	return c.JSON(http.StatusOK, responseFor(session))
}

func (h *Handler) logout(c *echo.Context) error {
	if cookie, err := c.Cookie("refresh_token"); err == nil {
		if err := h.service.Logout(c.Request().Context(), cookie.Value); err != nil {
			log.Printf("logout token revocation failed: %v", err)
		}
	}
	h.clearRefreshCookie(c)
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) forgotPassword(c *echo.Context) error {
	var request struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(c, &request); err != nil {
		return errorJSON(c, http.StatusBadRequest, "invalid_request", err.Error())
	}
	if err := h.service.RequestPasswordReset(c.Request().Context(), request.Email); err != nil {
		log.Printf("password reset request failed: %v", err)
	}
	return c.JSON(http.StatusAccepted, map[string]string{
		"message": "If an account exists for that email, a reset link has been sent.",
	})
}

func (h *Handler) resetPassword(c *echo.Context) error {
	var request struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := decodeJSON(c, &request); err != nil {
		return errorJSON(c, http.StatusBadRequest, "invalid_request", err.Error())
	}
	if err := h.service.ResetPassword(c.Request().Context(), request.Token, request.Password); err != nil {
		return h.authError(c, err)
	}
	h.clearRefreshCookie(c)
	return c.JSON(http.StatusOK, map[string]string{"message": "Password updated. Please log in again."})
}

func (h *Handler) me(c *echo.Context) error {
	id, _ := c.Get(userIDContextKey).(string)
	user, err := h.service.User(c.Request().Context(), id)
	if err != nil {
		return h.authError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]User{"user": user})
}

func (h *Handler) authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		header := c.Request().Header.Get("Authorization")
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return errorJSON(c, http.StatusUnauthorized, "invalid_token", "a valid bearer token is required")
		}
		claims, err := h.tokens.ParseAccess(parts[1])
		if err != nil {
			return errorJSON(c, http.StatusUnauthorized, "invalid_token", "a valid bearer token is required")
		}
		c.Set(userIDContextKey, claims.Subject)
		return next(c)
	}
}

func (h *Handler) authError(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return errorJSON(c, http.StatusBadRequest, "invalid_request", strings.TrimPrefix(err.Error(), ErrInvalidInput.Error()+": "))
	case errors.Is(err, ErrEmailInUse):
		return errorJSON(c, http.StatusConflict, "email_in_use", ErrEmailInUse.Error())
	case errors.Is(err, ErrInvalidCredential):
		return errorJSON(c, http.StatusUnauthorized, "invalid_credentials", ErrInvalidCredential.Error())
	case errors.Is(err, ErrInvalidToken):
		return errorJSON(c, http.StatusUnauthorized, "invalid_token", ErrInvalidToken.Error())
	default:
		log.Printf("authentication request failed: %v", err)
		return errorJSON(c, http.StatusInternalServerError, "internal_error", "The request could not be completed.")
	}
}

func (h *Handler) setRefreshCookie(c *echo.Context, value string) {
	c.SetCookie(&http.Cookie{
		Name: "refresh_token", Value: value, Path: "/", HttpOnly: true, Secure: h.cookieSecure,
		SameSite: http.SameSiteStrictMode, MaxAge: int(h.refreshTTL.Seconds()),
		Expires: time.Now().Add(h.refreshTTL),
	})
}

func (h *Handler) clearRefreshCookie(c *echo.Context) {
	c.SetCookie(&http.Cookie{
		Name: "refresh_token", Value: "", Path: "/", HttpOnly: true, Secure: h.cookieSecure,
		SameSite: http.SameSiteStrictMode, MaxAge: -1, Expires: time.Unix(1, 0),
	})
}

func responseFor(session Session) tokenResponse {
	return tokenResponse{User: session.User, AccessToken: session.AccessToken, TokenType: "Bearer", ExpiresIn: session.ExpiresIn}
}

func decodeJSON(c *echo.Context, destination any) error {
	mediaType, _, err := mime.ParseMediaType(c.Request().Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return errors.New("Content-Type must be application/json")
	}
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 1<<20)
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return errors.New("request body must be valid JSON")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func requireJSON(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		mediaType, _, err := mime.ParseMediaType(c.Request().Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			return errorJSON(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
		}
		return next(c)
	}
}

func errorJSON(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
