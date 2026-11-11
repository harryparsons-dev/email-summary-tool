package tests

import (
	"errors"
	"net"
	"strings"
	"testing"

	"mailai/database/factories"
	"mailai/database/models"
	"mailai/pkg/mailserver"
	"mailai/services/inbound"

	"github.com/emersion/go-smtp"
)

func startMailServer(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := mailserver.New(listener.Addr().String(), TestServer.Dependencies.InboundDomain, inbound.NewInboundEmailService(TestServer.Server))
	go server.Serve(listener)
	t.Cleanup(func() { server.Close() })
	return listener.Addr().String()
}

// sendMail delivers a message over plain SMTP, as another mail server would
// when the receiver does not offer STARTTLS.
func sendMail(t *testing.T, addr, from string, to []string, message string) error {
	t.Helper()
	client, err := smtp.Dial(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	return client.SendMail(from, to, strings.NewReader(message))
}

const testMailMessage = "From: sender@example.com\r\nSubject: hi\r\n\r\nhello\r\n"

func TestMailServerAcceptsKnownRecipient(t *testing.T) {
	// A separate user keeps this project out of the shared user's project list.
	user := &models.User{Email: "mailserver-user@example.com"}
	if _, err := TestServer.Db.NewInsert().Model(user).Exec(t.Context()); err != nil {
		t.Fatalf("create user fixture: %v", err)
	}
	project, err := factories.NewProject(t.Context(), TestServer.Db, user.ID)
	if err != nil {
		t.Fatalf("create project fixture: %v", err)
	}

	addr := startMailServer(t)

	err = sendMail(t, addr, "sender@example.com", []string{strings.ToUpper(project.EmailAddress)}, testMailMessage)
	if err != nil {
		t.Fatal(err)
	}

	var emails []models.Email
	if err := TestServer.Db.NewSelect().Model(&emails).Where("project_id = ?", project.ID).Scan(t.Context()); err != nil {
		t.Fatalf("load delivered emails: %v", err)
	}
	if len(emails) != 1 {
		t.Fatalf("saved %d emails, want 1", len(emails))
	}
	if emails[0].EnvelopeFrom != "sender@example.com" || emails[0].Subject != "hi" || !strings.Contains(emails[0].TextBody, "hello") {
		t.Errorf("saved email = from %q, subject %q, body %q", emails[0].EnvelopeFrom, emails[0].Subject, emails[0].TextBody)
	}
}

func TestMailServerRejectsUnknownRecipient(t *testing.T) {
	addr := startMailServer(t)

	err := sendMail(t, addr, "sender@example.com", []string{"nobody@" + TestServer.Dependencies.InboundDomain}, testMailMessage)
	var smtpErr *smtp.SMTPError
	if !errors.As(err, &smtpErr) || smtpErr.Code != 550 {
		t.Fatalf("expected 550, got %v", err)
	}
}
