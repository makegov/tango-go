// Package webhooks provides HMAC-SHA256 signing and verification for
// Tango webhook deliveries.
//
// Tango signs each delivery with:
//
//	X-Tango-Signature: sha256=<lowercase hex HMAC-SHA256 of raw body>
//
// Verify against the **raw request body** — re-serializing parsed JSON
// will produce a different signature.
package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const (
	// SignatureHeader is the HTTP header name Tango uses to sign deliveries.
	SignatureHeader = "X-Tango-Signature"

	// SignaturePrefix is the algorithm prefix on the header value.
	SignaturePrefix = "sha256="
)

// Generate returns the wire-form signature for body+secret.
// Example return: "sha256=4f3c1a…".
func Generate(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return SignaturePrefix + hex.EncodeToString(mac.Sum(nil))
}

// ParsedSignature is the decomposed form of an X-Tango-Signature header.
type ParsedSignature struct {
	Algorithm string // always "sha256" today
	Signature string // lowercase hex digest
}

// Parse decomposes an X-Tango-Signature header value. Accepts both the
// canonical "sha256=<hex>" form and a bare hex string (legacy), in which
// case Algorithm defaults to "sha256". Returns false for empty,
// malformed, or non-hex inputs.
func Parse(header string) (ParsedSignature, bool) {
	stripped := strings.TrimSpace(header)
	if stripped == "" {
		return ParsedSignature{}, false
	}

	var alg, sig string
	if i := strings.Index(stripped, "="); i > 0 {
		alg = strings.ToLower(stripped[:i])
		sig = stripped[i+1:]
	} else {
		alg = "sha256"
		sig = stripped
	}

	if sig == "" {
		return ParsedSignature{}, false
	}
	if !isHex(sig) {
		return ParsedSignature{}, false
	}
	return ParsedSignature{Algorithm: alg, Signature: strings.ToLower(sig)}, true
}

// Verify reports whether header is a valid Tango signature for body+secret.
// Returns false for absent, malformed, or mismatched headers — never panics.
// Comparison is constant-time via crypto/hmac.Equal.
func Verify(body []byte, header, secret string) bool {
	parsed, ok := Parse(header)
	if !ok {
		return false
	}
	if parsed.Algorithm != "sha256" {
		return false
	}
	expectedHex := strings.TrimPrefix(Generate(body, secret), SignaturePrefix)
	if len(expectedHex) != len(parsed.Signature) {
		return false
	}
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		return false
	}
	actual, err := hex.DecodeString(parsed.Signature)
	if err != nil {
		return false
	}
	return hmac.Equal(expected, actual)
}

func isHex(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
