package tango

import (
	"errors"
	"fmt"
)

// APIError is the base error type returned by the SDK for any HTTP-layer
// failure. Specific error types embed *APIError, so callers can use
// errors.As(err, &tango.AuthError{}) for the specific cases, or
// errors.As(err, &tango.APIError{}) for the catch-all.
type APIError struct {
	// StatusCode is the HTTP status code, or 0 for transport-level errors
	// (timeouts, DNS, connection refused).
	StatusCode int

	// Message is a human-readable description.
	Message string

	// ResponseData is the decoded JSON body of the error response, when
	// present. It is typed as any because the server returns a mix of
	// shapes (object, array, string, null).
	ResponseData any

	// Cause is the underlying error, if any (e.g. a *json.SyntaxError
	// from a malformed response body or a *net.OpError from a network
	// failure). Preserved so callers can errors.As/Is through it.
	Cause error
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e == nil {
		return "<nil tango.APIError>"
	}
	if e.StatusCode != 0 {
		return fmt.Sprintf("tango: %s (status %d)", e.Message, e.StatusCode)
	}
	return "tango: " + e.Message
}

// Unwrap returns the underlying cause, if any, so errors.Is / errors.As
// can traverse to wrapped errors (e.g. *json.SyntaxError).
func (e *APIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// AuthError is raised on HTTP 401.
type AuthError struct{ *APIError }

// Error implements the error interface.
func (e *AuthError) Error() string { return e.APIError.Error() }

// Unwrap returns the embedded *APIError so errors.As / errors.Is can
// traverse to the base error type.
func (e *AuthError) Unwrap() error { return e.APIError }

// NotFoundError is raised on HTTP 404.
type NotFoundError struct{ *APIError }

// Error implements the error interface.
func (e *NotFoundError) Error() string { return e.APIError.Error() }

// Unwrap returns the embedded *APIError so errors.As / errors.Is can
// traverse to the base error type.
func (e *NotFoundError) Unwrap() error { return e.APIError }

// ValidationError is raised on HTTP 400 (or other client-supplied invalid
// input that the SDK rejects locally).
type ValidationError struct{ *APIError }

// Error implements the error interface.
func (e *ValidationError) Error() string { return e.APIError.Error() }

// Unwrap returns the embedded *APIError so errors.As / errors.Is can
// traverse to the base error type.
func (e *ValidationError) Unwrap() error { return e.APIError }

// RateLimitError is raised on HTTP 429. RetryAfter is populated from the
// Retry-After header when present (in seconds).
type RateLimitError struct {
	*APIError
	// RetryAfter is the server-suggested wait before retrying, in seconds.
	// 0 when the Retry-After header is absent or unparseable.
	RetryAfter int
	// LimitType is the value of X-RateLimit-Type when set by the server
	// (used to distinguish per-minute / per-hour / per-day buckets). Empty
	// when absent.
	LimitType string
}

// Error implements the error interface.
func (e *RateLimitError) Error() string { return e.APIError.Error() }

// Unwrap returns the embedded *APIError so errors.As / errors.Is can
// traverse to the base error type.
func (e *RateLimitError) Unwrap() error { return e.APIError }

// TimeoutError is raised when a request exceeds the configured timeout.
// StatusCode is 0.
type TimeoutError struct{ *APIError }

// Error implements the error interface.
func (e *TimeoutError) Error() string { return e.APIError.Error() }

// Unwrap returns the embedded *APIError so errors.As / errors.Is can
// traverse to the base error type.
func (e *TimeoutError) Unwrap() error { return e.APIError }

// IsRetryable reports whether the SDK should retry after this error. Used
// internally by the transport's retry loop; exposed because external callers
// occasionally want to drive their own retry policies.
func IsRetryable(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.StatusCode == 0 {
		// Network / timeout — retryable.
		return true
	}
	if apiErr.StatusCode == 408 || apiErr.StatusCode == 429 {
		return true
	}
	if apiErr.StatusCode >= 500 && apiErr.StatusCode < 600 {
		return true
	}
	return false
}
