package main

import (
	"context"
	"database/sql"
	"errors"

	"log"
	"mailai/database/migrations"
	"mailai/pkg/jwt"
	"mailai/pkg/mailserver"
	"mailai/routes"
	"mailai/server"
	"mailai/services"
	"mailai/services/inbound"
	"mailai/services/summary"
	summarymanager "mailai/services/summaryManager"
	"net/url"
	"os"

	"github.com/emersion/go-smtp"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func main() {
	jwtConfig, err := jwt.NewConfig(os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	jwtConfig.SecureCookies = os.Getenv("COOKIE_SECURE") == "true"

	inboundDomain := envOrDefault("INBOUND_EMAIL_DOMAIN", "localhost")
	smtpAddr := envOrDefault("SMTP_ADDR", ":2525")

	// The role the API connects as: a restricted role, not the PostgreSQL
	// superuser.
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		log.Fatal("DB_USER is not set")
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		log.Fatal("DB_PASSWORD is not set")
	}

	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		log.Fatal("POSTGRES_DB is not set")
	}

	databaseURL := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(dbUser, dbPassword),
		Host:     "postgres:5432",
		Path:     postgresDB,
		RawQuery: "sslmode=disable",
	}).String()

	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(databaseURL)))
	db := bun.NewDB(sqlDB, pgdialect.New())

	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		log.Fatalf("Unable to connect to database: %v", err)
	}

	if err := migrations.Up(db); err != nil {
		_ = db.Close()
		log.Fatalf("Unable to apply database migrations: %v", err)
	}

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		log.Print("ANTHROPIC_API_KEY is not set; AI summaries will fail until it is")
	}

	s := server.NewServer(jwtConfig, db, services.NewDependencyService(inboundDomain, summary.NewService()))
	defer s.Close()

	if err := summarymanager.NewSummaryManager(s).FailInterrupted(context.Background()); err != nil {
		log.Printf("Unable to clean up interrupted summaries: %v", err)
	}

	routes.InitializeRoutes(s)
	routes.InitializeRestrictedRoutes(s)

	smtpServer := mailserver.New(smtpAddr, inboundDomain, inbound.NewInboundEmailService(s))
	defer smtpServer.Close()
	go func() {
		log.Printf("SMTP server accepting mail for @%s on %s", inboundDomain, smtpAddr)
		if err := smtpServer.ListenAndServe(); err != nil && !errors.Is(err, smtp.ErrServerClosed) {
			log.Fatalf("SMTP server stopped: %v", err)
		}
	}()

	if err := s.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
