package tests

import (
	"net/http"
	"strings"
	"testing"

	"email-summary-tool/database/models"
	"email-summary-tool/pkg"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestSignUp(t *testing.T) {
	// The test database outlives a single run, so each run signs up a new address.
	email := "signup-" + uuid.NewString() + "@example.com"

	request := pkg.Request{
		Method:          http.MethodPost,
		Path:            "/signup",
		Unauthenticated: true,
	}

	cases := []pkg.TestCase{
		{
			TestName: "Cannot sign up without an email and password",
			Request:  request,
			RequestBody: map[string]any{
				"email":    "",
				"password": "",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusBadRequest,
				Body: map[string]any{
					"error": "Email and password are required",
				},
			},
		},
		{
			TestName: "Cannot sign up with an invalid email address",
			Request:  request,
			RequestBody: map[string]any{
				"email":    "not-an-email",
				"password": "correct-horse-battery",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusBadRequest,
				Body: map[string]any{
					"error": "Enter a valid email address",
				},
			},
		},
		{
			TestName: "Cannot sign up with a short password",
			Request:  request,
			RequestBody: map[string]any{
				"email":    email,
				"password": "short",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusBadRequest,
				Body: map[string]any{
					"error": "Password must be at least 8 characters",
				},
			},
		},
		{
			TestName: "Cannot sign up with a password longer than bcrypt accepts",
			Request:  request,
			RequestBody: map[string]any{
				"email":    email,
				"password": strings.Repeat("p", 73),
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusBadRequest,
				Body: map[string]any{
					"error": "Password must be at most 72 bytes",
				},
			},
		},
		{
			TestName: "Can sign up",
			Request:  request,
			RequestBody: map[string]any{
				"email":    email,
				"password": "correct-horse-battery",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusCreated,
				Body: map[string]any{
					"message": "Signed up successfully",
				},
			},
		},
		{
			TestName: "Cannot sign up twice with the same email",
			Request:  request,
			RequestBody: map[string]any{
				"email":    email,
				"password": "correct-horse-battery",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusBadRequest,
				Body: map[string]any{
					"error": "User with this email already exists",
				},
			},
		},
	}

	TestServer.RunAll(t, cases)
}

func TestLogin(t *testing.T) {
	email := "login-" + uuid.NewString() + "@example.com"
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-horse-battery"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password fixture: %v", err)
	}
	user := &models.User{Email: email, PasswordHash: string(hash)}
	if _, err := TestServer.Db.NewInsert().Model(user).Exec(t.Context()); err != nil {
		t.Fatalf("create user fixture: %v", err)
	}

	request := pkg.Request{
		Method:          http.MethodPost,
		Path:            "/login",
		Unauthenticated: true,
	}

	cases := []pkg.TestCase{
		{
			TestName: "Can log in",
			Request:  request,
			RequestBody: map[string]any{
				"email":    email,
				"password": "correct-horse-battery",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusOK,
				Body: map[string]any{
					"message": "Logged in successfully",
				},
			},
		},
		{
			TestName: "Cannot log in with the wrong password",
			Request:  request,
			RequestBody: map[string]any{
				"email":    email,
				"password": "wrong-password",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusUnauthorized,
				Body: map[string]any{
					"error": "Invalid email or password",
				},
			},
		},
		{
			TestName: "Cannot log in with an unknown email",
			Request:  request,
			RequestBody: map[string]any{
				"email":    "unknown-" + email,
				"password": "correct-horse-battery",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusUnauthorized,
				Body: map[string]any{
					"error": "Invalid email or password",
				},
			},
		},
		{
			TestName: "Cannot log in without an email and password",
			Request:  request,
			RequestBody: map[string]any{
				"email":    "",
				"password": "",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusBadRequest,
				Body: map[string]any{
					"error": "Email and password are required",
				},
			},
		},
	}

	TestServer.RunAll(t, cases)
}

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
