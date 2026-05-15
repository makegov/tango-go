package tango

import (
	"errors"
	"testing"
)

func TestAPIErrorErrorMethod(t *testing.T) {
	e := &APIError{StatusCode: 404, Message: "not found"}
	got := e.Error()
	if got != "tango: not found (status 404)" {
		t.Errorf("unexpected: %q", got)
	}
}

func TestAPIErrorErrorMethodNoStatus(t *testing.T) {
	e := &APIError{Message: "network failure"}
	got := e.Error()
	if got != "tango: network failure" {
		t.Errorf("unexpected: %q", got)
	}
}

func TestAPIErrorNilSafe(t *testing.T) {
	var e *APIError
	got := e.Error()
	if got != "<nil tango.APIError>" {
		t.Errorf("unexpected nil: %q", got)
	}
}

func TestAuthErrorErrorMethod(t *testing.T) {
	e := &AuthError{&APIError{StatusCode: 401, Message: "invalid key"}}
	got := e.Error()
	if got == "" {
		t.Error("expected non-empty Error()")
	}
	// Also test Unwrap
	unwrapped := e.Unwrap()
	if unwrapped == nil {
		t.Error("expected non-nil Unwrap")
	}
	var apiErr *APIError
	if !errors.As(unwrapped, &apiErr) {
		t.Error("expected Unwrap to return *APIError")
	}
}

func TestNotFoundErrorErrorMethod(t *testing.T) {
	e := &NotFoundError{&APIError{StatusCode: 404, Message: "not found"}}
	got := e.Error()
	if got == "" {
		t.Error("expected non-empty Error()")
	}
	unwrapped := e.Unwrap()
	var apiErr *APIError
	if !errors.As(unwrapped, &apiErr) {
		t.Error("expected Unwrap to return *APIError")
	}
}

func TestValidationErrorErrorMethod(t *testing.T) {
	e := &ValidationError{&APIError{StatusCode: 400, Message: "bad request"}}
	got := e.Error()
	if got == "" {
		t.Error("expected non-empty Error()")
	}
}

func TestTimeoutErrorErrorMethod(t *testing.T) {
	e := &TimeoutError{&APIError{Message: "timed out"}}
	got := e.Error()
	if got == "" {
		t.Error("expected non-empty Error()")
	}
	unwrapped := e.Unwrap()
	var apiErr *APIError
	if !errors.As(unwrapped, &apiErr) {
		t.Error("expected Unwrap to return *APIError")
	}
}

func TestRateLimitErrorUnwrap(t *testing.T) {
	e := &RateLimitError{
		APIError:   &APIError{StatusCode: 429, Message: "rate limited"},
		RetryAfter: 5,
		LimitType:  "burst",
	}
	got := e.Error()
	if got == "" {
		t.Error("expected non-empty Error()")
	}
	unwrapped := e.Unwrap()
	var apiErr *APIError
	if !errors.As(unwrapped, &apiErr) {
		t.Error("expected Unwrap to return *APIError")
	}
}

func TestIsRetryableNonAPIError(t *testing.T) {
	// A plain Go error (not *APIError) should not be retryable
	err := errors.New("some other error")
	if IsRetryable(err) {
		t.Error("expected non-APIError to not be retryable")
	}
}

// Regression: callers should be able to errors.As through *APIError.Cause
// to identify the underlying transport/decode failure (F5 from the
// verifier's audit).
func TestAPIErrorUnwrapsCause(t *testing.T) {
	root := errors.New("kaboom")
	apiErr := &APIError{Message: "decode response", Cause: root}

	if !errors.Is(apiErr, root) {
		t.Error("errors.Is should match the Cause")
	}

	// Same traversal through a typed wrapper (404).
	wrapped := &NotFoundError{&APIError{StatusCode: 404, Message: "x", Cause: root}}
	if !errors.Is(wrapped, root) {
		t.Error("errors.Is should traverse *NotFoundError -> *APIError -> Cause")
	}
}

func TestAPIErrorUnwrapNilSafe(t *testing.T) {
	var e *APIError
	if got := e.Unwrap(); got != nil {
		t.Errorf("nil *APIError.Unwrap() should return nil, got %v", got)
	}
	noCause := &APIError{Message: "x"}
	if got := noCause.Unwrap(); got != nil {
		t.Errorf("*APIError with no Cause should Unwrap to nil, got %v", got)
	}
}
