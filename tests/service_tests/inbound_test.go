package servicetests

import (
	"strings"
	"testing"

	"email-summary-tool/services/inbound"
)

func TestGenerateAddress(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		address, err := inbound.GenerateAddress("Mail.Example.com")
		if err != nil {
			t.Fatal(err)
		}
		local, domain, ok := strings.Cut(address, "@")
		if !ok || domain != "mail.example.com" {
			t.Fatalf("unexpected address %q", address)
		}
		if len(local) != 12 || strings.Trim(local, "abcdefghijkmnpqrstuvwxyz23456789") != "" {
			t.Fatalf("unexpected local part %q", local)
		}
		if seen[address] {
			t.Fatalf("duplicate address %q", address)
		}
		seen[address] = true
	}
}

func TestNormalizeAddress(t *testing.T) {
	if got := inbound.NormalizeAddress(" <Abc@Example.COM> "); got != "abc@example.com" {
		t.Fatalf("got %q", got)
	}
	if got := inbound.AddressDomain("a@b@example.com"); got != "example.com" {
		t.Fatalf("got %q", got)
	}
}

func crlf(s string) []byte {
	return []byte(strings.ReplaceAll(s, "\n", "\r\n"))
}

func TestParseMultipart(t *testing.T) {
	raw := crlf(`From: "Jane Doe" <Jane@Example.com>
To: abc@mail.test, other@example.com
Subject: =?UTF-8?Q?Caf=C3=A9_update?=
Date: Mon, 02 Nov 2026 10:00:00 +0000
Message-ID: <msg-1@example.com>
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="outer"

--outer
Content-Type: multipart/alternative; boundary="inner"

--inner
Content-Type: text/plain; charset=iso-8859-1
Content-Transfer-Encoding: quoted-printable

Hello caf=E9
--inner
Content-Type: text/html; charset=utf-8

<p>Hello</p>
--inner--
--outer
Content-Type: application/pdf
Content-Disposition: attachment; filename="report.pdf"
Content-Transfer-Encoding: base64

aGVsbG8=
--outer--
`)

	parsed, err := inbound.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.FromAddress != "jane@example.com" || parsed.FromName != "Jane Doe" {
		t.Errorf("from = %q %q", parsed.FromAddress, parsed.FromName)
	}
	if parsed.ToAddresses != "abc@mail.test, other@example.com" {
		t.Errorf("to = %q", parsed.ToAddresses)
	}
	if parsed.Subject != "Café update" {
		t.Errorf("subject = %q", parsed.Subject)
	}
	if parsed.MessageID != "msg-1@example.com" {
		t.Errorf("message id = %q", parsed.MessageID)
	}
	if parsed.SentAt == nil || parsed.SentAt.Year() != 2026 {
		t.Errorf("sent at = %v", parsed.SentAt)
	}
	if strings.TrimSpace(parsed.TextBody) != "Hello café" {
		t.Errorf("text body = %q", parsed.TextBody)
	}
	if strings.TrimSpace(parsed.HTMLBody) != "<p>Hello</p>" {
		t.Errorf("html body = %q", parsed.HTMLBody)
	}
	if len(parsed.Attachments) != 1 {
		t.Fatalf("attachments = %+v", parsed.Attachments)
	}
	if a := parsed.Attachments[0]; a.Filename != "report.pdf" || a.ContentType != "application/pdf" || a.Size != 5 {
		t.Errorf("attachment = %+v", a)
	}
}

func TestParsePlainMessage(t *testing.T) {
	parsed, err := inbound.Parse(crlf("From: a@example.com\nSubject: hi\n\nplain body\x00\n"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(parsed.TextBody) != "plain body" {
		t.Errorf("text body = %q", parsed.TextBody)
	}
	if parsed.Attachments == nil {
		t.Error("attachments should be an empty slice, not nil")
	}
}

func TestParseGarbageKeepsGoing(t *testing.T) {
	parsed, _ := inbound.Parse([]byte("not an email at all"))
	if parsed == nil {
		t.Fatal("expected a non-nil result")
	}
}
