package tests

import (
	"net/http"
	"testing"

	"email-summary-tool/database/factories"
	"email-summary-tool/database/models"
	"email-summary-tool/pkg"
	"email-summary-tool/services/summary"

	"github.com/google/uuid"
)

// latestEmailSummary returns the newest summary of an email.
func latestEmailSummary(t *testing.T, emailID string) *models.Summary {
	t.Helper()
	summary := &models.Summary{}
	err := TestServer.Db.NewSelect().
		Model(summary).
		Where("s.email_id = ?", emailID).
		OrderExpr("s.created_at DESC").
		Limit(1).
		Scan(t.Context())
	if err != nil {
		t.Fatalf("load latest summary: %v", err)
	}
	return summary
}

func TestEmailSummariesCreate(t *testing.T) {
	project, err := factories.NewProject(t.Context(), TestServer.Db, TestServer.User.ID)
	if err != nil {
		t.Fatalf("create project fixture: %v", err)
	}
	email, err := factories.NewEmail(t.Context(), TestServer.Db, project.ID, "Invoice due Friday")
	if err != nil {
		t.Fatalf("create email fixture: %v", err)
	}

	otherProject, err := factories.NewProject(t.Context(), TestServer.Db, TestServer.User.ID)
	if err != nil {
		t.Fatalf("create other project fixture: %v", err)
	}
	otherProjectEmail, err := factories.NewEmail(t.Context(), TestServer.Db, otherProject.ID, "Other project email")
	if err != nil {
		t.Fatalf("create other project email fixture: %v", err)
	}

	otherUser := &models.User{Email: "summary-other-user@example.com"}
	if _, err := TestServer.Db.NewInsert().Model(otherUser).Exec(t.Context()); err != nil {
		t.Fatalf("create other user fixture: %v", err)
	}
	otherUsersProject, err := factories.NewProject(t.Context(), TestServer.Db, otherUser.ID)
	if err != nil {
		t.Fatalf("create other user's project fixture: %v", err)
	}
	otherUsersEmail, err := factories.NewEmail(t.Context(), TestServer.Db, otherUsersProject.ID, "Private email")
	if err != nil {
		t.Fatalf("create other user's email fixture: %v", err)
	}

	cases := []pkg.TestCase{
		{
			TestName: "Can start an email summary",
			Request: pkg.Request{
				Method: http.MethodPost,
				Path:   "/projects/" + project.ID + "/emails/" + email.ID + "/summarise",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusAccepted,
				Body: map[string]any{
					"project_id": project.ID,
					"email_id":   email.ID,
					"status":     "pending",
					"summary":    "",
				},
			},
		},
		{
			TestName: "Cannot summarise an email from another project",
			Request: pkg.Request{
				Method: http.MethodPost,
				Path:   "/projects/" + project.ID + "/emails/" + otherProjectEmail.ID + "/summarise",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Email not found",
			},
		},
		{
			TestName: "Cannot summarise an email that does not exist",
			Request: pkg.Request{
				Method: http.MethodPost,
				Path:   "/projects/" + project.ID + "/emails/" + uuid.NewString() + "/summarise",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Email not found",
			},
		},
		{
			TestName: "Cannot summarise an email with an invalid ID",
			Request: pkg.Request{
				Method: http.MethodPost,
				Path:   "/projects/" + project.ID + "/emails/not-a-uuid/summarise",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Email not found",
			},
		},
		{
			TestName: "Cannot summarise another user's email",
			Request: pkg.Request{
				Method: http.MethodPost,
				Path:   "/projects/" + otherUsersProject.ID + "/emails/" + otherUsersEmail.ID + "/summarise",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Project not found",
			},
		},
		{
			TestName: "Cannot summarise without authentication",
			Request: pkg.Request{
				Method:          http.MethodPost,
				Path:            "/projects/" + project.ID + "/emails/" + email.ID + "/summarise",
				Unauthenticated: true,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusUnauthorized,
				Body:       "Authentication required",
			},
		},
	}

	TestServer.RunAll(t, cases)
	TestServer.Background.Wait()
}

func TestEmailSummariesCreateReturnsPendingSummary(t *testing.T) {
	project, err := factories.NewProject(t.Context(), TestServer.Db, TestServer.User.ID)
	if err != nil {
		t.Fatalf("create project fixture: %v", err)
	}
	email, err := factories.NewEmail(t.Context(), TestServer.Db, project.ID, "Weekly update")
	if err != nil {
		t.Fatalf("create email fixture: %v", err)
	}

	pending := &models.Summary{
		UserId:    TestServer.User.ID,
		ProjectId: project.ID,
		EmailId:   &email.ID,
		Status:    models.SummaryStatusPending,
	}
	if _, err := TestServer.Db.NewInsert().Model(pending).Exec(t.Context()); err != nil {
		t.Fatalf("create pending summary fixture: %v", err)
	}

	cases := []pkg.TestCase{
		{
			TestName: "Returns the summary already in progress instead of starting another",
			Request: pkg.Request{
				Method: http.MethodPost,
				Path:   "/projects/" + project.ID + "/emails/" + email.ID + "/summarise",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusOK,
				Body: map[string]any{
					"id":       pending.ID,
					"email_id": email.ID,
					"status":   "pending",
				},
			},
		},
	}

	TestServer.RunAll(t, cases)
}

