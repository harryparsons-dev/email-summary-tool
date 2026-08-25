package tests

import (
	"context"
	"database/sql"
	"email-summary-tool/config"
	"email-summary-tool/database/migrations"
	"email-summary-tool/database/models"
	"email-summary-tool/routes"
	"email-summary-tool/server"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

const (
	testJWTSecret   = "integration-test-secret-key-32-characters-minimum"
	testUserEmail   = "api-test@example.com"
	testDatabaseURL = "postgres://postgres:postgres@127.0.0.1:5435/email_summary_test?sslmode=disable"
)

type APITestServer struct {
	*server.Server
	HTTPServer *httptest.Server
	Client     *http.Client
	AuthCookie *http.Cookie
	User       *models.User
}

// checkTestDatabase verifies the dependency without changing Docker state.
func checkTestDatabase() error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("locate test database configuration")
	}
	composeFile := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "compose.test.yaml"))

	command := exec.Command(
		"docker", "compose", "-f", composeFile,
		"ps", "--status", "running", "--services", "test-postgres",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("inspect PostgreSQL container: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if strings.TrimSpace(string(output)) != "test-postgres" {
		return fmt.Errorf(
			"test-postgres container is not running; start it with `docker compose -f compose.test.yaml up --detach --wait test-postgres`",
		)
	}

	if os.Getenv("TEST_DATABASE_URL") == "" {
		if err := os.Setenv("TEST_DATABASE_URL", testDatabaseURL); err != nil {
			return fmt.Errorf("set test database URL: %w", err)
		}
	}
	return nil
}

func NewTestServer() (*APITestServer, error) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("TEST_DATABASE_URL is not set")
	}

	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(databaseURL)))
	db := bun.NewDB(sqlDB, pgdialect.New())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect to test database: %w", err)
	}

	if err := migrations.Up(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply test database migrations: %w", err)
	}

	jwtConfig := &config.JWTConfig{
		SigningKey:      []byte(testJWTSecret),
		AccessLifetime:  time.Hour,
		RefreshLifetime: time.Hour,
	}
	s := server.NewServer(jwtConfig, db)
	routes.InitializeRoutes(s)

	user := &models.User{
		Email:        testUserEmail,
		PasswordHash: "not-used-by-integration-tests",
	}
	if _, err := db.NewInsert().Model(user).Exec(ctx); err != nil {
		s.Close()
		return nil, fmt.Errorf("create test user: %w", err)
	}

	accessToken, err := jwtConfig.SignAccessToken(user.ID)
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("create test access token: %w", err)
	}

	httpServer := httptest.NewServer(s.E)
	return &APITestServer{
		Server:     s,
		HTTPServer: httpServer,
		Client:     httpServer.Client(),
		AuthCookie: &http.Cookie{
			Name:  "Authorization",
			Value: accessToken,
			Path:  "/",
		},
		User: user,
	}, nil
}

func (ts *APITestServer) Close() {
	ts.HTTPServer.Close()
	ts.Server.Close()
}
