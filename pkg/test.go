package pkg

type Request struct {
	Method          string
	Path            string
	Headers         map[string]string
	Unauthenticated bool
}

type ExpectedResult struct {
	StatusCode int
	// Body can be a string to match response text or a JSON-compatible value.
	// JSON objects are subset-matched, so generated fields can be omitted.
	Body any
	// BodyDoesNotExist fails when a string or JSON subset appears anywhere in the response.
	BodyDoesNotExist any
}

type TestCase struct {
	TestName       string
	Request        Request
	RequestBody    any
	ExpectedResult ExpectedResult
}