func TestEmailSummariesGet(t *testing.T) {
	project, err := factories.NewProject(t.Context(), TestServer.Db, TestServer.User.ID)
	if err != nil {
		t.Fatalf("create project fixture: %v", err)
	}
	email, err := factories.NewEmail(t.Context(), TestServer.Db, project.ID, "Invoice due Friday")
	if err != nil {
		t.Fatalf("create email fixture: %v", err)
	}
	failingEmail, err := factories.NewEmail(t.Context(), TestServer.Db, project.ID, failSummarySubject)
	if err != nil {
		t.Fatalf("create failing email fixture: %v", err)
	}

	summaries := summary.NewManager(TestServer.Db, TestServer.Summarizer)
	for _, emailID := range []string{email.ID, failingEmail.ID} {
		emailSummary, _, err := summaries.Create(t.Context(), TestServer.User.ID, project, &emailID)
		if err != nil {
			t.Fatalf("create summary: %v", err)
		}
		summaries.Generate(emailSummary.ID, project, &emailID)
	}

	emailSummary := latestEmailSummary(t, email.ID)
	failedSummary := latestEmailSummary(t, failingEmail.ID)

	otherProject, err := factories.NewProject(t.Context(), TestServer.Db, TestServer.User.ID)
	if err != nil {
		t.Fatalf("create other project fixture: %v", err)
	}

	cases := []pkg.TestCase{
		{
			TestName: "Can get a completed email summary",
			Request: pkg.Request{
				Method: http.MethodGet,
				Path:   "/projects/" + project.ID + "/emails/" + email.ID + "/summarise/" + emailSummary.ID,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusOK,
				Body: map[string]any{
					"id":         emailSummary.ID,
					"project_id": project.ID,
					"email_id":   email.ID,
					"status":     "completed",
					"summary":    "Summary of Invoice due Friday",
					"error":      "",
				},
			},
		},
		{
			TestName: "Can get a failed email summary",
			Request: pkg.Request{
				Method: http.MethodGet,
				Path:   "/projects/" + project.ID + "/emails/" + failingEmail.ID + "/summarise/" + failedSummary.ID,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusOK,
				Body: map[string]any{
					"id":      failedSummary.ID,
					"status":  "failed",
					"summary": "",
					"error":   "The summary could not be generated. Please try again.",
				},
				BodyDoesNotExist: "fake summarizer failure",
			},
		},
		{
			TestName: "Cannot get a summary through a different email",
			Request: pkg.Request{
				Method: http.MethodGet,
				Path:   "/projects/" + project.ID + "/emails/" + failingEmail.ID + "/summarise/" + emailSummary.ID,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Summary not found",
			},
		},
		{
			TestName: "Cannot get a summary through a different project",
			Request: pkg.Request{
				Method: http.MethodGet,
				Path:   "/projects/" + otherProject.ID + "/emails/" + email.ID + "/summarise/" + emailSummary.ID,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Summary not found",
			},
		},
		{
			TestName: "Cannot get a summary that does not exist",
			Request: pkg.Request{
				Method: http.MethodGet,
				Path:   "/projects/" + project.ID + "/emails/" + email.ID + "/summarise/" + uuid.NewString(),
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Summary not found",
			},
		},
		{
			TestName: "Cannot get a summary without authentication",
			Request: pkg.Request{
				Method:          http.MethodGet,
				Path:            "/projects/" + project.ID + "/emails/" + email.ID + "/summarise/" + emailSummary.ID,
				Unauthenticated: true,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusUnauthorized,
				Body:       "Authentication required",
			},
		},
	}

	TestServer.RunAll(t, cases)
}

func TestEmailSummariesList(t *testing.T) {
	project, err := factories.NewProject(t.Context(), TestServer.Db, TestServer.User.ID)
	if err != nil {
		t.Fatalf("create project fixture: %v", err)
	}
	email, err := factories.NewEmail(t.Context(), TestServer.Db, project.ID, "Invoice due Friday")
	if err != nil {
		t.Fatalf("create email fixture: %v", err)
	}
	otherEmail, err := factories.NewEmail(t.Context(), TestServer.Db, project.ID, "Other email")
	if err != nil {
		t.Fatalf("create other email fixture: %v", err)
	}

	summaries := summary.NewManager(TestServer.Db, TestServer.Summarizer)
	for _, emailID := range []string{email.ID, otherEmail.ID} {
		emailSummary, _, err := summaries.Create(t.Context(), TestServer.User.ID, project, &emailID)
		if err != nil {
			t.Fatalf("create summary: %v", err)
		}
		summaries.Generate(emailSummary.ID, project, &emailID)
	}

	cases := []pkg.TestCase{
		{
			TestName: "Can list an email's summaries",
			Request: pkg.Request{
				Method: http.MethodGet,
				Path:   "/projects/" + project.ID + "/emails/" + email.ID + "/summarise",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusOK,
				Body: []any{
					map[string]any{
						"email_id": email.ID,
						"status":   "completed",
						"summary":  "Summary of Invoice due Friday",
					},
				},
				BodyDoesNotExist: map[string]any{
					"email_id": otherEmail.ID,
				},
			},
		},
		{
			TestName: "Cannot list summaries with an invalid email ID",
			Request: pkg.Request{
				Method: http.MethodGet,
				Path:   "/projects/" + project.ID + "/emails/not-a-uuid/summarise",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Email not found",
			},
		},
		{
			TestName: "Cannot list summaries without authentication",
			Request: pkg.Request{
				Method:          http.MethodGet,
				Path:            "/projects/" + project.ID + "/emails/" + email.ID + "/summarise",
				Unauthenticated: true,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusUnauthorized,
				Body:       "Authentication required",
			},
		},
	}

	TestServer.RunAll(t, cases)
}
