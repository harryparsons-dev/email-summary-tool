// Package mailserver runs an SMTP server that accepts mail for project
// addresses and hands it to the inbound email service to be stored.
package mailserver

import (
	"context"
	"io"
	"log"
	"time"

	"email-summary-tool/services/inbound"

	"github.com/emersion/go-smtp"
)

const (
	maxMessageBytes = 25 << 20 // 25 MB, in line with common provider limits
	maxRecipients   = 50
	deliverTimeout  = 30 * time.Second
)

// New returns an SMTP server listening on addr that accepts mail for domain.
func New(addr, domain string, inbox *inbound.InboundEmailService) *smtp.Server {
	server := smtp.NewServer(&backend{inbox: inbox})
	server.Addr = addr
	server.Domain = domain
	server.MaxMessageBytes = maxMessageBytes
	server.MaxRecipients = maxRecipients
	server.ReadTimeout = time.Minute
	server.WriteTimeout = time.Minute
	return server
}

type backend struct {
	inbox *inbound.InboundEmailService
}

func (b *backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &session{inbox: b.inbox}, nil
}

// session tracks one SMTP transaction. It does not implement AUTH, so the
// server only accepts mail for our own addresses and never relays.
type session struct {
	inbox      *inbound.InboundEmailService
	from       string
	recipients []inbound.Recipient
}

func (s *session) Mail(from string, _ *smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *session) Rcpt(to string, _ *smtp.RcptOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := s.inbox.FindProject(ctx, to)
	if err != nil {
		log.Printf("smtp: %v", err)
		// The mailbox may exist, so ask the sending server to retry rather
		// than bounce the message.
		return &smtp.SMTPError{
			Code:         451,
			EnhancedCode: smtp.EnhancedCode{4, 3, 0},
			Message:      "Temporary error checking mailbox, please try again later",
		}
	}
	if project == nil {
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 1, 1},
			Message:      "No such mailbox",
		}
	}

	s.recipients = append(s.recipients, inbound.Recipient{Address: to, ProjectID: project.ID})
	return nil
}

func (s *session) Data(r io.Reader) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), deliverTimeout)
	defer cancel()

	if err := s.inbox.Deliver(ctx, s.from, s.recipients, raw); err != nil {
		log.Printf("smtp: %v", err)
		// A temporary failure makes the sending server retry later.
		return &smtp.SMTPError{
			Code:         451,
			EnhancedCode: smtp.EnhancedCode{4, 3, 0},
			Message:      "Temporary error saving message, please try again later",
		}
	}
	log.Printf("smtp: received message from %q for %d recipient(s)", s.from, len(s.recipients))
	return nil
}

func (s *session) Reset() {
	s.from = ""
	s.recipients = nil
}

func (s *session) Logout() error {
	return nil
}
