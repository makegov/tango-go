package webhooks

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

// ErrInvalidSignature is returned by VerifyRequest when the request's
// X-Tango-Signature header is missing or doesn't match the computed HMAC.
var ErrInvalidSignature = errors.New("tango/webhooks: invalid signature")

// VerifyRequest reads and verifies the body of a webhook request. On
// success it returns the raw body and resets r.Body so downstream handlers
// can re-read it. On failure it returns ErrInvalidSignature (or the read
// error). The caller is responsible for sending an appropriate HTTP
// response — VerifyRequest doesn't write to w.
//
// Typical usage in a handler:
//
//	body, err := webhooks.VerifyRequest(r, secret)
//	if err != nil {
//	    http.Error(w, "invalid signature", http.StatusUnauthorized)
//	    return
//	}
//	// decode body, process delivery, ...
func VerifyRequest(r *http.Request, secret string) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	_ = r.Body.Close()
	// Reset so downstream handlers / decoders can re-read.
	r.Body = io.NopCloser(bytes.NewReader(body))

	if !Verify(body, r.Header.Get(SignatureHeader), secret) {
		return nil, ErrInvalidSignature
	}
	return body, nil
}

// Middleware wraps next so that requests with an invalid Tango signature
// are rejected with 401 before reaching next. The verified body is left
// readable on r.Body.
func Middleware(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := VerifyRequest(r, secret); err != nil {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
