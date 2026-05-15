package webhooks

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateMatchesKnownVector(t *testing.T) {
	// A known HMAC-SHA256 of "hello" with secret "shh".
	// Computed externally; locks the implementation to the canonical algorithm.
	got := Generate([]byte("hello"), "shh")
	const want = "sha256=0e396369ee043c5b6b922743631745b2249cf7cb2c4722e61e802447d5d14c70"
	if got != want {
		t.Errorf("Generate mismatch:\n got  %s\n want %s", got, want)
	}
}

func TestVerifyRoundtrip(t *testing.T) {
	body := []byte(`{"event":"contract.updated","id":"123"}`)
	header := Generate(body, "topsecret")
	if !Verify(body, header, "topsecret") {
		t.Error("expected Verify to accept a matching signature")
	}
	if Verify(body, header, "wrong-secret") {
		t.Error("Verify accepted a signature with the wrong secret")
	}
	if Verify([]byte("tampered"), header, "topsecret") {
		t.Error("Verify accepted a signature for a different body")
	}
}

func TestVerifyRejectsMalformed(t *testing.T) {
	body := []byte("body")
	cases := []string{"", "    ", "sha256=", "sha256=zzz", "md5=abc", "not-hex"}
	for _, h := range cases {
		if Verify(body, h, "s") {
			t.Errorf("Verify accepted malformed header %q", h)
		}
	}
}

func TestParseBareHexDefaultsToSha256(t *testing.T) {
	body := []byte("body")
	withPrefix := Generate(body, "s")
	bare := strings.TrimPrefix(withPrefix, SignaturePrefix)
	parsed, ok := Parse(bare)
	if !ok {
		t.Fatal("expected Parse to accept bare hex")
	}
	if parsed.Algorithm != "sha256" {
		t.Errorf("expected default alg sha256, got %q", parsed.Algorithm)
	}
	if !Verify(body, bare, "s") {
		t.Error("Verify should accept bare-hex signature")
	}
}

func TestVerifyRequestResetsBody(t *testing.T) {
	body := []byte(`{"hello":"world"}`)
	sig := Generate(body, "secret")
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set(SignatureHeader, sig)

	got, err := VerifyRequest(req, "secret")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("VerifyRequest returned wrong body: %s", got)
	}
	// Body should still be readable by downstream code.
	again, _ := io.ReadAll(req.Body)
	if !bytes.Equal(again, body) {
		t.Errorf("r.Body was not reset; got %s", again)
	}
}

func TestVerifyRequestRejectsBadSignature(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{}`))
	req.Header.Set(SignatureHeader, "sha256=deadbeef")
	if _, err := VerifyRequest(req, "secret"); err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestMiddlewareRejects401(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })
	srv := httptest.NewServer(Middleware("secret", inner))
	defer srv.Close()

	resp, err := http.Post(srv.URL, "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	if called {
		t.Error("inner handler should not be called when signature is missing")
	}
}

func TestMiddlewareAcceptsValid(t *testing.T) {
	body := []byte(`{"event":"ok"}`)
	sig := Generate(body, "secret")

	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(Middleware("secret", inner))
	defer srv.Close()

	req, _ := http.NewRequest("POST", srv.URL, bytes.NewReader(body))
	req.Header.Set(SignatureHeader, sig)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
	if !called {
		t.Error("inner handler should have been called")
	}
}
