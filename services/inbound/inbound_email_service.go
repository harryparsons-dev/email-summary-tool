// Package inbound generates project email addresses, parses the mail
// delivered to them and saves it.
package inbound

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"email-summary-tool/database/models"
	"email-summary-tool/server"

	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset" // decode non-UTF-8 charsets
	"github.com/emersion/go-message/mail"
)

// InboundEmailService looks up projects by inbound address and saves delivered
// mail. The SMTP server uses it today; a provider webhook could call the same
// methods. It uses the server's database rather than holding its own.
type InboundEmailService struct {
	Server *server.Server
}

func NewInboundEmailService(s *server.Server) *InboundEmailService {
	return &InboundEmailService{Server: s}
}

// FindProjectID returns the ID of the active project that owns address, or
// ok=false if no project does.
func (s *InboundEmailService) FindProjectID(ctx context.Context, address string) (projectID string, ok bool, err error) {
	address = NormalizeAddress(address)
	if AddressDomain(address) != NormalizeAddress(s.Server.InboundDomain) {
		return "", false, nil
	}

	project := &models.Project{}
	err = s.Server.Db.NewSelect().
		Model(project).
		Column("id").
		Where("email_address = ?", address).
		Limit(1).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("look up project for %s: %w", address, err)
	}
	return project.ID, true, nil
}

// Recipient is one accepted envelope recipient and the project it belongs to.
type Recipient struct {
	Address   string
	ProjectID string
}

// Deliver parses raw once and saves a copy for each recipient's project in a
// single insert, so a failure never leaves the message half-delivered. A
// project listed more than once gets one copy.
func (s *InboundEmailService) Deliver(ctx context.Context, envelopeFrom string, recipients []Recipient, raw []byte) error {
	parsed, parseErr := Parse(raw)
	if parseErr != nil {
		// The raw message is still saved, so nothing is lost and it can be
		// re-parsed later.
		log.Printf("inbound email only partly parsed: %v", parseErr)
	}

	seen := make(map[string]bool, len(recipients))
	emails := make([]*models.Email, 0, len(recipients))
	for _, recipient := range recipients {
		if seen[recipient.ProjectID] {
			continue
		}
		seen[recipient.ProjectID] = true
		emails = append(emails, &models.Email{
			ProjectID:    recipient.ProjectID,
			EnvelopeFrom: cleanText(envelopeFrom),
			EnvelopeTo:   NormalizeAddress(recipient.Address),
			MessageID:    parsed.MessageID,
			FromAddress:  parsed.FromAddress,
			FromName:     parsed.FromName,
			ToAddresses:  parsed.ToAddresses,
			Subject:      parsed.Subject,
			SentAt:       parsed.SentAt,
			TextBody:     parsed.TextBody,
			HTMLBody:     parsed.HTMLBody,
			Attachments:  parsed.Attachments,
			RawMessage:   raw,
		})
	}
	if len(emails) == 0 {
		return nil
	}

	if _, err := s.Server.Db.NewInsert().Model(&emails).Exec(ctx); err != nil {
		return fmt.Errorf("save inbound email: %w", err)
	}
	return nil
}

// localPartAlphabet avoids characters that are easy to misread (0/o, 1/l).
const localPartAlphabet = "abcdefghijkmnpqrstuvwxyz23456789"

const localPartLength = 12

// GenerateAddress returns a new random address such as "k3x9p2qahm7d@domain".
func GenerateAddress(domain string) (string, error) {
	buf := make([]byte, localPartLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate email address: %w", err)
	}
	for i, b := range buf {
		// 256 is a multiple of the 32-character alphabet, so this is unbiased.
		buf[i] = localPartAlphabet[int(b)%len(localPartAlphabet)]
	}
	return string(buf) + "@" + strings.ToLower(domain), nil
}

// NormalizeAddress lowercases an address and strips surrounding whitespace and
// angle brackets so it can be compared with stored project addresses.
func NormalizeAddress(address string) string {
	address = strings.TrimSpace(address)
	address = strings.TrimPrefix(address, "<")
	address = strings.TrimSuffix(address, ">")
	return strings.ToLower(address)
}

// AddressDomain returns the part of the address after the last "@".
func AddressDomain(address string) string {
	at := strings.LastIndex(address, "@")
	if at < 0 {
		return ""
	}
	return address[at+1:]
}

