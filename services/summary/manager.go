package summary

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"email-summary-tool/database/models"

	"github.com/uptrace/bun"
)

const (
	// jobTimeout bounds how long one background summary may take.
	jobTimeout = 3 * time.Minute

	genericFailureMessage     = "The summary could not be generated. Please try again."
	interruptedFailureMessage = "The summary was interrupted. Please try again."
)

// Summarizer produces summary text. Service implements it; tests can supply a
// fake so no API calls are made.
type Summarizer interface {
	Summarize(ctx context.Context, email *models.Email) (*Result, error)
	SummarizeProject(ctx context.Context, project *models.Project, emails []models.Email) (*Result, error)
}

// Manager stores summaries and generates their text.
type Manager struct {
	db         *bun.DB
	summarizer Summarizer
}

func NewManager(db *bun.DB, summarizer Summarizer) *Manager {
	return &Manager{db: db, summarizer: summarizer}
}

// Create saves a pending summary. The caller then runs Generate, usually in a
// background goroutine. emailID nil means the whole project is summarised. If
// a matching summary is already pending, that one is returned instead and
// created is false.
func (m *Manager) Create(ctx context.Context, userID string, project *models.Project, emailID *string) (summary *models.Summary, created bool, err error) {
	existing := &models.Summary{}
	query := m.db.NewSelect().
		Model(existing).
		Where("s.project_id = ?", project.ID).
		Where("s.user_id = ?", userID).
		Where("s.status = ?", models.SummaryStatusPending).
		OrderExpr("s.created_at DESC").
		Limit(1)
	if emailID == nil {
		query = query.Where("s.email_id IS NULL")
	} else {
		query = query.Where("s.email_id = ?", *emailID)
	}
	err = query.Scan(ctx)
	if err == nil {
		return existing, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, fmt.Errorf("check for pending summary: %w", err)
	}

	summary = &models.Summary{
		UserId:    userID,
		ProjectId: project.ID,
		EmailId:   emailID,
		Status:    models.SummaryStatusPending,
	}
	if _, err := m.db.NewInsert().Model(summary).Returning("*").Exec(ctx); err != nil {
		return nil, false, fmt.Errorf("create summary: %w", err)
	}

	return summary, true, nil
}

// FailInterrupted marks summaries still pending from a previous run of the
// server as failed. Call it once at startup, before serving requests.
func (m *Manager) FailInterrupted(ctx context.Context) error {
	_, err := m.db.NewUpdate().
		Model((*models.Summary)(nil)).
		Set("status = ?", models.SummaryStatusFailed).
		Set("error = ?", interruptedFailureMessage).
		Set("updated_at = ?", time.Now()).
		Where("status = ?", models.SummaryStatusPending).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("fail interrupted summaries: %w", err)
	}
	return nil
}

// Generate creates the text for a pending summary and saves the outcome. It
// is meant to run in a background goroutine after the request that created
// the summary has returned, so it uses its own context rather than the
// request's.
func (m *Manager) Generate(summaryID string, project *models.Project, emailID *string) {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	var (
		result *Result
		err    error
	)
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()
		result, err = m.generate(ctx, project, emailID)
	}()

	update := m.db.NewUpdate().
		Model((*models.Summary)(nil)).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", summaryID)
	if err != nil {
		log.Printf("Error generating summary %s: %v", summaryID, err)
		update = update.
			Set("status = ?", models.SummaryStatusFailed).
			Set("error = ?", failureMessage(err))
	} else {
		update = update.
			Set("status = ?", models.SummaryStatusCompleted).
			Set("summary = ?", result.Summary)
	}

	// Save the outcome even if the job ran out of time.
	saveCtx, saveCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer saveCancel()
	if _, err := update.Exec(saveCtx); err != nil {
		log.Printf("Error saving summary %s: %v", summaryID, err)
	}
}

func (m *Manager) generate(ctx context.Context, project *models.Project, emailID *string) (*Result, error) {
	if emailID != nil {
		email := &models.Email{}
		err := m.db.NewSelect().
			Model(email).
			ExcludeColumn("raw_message").
			Where("id = ?", *emailID).
			Where("project_id = ?", project.ID).
			Scan(ctx)
		if err != nil {
			return nil, fmt.Errorf("load email: %w", err)
		}
		return m.summarizer.Summarize(ctx, email)
	}

	emails := []models.Email{}
	err := m.db.NewSelect().
		Model(&emails).
		ExcludeColumn("raw_message").
		Where("project_id = ?", project.ID).
		Order("received_at DESC").
		Limit(maxProjectEmails + 1). // one extra so the service can tell emails were left out
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("load project emails: %w", err)
	}
	return m.summarizer.SummarizeProject(ctx, project, emails)
}

// failureMessage turns an error into text that is safe to show the user.
func failureMessage(err error) string {
	switch {
	case errors.Is(err, ErrEmptyEmail):
		return "This email has no content to summarise."
	case errors.Is(err, ErrNoEmails):
		return "This project has no emails to summarise."
	default:
		return genericFailureMessage
	}
}
