package models

import "github.com/uptrace/bun"

type SummaryStatus string

const (
	SummaryStatusPending   SummaryStatus = "pending"
	SummaryStatusCompleted SummaryStatus = "completed"
	SummaryStatusFailed    SummaryStatus = "failed"
)

// Summary is an AI-generated summary. EmailId is nil when the summary covers
// the whole project rather than a single email.
//
// Summaries are generated in the background, so a new row starts as pending
// and later becomes completed (Summary set) or failed (Error set).
type Summary struct {
	bun.BaseModel `bun:"table:summaries,alias:s"`
	UuidBase

	UserId    string        `json:"user_id"`
	ProjectId string        `json:"project_id"`
	EmailId   *string       `json:"email_id"`
	Status    SummaryStatus `json:"status"`
	Summary   string        `json:"summary"`
	Error     string        `json:"error"`
}

// IsProjectSummary reports whether the summary covers the whole project.
func (s *Summary) IsProjectSummary() bool {
	return s.EmailId == nil
}
