package tokenservice

import (
	"email-summary-tool/pkg/jwt"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

const (
	accessCookieName  = "Authorization"
	refreshCookieName = "RefreshToken"
)

type TokenService struct {
	jwtConfig *jwt.Config
}

func NewTokenService(jwtConfig *jwt.Config) *TokenService {
	return &TokenService{jwtConfig: jwtConfig}
}

func (s *TokenService) SetAccessCookie(c *echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     accessCookieName,
		Value:    token,
		MaxAge:   s.jwtConfig.AccessCookieMaxAge(),
		Path:     "/",
		Secure:   s.jwtConfig.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *TokenService) SetRefreshCookie(c *echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		MaxAge:   s.jwtConfig.RefreshCookieMaxAge(),
		Path:     "/",
		Secure:   s.jwtConfig.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *TokenService) ClearAuthCookies(c *echo.Context) {
	s.clearCookie(c, accessCookieName)
	s.clearCookie(c, refreshCookieName)
}

func (s *TokenService) clearCookie(c *echo.Context, name string) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		Secure:   s.jwtConfig.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
