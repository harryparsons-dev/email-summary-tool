package main

import (
	"email-summary-tool/config"
	"email-summary-tool/mailserver"
	"email-summary-tool/routes"
	"email-summary-tool/server"
	"email-summary-tool/services/inbound"
	"log"
	"os"
)

func main() {
	jwtConfig, err := config.NewJWTConfig(os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	inboundDomain := envOrDefault("INBOUND_EMAIL_DOMAIN", "localhost")
	smtpAddr := envOrDefault("SMTP_ADDR", ":2525")

	s := server.NewServer(jwtConfig, inboundDomain)
	defer s.Close()

	routes.InitializeRoutes(s)

	smtpServer := mailserver.New(smtpAddr, inboundDomain, inbound.NewStore(s.DB, inboundDomain))
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
