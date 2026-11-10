package tests

import (
	"strings"
	"testing"

	"email-summary-tool/database/factories"
	"email-summary-tool/database/models"
	"email-summary-tool/services/inbound"
)

func TestInboundEmailServiceFindProject(t *testing.T) {
	service := inbound.NewInboundEmailService(TestServer.Server)
	// A separate user keeps these projects out of the shared user's project list.
	user := &models.User{Email: "inbound-find-user@example.com"}
	if _, err := TestServer.Db.NewInsert().Model(user).Exec(t.Context()); err != nil {
		t.Fatalf("create user fixture: %v", err)
	}
	project, err := factories.NewProject(t.Context(), TestServer.Db, user.ID)
	if err != nil {
		t.Fatalf("create project fixture: %v", err)
	}
	localPart, _, _ := strings.Cut(project.EmailAddress, "@")

	cases := []struct {
		name          string
		address       string
		wantProjectID string
	}{
		{"exact address", project.EmailAddress, project.ID},
		{"mixed case in angle brackets", "<" + strings.ToUpper(project.EmailAddress) + ">", project.ID},
		{"unknown address", "nobody@example.com", ""},
		{"other domain", localPart + "@other.example", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			found, err := service.FindProject(t.Context(), tc.address)
			if err != nil {
				t.Fatalf("FindProject() error = %v", err)
			}
			projectID := ""
			if found != nil {
				projectID = found.ID
			}
			if projectID != tc.wantProjectID {
				t.Fatalf("FindProject() project ID = %q, want %q", projectID, tc.wantProjectID)
			}
		})
	}
}

func TestInboundEmailServiceDeliver(t *testing.T) {
	service := inbound.NewInboundEmailService(TestServer.Server)
	// A separate user keeps these projects out of the shared user's project list.
	user := &models.User{Email: "inbound-deliver-user@example.com"}
	if _, err := TestServer.Db.NewInsert().Model(user).Exec(t.Context()); err != nil {
		t.Fatalf("create user fixture: %v", err)
	}
	project, err := factories.NewProject(t.Context(), TestServer.Db, user.ID)
	if err != nil {
		t.Fatalf("create project fixture: %v", err)
	}

	raw := []byte("From: Sender <sender@example.org>\r\nTo: " + project.EmailAddress +
		"\r\nSubject: Delivered via service\r\n\r\nHello\r\n")
	recipients := []inbound.Recipient{
		{Address: project.EmailAddress, ProjectID: project.ID},
		// The same project twice should still store one copy.
		{Address: strings.ToUpper(project.EmailAddress), ProjectID: project.ID},
	}
	if err := service.Deliver(t.Context(), "sender@example.org", recipients, raw); err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}

	var emails []models.Email
	if err := TestServer.Db.NewSelect().Model(&emails).Where("project_id = ?", project.ID).Scan(t.Context()); err != nil {
		t.Fatalf("load delivered emails: %v", err)
	}
	if len(emails) != 1 {
		t.Fatalf("delivered %d emails, want 1", len(emails))
	}
	if emails[0].Subject != "Delivered via service" || emails[0].FromAddress != "sender@example.org" {
		t.Fatalf("delivered email = subject %q from %q", emails[0].Subject, emails[0].FromAddress)
	}
}
