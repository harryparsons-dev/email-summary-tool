// Package mocks holds stand-ins for external dependencies used in tests.
package mocks

import (
	"context"
	"fmt"

	"mailai/database/models"
	"mailai/services/summary"
)

// MockSummarizer stands in for the Claude API so tests are free and
// predictable. Tests set SummarizeFunc or SummarizeProjectFunc to control the
// result, such as returning an error; when a func is nil a default summary is
// returned.
type MockSummarizer struct {
	SummarizeFunc        func(ctx context.Context, email *models.Email) (*summary.Result, error)
	SummarizeProjectFunc func(ctx context.Context, project *models.Project, emails []models.Email) (*summary.Result, error)
}

func (m *MockSummarizer) Summarize(ctx context.Context, email *models.Email) (*summary.Result, error) {
	if m.SummarizeFunc != nil {
		return m.SummarizeFunc(ctx, email)
	}
	return &summary.Result{Summary: "Summary of " + email.Subject}, nil
}

func (m *MockSummarizer) SummarizeProject(ctx context.Context, project *models.Project, emails []models.Email) (*summary.Result, error) {
	if m.SummarizeProjectFunc != nil {
		return m.SummarizeProjectFunc(ctx, project, emails)
	}
	return &summary.Result{Summary: fmt.Sprintf("Summary of %d emails in %s", len(emails), project.Title)}, nil
}
