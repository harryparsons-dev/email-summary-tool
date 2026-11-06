package inbound

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"email-summary-tool/database/models"

	"github.com/uptrace/bun"
)

// Store looks up projects by inbound address and saves delivered mail. The
// SMTP server uses it today; a provider webhook could call the same methods.
type Store struct {
	DB     *bun.DB
	Domain string
}

func NewStore(db *bun.DB, domain string) *Store {
	return &Store{DB: db, Domain: NormalizeAddress(domain)}
}

// FindProjectID returns the ID of the active project that owns address, or
// ok=false if no project does.
func (s *Store) FindProjectID(ctx context.Context, address string) (projectID string, ok bool, err error) {
	address = NormalizeAddress(address)
	if AddressDomain(address) != s.Domain {
		return "", false, nil
	}

	project := &models.Project{}
	err = s.DB.NewSelect().
		Model(project).
		Column("id").
		Where("email_address = ?", address).
		Where("deleted_at IS NULL").
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
func (s *Store) Deliver(ctx context.Context, envelopeFrom string, recipients []Recipient, raw []byte) error {
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

	if _, err := s.DB.NewInsert().Model(&emails).Exec(ctx); err != nil {
		return fmt.Errorf("save inbound email: %w", err)
	}
	return nil
}
