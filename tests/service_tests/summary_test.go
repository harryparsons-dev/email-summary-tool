package servicetests

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mailai/database/models"
	"mailai/services/summary"

	"github.com/anthropics/anthropic-sdk-go/v2/option"
)

// These match the limits in services/summary, which are not exported.
const (
	summaryModel         = "claude-haiku-5-5"
	maxBodyChars         = 100_000
	maxProjectEmails     = 50
	maxProjectEmailChars = 4_000
)

type capturedRequest struct {
	Model        string `json:"model"`
	OutputConfig struct {
		Effort string `json:"effort"`
	} `json:"output_config"`
	Messages []struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	} `json:"messages"`
}

func (r capturedRequest) userText() string {
	if len(r.Messages) == 0 || len(r.Messages[0].Content) == 0 {
		return ""
	}
	return r.Messages[0].Content[0].Text
}

func messageResponse(stopReason, text string) string {
	response := map[string]any{
		"id":            "msg_test",
		"type":          "message",
		"role":          "assistant",
		"model":         summaryModel,
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"content":       []map[string]any{{"type": "text", "text": text}},
		"usage":         map[string]any{"input_tokens": 10, "output_tokens": 5},
	}
	if stopReason == "refusal" {
		response["stop_details"] = map[string]any{"type": "refusal", "category": "cyber", "explanation": ""}
	}
	encoded, _ := json.Marshal(response)
	return string(encoded)
}

// newTestSummaryService returns a service pointed at a mock API that replies with
// the given body and records the request it received.
func newTestSummaryService(t *testing.T, status int, reply string) (*summary.Service, *capturedRequest) {
	t.Helper()
	captured := &capturedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, captured); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, reply)
	}))
	t.Cleanup(server.Close)

	return summary.NewService(
		option.WithBaseURL(server.URL),
		option.WithAPIKey("test-key"),
		option.WithMaxRetries(0),
	), captured
}

func testEmail() *models.Email {
	return &models.Email{
		FromAddress: "jane@example.com",
		FromName:    "Jane Doe",
		ToAddresses: "abc@mail.test",
		Subject:     "Invoice due Friday",
		TextBody:    "Please pay invoice 42 by Friday.",
		Attachments: []models.EmailAttachment{{Filename: "invoice.pdf"}},
	}
}

