package mailserver

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"

	"email-summary-tool/services/inbound"

	"github.com/emersion/go-smtp"
)

type fakeInbox struct {
	mu         sync.Mutex
	projects   map[string]string
	from       string
	recipients []inbound.Recipient
	raw        []byte
	lookupErr  error
}

func (f *fakeInbox) FindProjectID(_ context.Context, address string) (string, bool, error) {
	if f.lookupErr != nil {
		return "", false, f.lookupErr
	}
	id, ok := f.projects[inbound.NormalizeAddress(address)]
	return id, ok, nil
}

func (f *fakeInbox) Deliver(_ context.Context, from string, recipients []inbound.Recipient, raw []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.from, f.recipients, f.raw = from, recipients, raw
	return nil
}

func startServer(t *testing.T, inbox Inbox) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := New(listener.Addr().String(), "mail.test", inbox)
	go server.Serve(listener)
	t.Cleanup(func() { server.Close() })
	return listener.Addr().String()
}

// send delivers a message over plain SMTP, as another mail server would
// when the receiver does not offer STARTTLS.
func send(t *testing.T, addr, from string, to []string, r *strings.Reader) error {
	t.Helper()
	client, err := smtp.Dial(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	return client.SendMail(from, to, r)
}

const message = "From: sender@example.com\r\nSubject: hi\r\n\r\nhello\r\n"

func TestAcceptsKnownRecipient(t *testing.T) {
	inbox := &fakeInbox{projects: map[string]string{"abc@mail.test": "project-1"}}
	addr := startServer(t, inbox)

	err := send(t, addr, "sender@example.com", []string{"ABC@mail.test"}, strings.NewReader(message))
	if err != nil {
		t.Fatal(err)
	}

	inbox.mu.Lock()
	defer inbox.mu.Unlock()
	if inbox.from != "sender@example.com" {
		t.Errorf("from = %q", inbox.from)
	}
	if len(inbox.recipients) != 1 || inbox.recipients[0].ProjectID != "project-1" {
		t.Errorf("recipients = %+v", inbox.recipients)
	}
	if !strings.Contains(string(inbox.raw), "hello") {
		t.Errorf("raw = %q", inbox.raw)
	}
}

func TestRejectsUnknownRecipient(t *testing.T) {
	addr := startServer(t, &fakeInbox{projects: map[string]string{}})

	err := send(t, addr, "sender@example.com", []string{"nobody@mail.test"}, strings.NewReader(message))
	var smtpErr *smtp.SMTPError
	if !errors.As(err, &smtpErr) || smtpErr.Code != 550 {
		t.Fatalf("expected 550, got %v", err)
	}
}

func TestLookupFailureIsTemporary(t *testing.T) {
	addr := startServer(t, &fakeInbox{lookupErr: errors.New("db down")})

	err := send(t, addr, "sender@example.com", []string{"abc@mail.test"}, strings.NewReader(message))
	var smtpErr *smtp.SMTPError
	if !errors.As(err, &smtpErr) || smtpErr.Code != 451 {
		t.Fatalf("expected 451, got %v", err)
	}
}
