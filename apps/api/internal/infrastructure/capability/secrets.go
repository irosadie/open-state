package capability

import (
	"os"
	"strings"
)

// CredentialResolver resolves a credential_reference to the actual credential
// from secure infrastructure (env / secret manager / Vault) (PRD §61).
// It never stores secrets in workflow definitions and never returns them for
// logging (PRD §91).
type CredentialResolver interface {
	Resolve(ref string) (string, bool)
}

// EnvCredentialResolver resolves credential references from environment
// variables using an optional prefix (e.g. "CRED_").
type EnvCredentialResolver struct {
	Prefix string
}

// Resolve implements CredentialResolver. It returns the env value for the
// variable named by prefix+ref (uppercased, non-alphanumeric → underscore).
func (r EnvCredentialResolver) Resolve(ref string) (string, bool) {
	if ref == "" {
		return "", false
	}
	key := r.Prefix + sanitizeEnvKey(ref)
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return "", false
	}
	return v, true
}

// sanitizeEnvKey converts a credential reference to a safe env var name.
func sanitizeEnvKey(ref string) string {
	var b strings.Builder
	for _, r := range ref {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return strings.ToUpper(b.String())
}