func TestSummarizeSendsEmailAndReturnsSummary(t *testing.T) {
	service, captured := newTestSummaryService(t, http.StatusOK, messageResponse("end_turn", "  Jane asks you to pay invoice 42 by Friday. "))

	result, err := service.Summarize(context.Background(), testEmail())
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary != "Jane asks you to pay invoice 42 by Friday." || result.Truncated {
		t.Fatalf("unexpected result %+v", result)
	}

	if captured.Model != "claude-haiku-5-5" {
		t.Errorf("model = %q", captured.Model)
	}
	if captured.OutputConfig.Effort != "low" {
		t.Errorf("effort = %q", captured.OutputConfig.Effort)
	}
	prompt := captured.userText()
	for _, want := range []string{"Jane Doe <jane@example.com>", "Subject: Invoice due Friday", "invoice.pdf", "Please pay invoice 42 by Friday."} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestSummarizeFailures(t *testing.T) {
	tests := []struct {
		name   string
		status int
		reply  string
	}{
		{name: "refusal", status: http.StatusOK, reply: messageResponse("refusal", "")},
		{name: "cut off", status: http.StatusOK, reply: messageResponse("max_tokens", "Jane asks")},
		{name: "empty summary", status: http.StatusOK, reply: messageResponse("end_turn", "   ")},
		{name: "API error", status: http.StatusBadRequest, reply: `{"type":"error","error":{"type":"invalid_request_error","message":"bad"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, _ := newTestSummaryService(t, tt.status, tt.reply)
			if _, err := service.Summarize(context.Background(), testEmail()); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestSummarizeRejectsEmptyEmail(t *testing.T) {
	service, _ := newTestSummaryService(t, http.StatusOK, messageResponse("end_turn", "unused"))

	_, err := service.Summarize(context.Background(), &models.Email{HTMLBody: "<p> </p>"})
	if !errors.Is(err, summary.ErrEmptyEmail) {
		t.Fatalf("err = %v", err)
	}
}

func TestSummarizeTruncatesLongBodies(t *testing.T) {
	service, captured := newTestSummaryService(t, http.StatusOK, messageResponse("end_turn", "Long email."))
	email := testEmail()
	email.TextBody = strings.Repeat("a", maxBodyChars+500)

	result, err := service.Summarize(context.Background(), email)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Truncated {
		t.Error("expected Truncated to be true")
	}
	prompt := captured.userText()
	if strings.Count(prompt, "a") > maxBodyChars+100 {
		t.Error("body was not truncated")
	}
	if !strings.Contains(prompt, "cut short") {
		t.Error("prompt does not tell the model the body was cut short")
	}
}

func TestSummarizeFallsBackToHTMLBody(t *testing.T) {
	service, captured := newTestSummaryService(t, http.StatusOK, messageResponse("end_turn", "Welcome email."))
	email := &models.Email{
		HTMLBody: `<html><head><style>p{color:red}</style></head><body><p>Hello &amp; welcome</p><script>alert(1)</script><p>Second</p></body></html>`,
	}

	if _, err := service.Summarize(context.Background(), email); err != nil {
		t.Fatal(err)
	}
	got := captured.userText()
	if !strings.Contains(got, "Hello & welcome") || !strings.Contains(got, "Second") {
		t.Errorf("missing text: %q", got)
	}
	if strings.Contains(got, "color:red") || strings.Contains(got, "alert") || strings.Contains(got, "<p>") {
		t.Errorf("markup left in text: %q", got)
	}
}

func TestSummarizeProject(t *testing.T) {
	service, captured := newTestSummaryService(t, http.StatusOK, messageResponse("end_turn", "Two invoices are due."))
	project := &models.Project{Title: "Billing", Description: "Supplier invoices"}
	second := testEmail()
	second.Subject = "Second invoice"

	result, err := service.SummarizeProject(context.Background(), project, []models.Email{*testEmail(), *second})
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary != "Two invoices are due." || result.Truncated {
		t.Fatalf("unexpected result %+v", result)
	}
	prompt := captured.userText()
	for _, want := range []string{"Project: Billing", "Supplier invoices", "The 2 most recent emails", "Subject: Invoice due Friday", "Subject: Second invoice"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}

func TestSummarizeProjectLimitsEmails(t *testing.T) {
	service, captured := newTestSummaryService(t, http.StatusOK, messageResponse("end_turn", "Lots of mail."))
	emails := make([]models.Email, maxProjectEmails+1)
	for i := range emails {
		emails[i] = *testEmail()
		emails[i].TextBody = strings.Repeat("b", maxProjectEmailChars+10)
	}

	result, err := service.SummarizeProject(context.Background(), &models.Project{Title: "Busy"}, emails)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Truncated {
		t.Error("expected Truncated to be true")
	}
	prompt := captured.userText()
	if got := strings.Count(prompt, "<email>"); got != maxProjectEmails {
		t.Errorf("prompt has %d emails, want %d", got, maxProjectEmails)
	}
	if strings.Count(prompt, "b") > maxBodyChars+1000 {
		t.Error("project prompt is larger than the overall limit")
	}
}

func TestSummarizeProjectRejectsNoEmails(t *testing.T) {
	service, _ := newTestSummaryService(t, http.StatusOK, messageResponse("end_turn", "unused"))

	_, err := service.SummarizeProject(context.Background(), &models.Project{Title: "Empty"}, nil)
	if !errors.Is(err, summary.ErrNoEmails) {
		t.Fatalf("err = %v", err)
	}
}
