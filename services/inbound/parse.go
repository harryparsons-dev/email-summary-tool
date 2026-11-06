package inbound

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"email-summary-tool/database/models"

	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset" // decode non-UTF-8 charsets
	"github.com/emersion/go-message/mail"
)

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
