package tests

import (
	"context"
	"errors"
	"fmt"

	"email-summary-tool/database/models"
	"email-summary-tool/services/summary"
)

// failSummarySubject makes fakeSummarizer fail, to test failed summaries.
const failSummarySubject = "make the summary fail"

// fakeSummarizer stands in for the Claude API so tests are free and
// predictable.
type fakeSummarizer struct{}

func (fakeSummarizer) Summarize(_ context.Context, email *models.Email) (*summary.Result, error) {
	if email.Subject == failSummarySubject {
		return nil, errors.New("fake summarizer failure")
	}
	return &summary.Result{Summary: "Summary of " + email.Subject}, nil
}

func (fakeSummarizer) SummarizeProject(_ context.Context, project *models.Project, emails []models.Email) (*summary.Result, error) {
	return &summary.Result{Summary: fmt.Sprintf("Summary of %d emails in %s", len(emails), project.Title)}, nil
}
