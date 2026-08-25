package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"email-summary-tool/pkg"
)

// TestServer is created once and shared by every test file in this package.
var TestServer *APITestServer

func TestMain(m *testing.M) {
	if err := checkTestDatabase(); err != nil {
		fmt.Fprintf(os.Stderr, "check test database: %v\n", err)
		os.Exit(1)
	}

	var err error
	TestServer, err = NewTestServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "start shared test server: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	TestServer.Close()
	os.Exit(code)
}

func (ts *APITestServer) RunAll(t *testing.T, cases []pkg.TestCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.TestName, func(t *testing.T) {
			ts.Run(t, tc)
		})
	}
}

func (ts *APITestServer) Run(t *testing.T, tc pkg.TestCase) {
	t.Helper()

	body, err := requestBody(tc.RequestBody)
	if err != nil {
		t.Fatalf("encode request body: %v", err)
	}

	req, err := http.NewRequest(tc.Request.Method, ts.HTTPServer.URL+tc.Request.Path, body)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if tc.RequestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for name, value := range tc.Request.Headers {
		req.Header.Set(name, value)
	}
	if !tc.Request.Unauthenticated {
		req.AddCookie(ts.AuthCookie)
	}

	response, err := ts.Client.Do(req)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if response.StatusCode != tc.ExpectedResult.StatusCode {
		t.Errorf("status code: got %d, want %d; body: %s", response.StatusCode, tc.ExpectedResult.StatusCode, responseBody)
	}
	assertResponseBody(t, responseBody, tc.ExpectedResult.Body)
	assertResponseBodyDoesNotExist(t, responseBody, tc.ExpectedResult.BodyDoesNotExist)
}

func requestBody(value any) (io.Reader, error) {
	switch value := value.(type) {
	case nil:
		return nil, nil
	case string:
		return strings.NewReader(value), nil
	case []byte:
		return bytes.NewReader(value), nil
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(encoded), nil
	}
}

func assertResponseBody(t *testing.T, actual []byte, expected any) {
	t.Helper()
	if expected == nil {
		return
	}
	if expectedText, ok := expected.(string); ok {
		if !strings.Contains(string(actual), expectedText) {
			t.Errorf("response body %q does not contain %q", actual, expectedText)
		}
		return
	}

	var actualJSON any
	if err := json.Unmarshal(actual, &actualJSON); err != nil {
		t.Errorf("response body is not valid JSON: %v; body: %s", err, actual)
		return
	}
	expectedJSON, err := normalizeJSON(expected)
	if err != nil {
		t.Fatalf("encode expected response body: %v", err)
	}
	if err := matchJSONSubset(actualJSON, expectedJSON, "body"); err != nil {
		t.Errorf("response JSON mismatch: %v; body: %s", err, actual)
	}
}

func assertResponseBodyDoesNotExist(t *testing.T, actual []byte, unexpected any) {
	t.Helper()
	if unexpected == nil {
		return
	}
	if unexpectedText, ok := unexpected.(string); ok {
		if strings.Contains(string(actual), unexpectedText) {
			t.Errorf("response body %q contains unexpected value %q", actual, unexpectedText)
		}
		return
	}

	var actualJSON any
	if err := json.Unmarshal(actual, &actualJSON); err != nil {
		t.Errorf("response body is not valid JSON: %v; body: %s", err, actual)
		return
	}
	unexpectedJSON, err := normalizeJSON(unexpected)
	if err != nil {
		t.Fatalf("encode unexpected response body: %v", err)
	}
	if containsJSONSubset(actualJSON, unexpectedJSON) {
		t.Errorf("response JSON contains unexpected value %#v; body: %s", unexpected, actual)
	}
}

func containsJSONSubset(actual, unexpected any) bool {
	if matchJSONSubset(actual, unexpected, "body") == nil {
		return true
	}

	switch actual := actual.(type) {
	case map[string]any:
		for _, value := range actual {
			if containsJSONSubset(value, unexpected) {
				return true
			}
		}
	case []any:
		for _, value := range actual {
			if containsJSONSubset(value, unexpected) {
				return true
			}
		}
	}

	return false
}

func normalizeJSON(value any) (any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var normalized any
	if err := json.Unmarshal(encoded, &normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

func matchJSONSubset(actual, expected any, path string) error {
	switch expected := expected.(type) {
	case map[string]any:
		actualObject, ok := actual.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: got %T, want object", path, actual)
		}
		for key, expectedValue := range expected {
			actualValue, exists := actualObject[key]
			if !exists {
				return fmt.Errorf("%s.%s: missing key", path, key)
			}
			if err := matchJSONSubset(actualValue, expectedValue, path+"."+key); err != nil {
				return err
			}
		}
		return nil
	case []any:
		actualArray, ok := actual.([]any)
		if !ok {
			return fmt.Errorf("%s: got %T, want array", path, actual)
		}
		if len(actualArray) != len(expected) {
			return fmt.Errorf("%s: got array length %d, want %d", path, len(actualArray), len(expected))
		}
		for index := range expected {
			if err := matchJSONSubset(actualArray[index], expected[index], fmt.Sprintf("%s[%d]", path, index)); err != nil {
				return err
			}
		}
		return nil
	default:
		if !reflect.DeepEqual(actual, expected) {
			return fmt.Errorf("%s: got %#v, want %#v", path, actual, expected)
		}
		return nil
	}
}
