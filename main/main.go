package main

import (
	"context"
	"database/sql"
	"email-summary-tool/database/migrations"
	"email-summary-tool/pkg/jwt"
	"email-summary-tool/pkg/mailserver"
	"email-summary-tool/routes"
	"email-summary-tool/server"
	"email-summary-tool/services"
	"email-summary-tool/services/inbound"
	"email-summary-tool/services/summary"
	summarymanager "email-summary-tool/services/summaryManager"
	"log"
	"net/url"
	"os"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func main() {
	jwtConfig, err := jwt.NewConfig(os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	inboundDomain := envOrDefault("INBOUND_EMAIL_DOMAIN", "localhost")
	smtpAddr := envOrDefault("SMTP_ADDR", ":2525")

	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		log.Fatal("POSTGRES_USER is not set")
	}

	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	if postgresPassword == "" {
		log.Fatal("POSTGRES_PASSWORD is not set")
	}

	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		log.Fatal("POSTGRES_DB is not set")
	}

	databaseURL := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(postgresUser, postgresPassword),
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
		if err := smtpServer.ListenAndServe(); err != nil {
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
