package tests

import (
	"net/http"
	"testing"

	"email-summary-tool/pkg"
)

func TestRefresh(t *testing.T) {
	refreshToken, err := TestServer.JWT.SignRefreshToken(TestServer.User.ID)
	if err != nil {
		t.Fatalf("sign refresh token: %v", err)
	}

	cases := []pkg.TestCase{
		{
			TestName: "Refresh returns the signed-in user",
			Request: pkg.Request{
				Method:          http.MethodGet,
				Path:            "/refresh",
				Headers:         map[string]string{"Cookie": "RefreshToken=" + refreshToken},
				Unauthenticated: true,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusOK,
				Body: map[string]any{
					"id":    TestServer.User.ID,
					"email": testUserEmail,
				},
				BodyDoesNotExist: map[string]any{
					"password_hash": TestServer.User.PasswordHash,
				},
			},
		},
		{
			TestName: "Refresh without a refresh token is rejected",
			Request: pkg.Request{
				Method:          http.MethodGet,
				Path:            "/refresh",
				Unauthenticated: true,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusUnauthorized,
				Body:       "Refresh token required",
			},
		},
		{
			TestName: "Refresh with an invalid refresh token is rejected",
			Request: pkg.Request{
				Method:          http.MethodGet,
				Path:            "/refresh",
				Headers:         map[string]string{"Cookie": "RefreshToken=not-a-token"},
				Unauthenticated: true,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusUnauthorized,
				Body:       "Invalid or expired refresh token",
			},
		},
	}

	TestServer.RunAll(t, cases)
}