// maxBodyBytes caps how much of each text or HTML part is stored.
const maxBodyBytes = 5 << 20

// ParsedEmail holds the parts of a message that are stored in their own
// columns.
type ParsedEmail struct {
	MessageID   string
	FromAddress string
	FromName    string
	ToAddresses string
	Subject     string
	SentAt      *time.Time
	TextBody    string
	HTMLBody    string
	Attachments []models.EmailAttachment
}

// Parse extracts headers, bodies and attachment details from a raw RFC 5322
// message. Malformed parts are skipped where possible, so a partly broken
// message still returns whatever could be read alongside the error.
func Parse(raw []byte) (*ParsedEmail, error) {
	parsed := &ParsedEmail{Attachments: []models.EmailAttachment{}}

	reader, err := mail.CreateReader(bytes.NewReader(raw))
	if err != nil && !message.IsUnknownCharset(err) {
		return parsed, fmt.Errorf("read message: %w", err)
	}
	defer reader.Close()

	header := reader.Header
	parsed.MessageID, _ = header.MessageID()
	parsed.Subject, _ = header.Subject()
	if parsed.Subject == "" {
		parsed.Subject = header.Get("Subject")
	}
	if from, err := header.AddressList("From"); err == nil && len(from) > 0 {
		parsed.FromAddress = strings.ToLower(from[0].Address)
		parsed.FromName = from[0].Name
	} else {
		parsed.FromAddress = header.Get("From")
	}
	if to, err := header.AddressList("To"); err == nil {
		addresses := make([]string, 0, len(to))
		for _, address := range to {
			addresses = append(addresses, address.Address)
		}
		parsed.ToAddresses = strings.Join(addresses, ", ")
	} else {
		parsed.ToAddresses = header.Get("To")
	}
	if sentAt, err := header.Date(); err == nil && !sentAt.IsZero() {
		parsed.SentAt = &sentAt
	}

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			if message.IsUnknownCharset(err) || message.IsUnknownEncoding(err) {
				continue
			}
			return clean(parsed), fmt.Errorf("read message part: %w", err)
		}

		switch partHeader := part.Header.(type) {
		case *mail.InlineHeader:
			contentType, _, _ := partHeader.ContentType()
			switch contentType {
			case "text/plain", "":
				if parsed.TextBody == "" {
					parsed.TextBody = readBody(part.Body)
				}
			case "text/html":
				if parsed.HTMLBody == "" {
					parsed.HTMLBody = readBody(part.Body)
				}
			default:
				// Inline images and similar parts are recorded like attachments.
				parsed.Attachments = append(parsed.Attachments, attachmentInfo(part.Body, "", contentType))
			}
		case *mail.AttachmentHeader:
			filename, _ := partHeader.Filename()
			contentType, _, _ := partHeader.ContentType()
			parsed.Attachments = append(parsed.Attachments, attachmentInfo(part.Body, filename, contentType))
		}
	}

	return clean(parsed), nil
}

func readBody(body io.Reader) string {
	content, _ := io.ReadAll(io.LimitReader(body, maxBodyBytes))
	return string(content)
}

func attachmentInfo(body io.Reader, filename, contentType string) models.EmailAttachment {
	size, _ := io.Copy(io.Discard, body)
	return models.EmailAttachment{Filename: filename, ContentType: contentType, Size: size}
}

// clean makes every string safe for a PostgreSQL TEXT column, which rejects
// NUL bytes and invalid UTF-8.
func clean(parsed *ParsedEmail) *ParsedEmail {
	for _, field := range []*string{
		&parsed.MessageID, &parsed.FromAddress, &parsed.FromName, &parsed.ToAddresses,
		&parsed.Subject, &parsed.TextBody, &parsed.HTMLBody,
	} {
		*field = cleanText(*field)
	}
	for i := range parsed.Attachments {
		parsed.Attachments[i].Filename = cleanText(parsed.Attachments[i].Filename)
		parsed.Attachments[i].ContentType = cleanText(parsed.Attachments[i].ContentType)
	}
	return parsed
}

func cleanText(value string) string {
	if !utf8.ValidString(value) {
		value = strings.ToValidUTF8(value, "�")
	}
	return strings.ReplaceAll(value, "\x00", "")
}
