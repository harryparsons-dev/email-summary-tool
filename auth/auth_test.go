package auth

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

type memoryRefresh struct {
	token   StoredRefreshToken
	revoked bool
}

type memoryReset struct {
	userID    string
	expiresAt time.Time
	used      bool
}

type memoryStore struct {
	mu        sync.Mutex
	users     map[string]User
	byEmail   map[string]string
	refreshes map[string]memoryRefresh
	resets    map[string]memoryReset
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		users: make(map[string]User), byEmail: make(map[string]string),
		refreshes: make(map[string]memoryRefresh), resets: make(map[string]memoryReset),
	}
}

func (s *memoryStore) CreateUserWithRefresh(_ context.Context, user User, token StoredRefreshToken) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byEmail[user.Email]; exists {
		return User{}, ErrEmailInUse
	}
	user.CreatedAt = time.Now().UTC()
	s.users[user.ID] = user
	s.byEmail[user.Email] = user.ID
	s.refreshes[key(token.Hash)] = memoryRefresh{token: token}
	return user, nil
}

func (s *memoryStore) FindUserByEmail(_ context.Context, email string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byEmail[email]
	if !ok {
		return User{}, errNotFound
	}
	return s.users[id], nil
}

func (s *memoryStore) FindUserByID(_ context.Context, id string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return User{}, errNotFound
	}
	return user, nil
}

func (s *memoryStore) StoreRefresh(_ context.Context, token StoredRefreshToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshes[key(token.Hash)] = memoryRefresh{token: token}
	return nil
}

func (s *memoryStore) RotateRefresh(_ context.Context, oldHash []byte, userID, tokenID string, next StoredRefreshToken) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.refreshes[key(oldHash)]
	if !ok || old.revoked || old.token.UserID != userID || old.token.ID != tokenID || !old.token.ExpiresAt.After(time.Now()) {
		return User{}, ErrInvalidToken
	}
	old.revoked = true
	s.refreshes[key(oldHash)] = old
	s.refreshes[key(next.Hash)] = memoryRefresh{token: next}
	user, ok := s.users[userID]
	if !ok {
		return User{}, ErrInvalidToken
	}
	return user, nil
}

func (s *memoryStore) RevokeRefresh(_ context.Context, hash []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.refreshes[key(hash)]
	if ok {
		value.revoked = true
		s.refreshes[key(hash)] = value
	}
	return nil
}

func (s *memoryStore) StorePasswordReset(_ context.Context, userID string, hash []byte, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for hashKey, reset := range s.resets {
		if reset.userID == userID {
			reset.used = true
			s.resets[hashKey] = reset
		}
	}
	s.resets[key(hash)] = memoryReset{userID: userID, expiresAt: expiresAt}
	return nil
}

func (s *memoryStore) ConsumePasswordReset(_ context.Context, hash []byte, passwordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	reset, ok := s.resets[key(hash)]
	if !ok || reset.used || !reset.expiresAt.After(time.Now()) {
		return ErrInvalidToken
	}
	user := s.users[reset.userID]
	user.PasswordHash = passwordHash
	s.users[user.ID] = user
	for hashKey, candidate := range s.resets {
		if candidate.userID == user.ID {
			candidate.used = true
			s.resets[hashKey] = candidate
		}
	}
	for hashKey, refresh := range s.refreshes {
		if refresh.token.UserID == user.ID {
			refresh.revoked = true
			s.refreshes[hashKey] = refresh
		}
	}
	return nil
}

type captureMailer struct {
	email string
	token string
}

func (m *captureMailer) SendPasswordReset(_ context.Context, email, token string) error {
	m.email, m.token = email, token
	return nil
}

