package servicetests

import (
	"mailai/pkg/jwt"
	tokenservice "mailai/services/tokenService"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestTokenServiceSetsAuthenticationCookies(t *testing.T) {
	service := newTestTokenService(t)
	response := httptest.NewRecorder()
	context := echo.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), response)

	service.SetAccessCookie(context, "access-token")
	service.SetRefreshCookie(context, "refresh-token")

	cookies := response.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("got %d cookies, want 2", len(cookies))
	}

	assertCookie(t, cookies[0], "Authorization", "access-token", 15*60)
	assertCookie(t, cookies[1], "RefreshToken", "refresh-token", 7*24*60*60)
}

func TestTokenServiceCookieSecureFlagFollowsConfig(t *testing.T) {
	tests := []struct {
		name          string
		secureCookies bool
	}{
		{name: "HTTPS-only when secure cookies are enabled", secureCookies: true},
		{name: "allowed over HTTP when secure cookies are disabled", secureCookies: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jwtConfig, err := jwt.NewConfig("a-test-secret-that-is-at-least-32-characters")
			if err != nil {
				t.Fatalf("NewConfig() error = %v", err)
			}
			jwtConfig.SecureCookies = test.secureCookies
			service := tokenservice.NewTokenService(jwtConfig)

			response := httptest.NewRecorder()
			context := echo.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), response)

			service.SetAccessCookie(context, "access-token")
			service.SetRefreshCookie(context, "refresh-token")
			service.ClearAuthCookies(context)

			cookies := response.Result().Cookies()
			if len(cookies) != 4 {
				t.Fatalf("got %d cookies, want 4", len(cookies))
			}
			for _, cookie := range cookies {
				if cookie.Secure != test.secureCookies {
					t.Errorf("cookie %q Secure = %t, want %t", cookie.Name, cookie.Secure, test.secureCookies)
				}
			}
		})
	}
}

func TestTokenServiceClearsAuthenticationCookies(t *testing.T) {
	service := newTestTokenService(t)
	response := httptest.NewRecorder()
	context := echo.NewContext(httptest.NewRequest(http.MethodPost, "/logout", nil), response)

	service.ClearAuthCookies(context)

	cookies := response.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("got %d cookies, want 2", len(cookies))
	}

	for _, cookie := range cookies {
		if cookie.MaxAge >= 0 {
			t.Errorf("cookie %q MaxAge = %d, want a negative value", cookie.Name, cookie.MaxAge)
		}
		if cookie.Value != "" {
			t.Errorf("cookie %q value = %q, want empty", cookie.Name, cookie.Value)
		}
	}
}

func newTestTokenService(t *testing.T) *tokenservice.TokenService {
	t.Helper()

	jwtConfig, err := jwt.NewConfig("a-test-secret-that-is-at-least-32-characters")
	if err != nil {
		t.Fatalf("NewJWTConfig() error = %v", err)
	}

	return tokenservice.NewTokenService(jwtConfig)
}

func assertCookie(t *testing.T, cookie *http.Cookie, name, value string, maxAge int) {
	t.Helper()

	if cookie.Name != name {
		t.Errorf("cookie name = %q, want %q", cookie.Name, name)
	}
	if cookie.Value != value {
		t.Errorf("cookie value = %q, want %q", cookie.Value, value)
	}
	if cookie.MaxAge != maxAge {
		t.Errorf("cookie MaxAge = %d, want %d", cookie.MaxAge, maxAge)
	}
	if cookie.Path != "/" {
		t.Errorf("cookie Path = %q, want /", cookie.Path)
	}
	if !cookie.HttpOnly {
		t.Error("cookie HttpOnly = false, want true")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie SameSite = %v, want %v", cookie.SameSite, http.SameSiteLaxMode)
	}
}
