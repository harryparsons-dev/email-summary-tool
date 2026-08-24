package tests

import (
	"email-summary-tool/pkg"
	"net/http"
	"testing"
)

func TestProjects(t *testing.T) {

	cases := []pkg.TestCase{
		{
			TestName: "create a project",
			Request: pkg.Request{
				Method: http.MethodPost,
				Path:   "/projects",
			},
			RequestBody: map[string]interface{}{
				"name":        "Daily summary",
				"description": "Summarise the daily inbox",
				"status":      "pending",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusCreated,
				Body: map[string]any{
					"name":        "Daily summary",
					"description": "Summarise the daily inbox",
					"status":      "pending",
				},
			},
		},
	}

	TestServer.RunAll(t, cases)

}