func newTestService(t *testing.T) (*Service, *TokenManager, *memoryStore, *captureMailer) {
	t.Helper()
	tokens, err := NewTokenManager("test-secret-that-is-definitely-long-enough", "test", "test-web", 15*time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	store := newMemoryStore()
	mailer := &captureMailer{}
	service, err := NewService(store, tokens, mailer, 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return service, tokens, store, mailer
}

func TestAuthenticationLifecycle(t *testing.T) {
	service, tokens, _, mailer := newTestService(t)
	ctx := context.Background()

	registered, err := service.Register(ctx, "  Person@Example.COM ", "correct horse battery staple")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if registered.User.Email != "person@example.com" {
		t.Fatalf("email was not normalized: %q", registered.User.Email)
	}
	if registered.User.PasswordHash != "" {
		t.Fatal("password hash must not be exposed in session user JSON")
	}
	if _, err := tokens.ParseAccess(registered.AccessToken); err != nil {
		t.Fatalf("access token: %v", err)
	}
	if _, err := tokens.ParseRefresh(registered.RefreshToken); err != nil {
		t.Fatalf("refresh token: %v", err)
	}

	if _, err := service.Register(ctx, "person@example.com", "another valid password"); !errors.Is(err, ErrEmailInUse) {
		t.Fatalf("duplicate registration error = %v", err)
	}
	if _, err := service.Login(ctx, "person@example.com", "wrong password here"); !errors.Is(err, ErrInvalidCredential) {
		t.Fatalf("bad login error = %v", err)
	}
	loggedIn, err := service.Login(ctx, "PERSON@example.com", "correct horse battery staple")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	rotated, err := service.Refresh(ctx, loggedIn.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if rotated.RefreshToken == loggedIn.RefreshToken {
		t.Fatal("refresh token was not rotated")
	}
	if _, err := service.Refresh(ctx, loggedIn.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("replayed refresh error = %v", err)
	}

	if err := service.RequestPasswordReset(ctx, "person@example.com"); err != nil {
		t.Fatalf("request password reset: %v", err)
	}
	if mailer.email != "person@example.com" || mailer.token == "" {
		t.Fatalf("reset message was not captured: %#v", mailer)
	}
	if err := service.ResetPassword(ctx, mailer.token, "a completely new password"); err != nil {
		t.Fatalf("reset password: %v", err)
	}
	if err := service.ResetPassword(ctx, mailer.token, "yet another new password"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("reused reset token error = %v", err)
	}
	if _, err := service.Refresh(ctx, rotated.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("session survived reset: %v", err)
	}
	if _, err := service.Login(ctx, "person@example.com", "correct horse battery staple"); !errors.Is(err, ErrInvalidCredential) {
		t.Fatalf("old password login error = %v", err)
	}
	if _, err := service.Login(ctx, "person@example.com", "a completely new password"); err != nil {
		t.Fatalf("new password login: %v", err)
	}
}

func TestPasswordAndTokenValidation(t *testing.T) {
	if _, err := NormalizeEmail("Name <person@example.com>"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("display-name email error = %v", err)
	}
	if err := ValidatePassword("too short"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("short password error = %v", err)
	}

	_, tokens, _, _ := newTestService(t)
	access, refresh, _, err := tokens.NewPair("e36f1aa0-1841-46e3-bf0c-42ddc72bbaae")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tokens.ParseRefresh(access); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("access accepted as refresh: %v", err)
	}
	if _, err := tokens.ParseAccess(refresh); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("refresh accepted as access: %v", err)
	}
	if _, err := tokens.ParseAccess(access + "tampered"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("tampered token error = %v", err)
	}
}

func TestPasswordResetDoesNotRevealUnknownAccounts(t *testing.T) {
	service, _, _, mailer := newTestService(t)
	if err := service.RequestPasswordReset(context.Background(), "missing@example.com"); err != nil {
		t.Fatalf("unknown account: %v", err)
	}
	if mailer.token != "" {
		t.Fatal("mailer called for unknown account")
	}
}

func TestProductionConfigRequiresSecureCookies(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-that-is-definitely-long-enough")
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_COOKIE_SECURE", "false")
	if _, err := ConfigFromEnv(); err == nil {
		t.Fatal("production config accepted an insecure refresh cookie")
	}
	t.Setenv("AUTH_COOKIE_SECURE", "true")
	if _, err := ConfigFromEnv(); err != nil {
		t.Fatalf("secure production config: %v", err)
	}
}

func TestAuthenticationHTTPFlow(t *testing.T) {
	service, tokens, _, mailer := newTestService(t)
	e := echo.New()
	NewHandler(service, tokens, false, time.Hour).RegisterRoutes(e)

	register := performJSON(t, e, http.MethodPost, "/auth/register", map[string]string{
		"email": "web@example.com", "password": "starting password",
	}, nil, "")
	if register.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", register.Code, register.Body.String())
	}
	var registration tokenResponse
	if err := json.Unmarshal(register.Body.Bytes(), &registration); err != nil {
		t.Fatal(err)
	}
	if registration.AccessToken == "" || bytes.Contains(register.Body.Bytes(), []byte("refresh_token")) {
		t.Fatalf("unexpected registration response: %s", register.Body.String())
	}
	refreshCookie := findCookie(t, register.Result(), "refresh_token")
	if !refreshCookie.HttpOnly || refreshCookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("insecure refresh cookie: %#v", refreshCookie)
	}

	me := performJSON(t, e, http.MethodGet, "/auth/me", nil, nil, "Bearer "+registration.AccessToken)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d, body = %s", me.Code, me.Body.String())
	}

	refreshed := performJSON(t, e, http.MethodPost, "/auth/refresh", map[string]string{}, refreshCookie, "")
	if refreshed.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, body = %s", refreshed.Code, refreshed.Body.String())
	}
	nextCookie := findCookie(t, refreshed.Result(), "refresh_token")
	if nextCookie.Value == refreshCookie.Value {
		t.Fatal("HTTP refresh did not rotate cookie")
	}
	replay := performJSON(t, e, http.MethodPost, "/auth/refresh", map[string]string{}, refreshCookie, "")
	if replay.Code != http.StatusUnauthorized {
		t.Fatalf("refresh replay status = %d", replay.Code)
	}

	forgot := performJSON(t, e, http.MethodPost, "/auth/password/forgot", map[string]string{"email": "web@example.com"}, nil, "")
	if forgot.Code != http.StatusAccepted || mailer.token == "" {
		t.Fatalf("forgot status = %d, token = %q", forgot.Code, mailer.token)
	}
	reset := performJSON(t, e, http.MethodPost, "/auth/password/reset", map[string]string{
		"token": mailer.token, "password": "replacement password",
	}, nil, "")
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status = %d, body = %s", reset.Code, reset.Body.String())
	}

	oldLogin := performJSON(t, e, http.MethodPost, "/auth/login", map[string]string{
		"email": "web@example.com", "password": "starting password",
	}, nil, "")
	if oldLogin.Code != http.StatusUnauthorized {
		t.Fatalf("old password login status = %d", oldLogin.Code)
	}
	newLogin := performJSON(t, e, http.MethodPost, "/auth/login", map[string]string{
		"email": "web@example.com", "password": "replacement password",
	}, nil, "")
	if newLogin.Code != http.StatusOK {
		t.Fatalf("new password login status = %d, body = %s", newLogin.Code, newLogin.Body.String())
	}
	loginCookie := findCookie(t, newLogin.Result(), "refresh_token")
	logout := performJSON(t, e, http.MethodPost, "/auth/logout", map[string]string{}, loginCookie, "")
	if logout.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d", logout.Code)
	}
	cleared := findCookie(t, logout.Result(), "refresh_token")
	if cleared.MaxAge >= 0 || cleared.Value != "" {
		t.Fatalf("logout cookie was not cleared: %#v", cleared)
	}
	loggedOutRefresh := performJSON(t, e, http.MethodPost, "/auth/refresh", map[string]string{}, loginCookie, "")
	if loggedOutRefresh.Code != http.StatusUnauthorized {
		t.Fatalf("logged-out refresh status = %d", loggedOutRefresh.Code)
	}
}

func performJSON(t *testing.T, e *echo.Echo, method, path string, body any, cookie *http.Cookie, authorization string) *httptest.ResponseRecorder {
	t.Helper()
	var encoded []byte
	if body != nil {
		var err error
		encoded, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(encoded))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		request.AddCookie(cookie)
	}
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)
	return recorder
}

func findCookie(t *testing.T, response *http.Response, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}

func key(value []byte) string { return hex.EncodeToString(value) }
