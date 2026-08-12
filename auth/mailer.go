package auth

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"net/url"
	"os"
	"strings"
)

type LogMailer struct{ BaseURL string }

func (m LogMailer) SendPasswordReset(_ context.Context, email, token string) error {
	link, err := resetLink(m.BaseURL, token)
	if err != nil {
		return err
	}
	log.Printf("development password reset for %s: %s", email, link)
	return nil
}

type SMTPMailer struct {
	Host, Port, Username, Password, From, BaseURL string
}

func (m SMTPMailer) SendPasswordReset(_ context.Context, email, token string) error {
	link, err := resetLink(m.BaseURL, token)
	if err != nil {
		return err
	}
	subject := "Reset your Briefly password"
	body := "A password reset was requested for your Briefly account.\r\n\r\n" + link +
		"\r\n\r\nIf you did not request this, you can ignore this email."
	message := []byte("From: " + m.From + "\r\nTo: " + email + "\r\nSubject: " + subject +
		"\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body)
	var auth smtp.Auth
	if m.Username != "" {
		auth = smtp.PlainAuth("", m.Username, m.Password, m.Host)
	}
	if err := smtp.SendMail(m.Host+":"+m.Port, auth, m.From, []string{email}, message); err != nil {
		return fmt.Errorf("deliver reset email: %w", err)
	}
	return nil
}

func MailerFromEnv(baseURL string) (ResetMailer, error) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		if strings.EqualFold(os.Getenv("APP_ENV"), "production") {
			return nil, fmt.Errorf("SMTP_HOST is required in production")
		}
		return LogMailer{BaseURL: baseURL}, nil
	}
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		return nil, fmt.Errorf("SMTP_FROM is required when SMTP_HOST is configured")
	}
	return SMTPMailer{
		Host: host, Port: valueOrDefault("SMTP_PORT", "587"), Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"), From: from, BaseURL: baseURL,
	}, nil
}

func resetLink(baseURL, token string) (string, error) {
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/") + "/reset-password")
	if err != nil {
		return "", fmt.Errorf("build password reset URL: %w", err)
	}
	query := parsed.Query()
	query.Set("token", token)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
