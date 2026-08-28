package oidc

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestPKCEChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	got := pkceChallenge(verifier)

	// Reference: RFC 7636 Appendix B.
	want := base64.RawURLEncoding.EncodeToString(func() []byte {
		s := sha256.Sum256([]byte(verifier))
		return s[:]
	}())
	if got != want {
		t.Errorf("pkceChallenge = %q, want %q", got, want)
	}
}

func TestPKCEChallengeIsDeterministicAndScoped(t *testing.T) {
	if pkceChallenge("verifier-a") != pkceChallenge("verifier-a") {
		t.Error("expected deterministic challenge for same verifier")
	}
	if pkceChallenge("verifier-a") == pkceChallenge("verifier-b") {
		t.Error("expected different challenges for different verifiers")
	}
}
