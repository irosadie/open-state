package services

import "context"

// OIDCIdentity is the normalized external identity returned by an OIDC provider
// (PRD §79). The (Provider, Subject) pair uniquely identifies a user.
type OIDCIdentity struct {
	Provider string
	Subject  string
	Email    string
	Name     string
	Photo    string
}

// OIDCProvider abstracts external OIDC authentication (PRD §79). It supports the
// Authorization Code flow with PKCE: generating an authorization URL, exchanging
// the code, and verifying the ID token.
type OIDCProvider interface {
	// AuthURL returns the provider authorization URL for the given state and
	// PKCE code verifier.
	AuthURL(state, codeVerifier string) string
	// Exchange exchanges an authorization code (with PKCE) for tokens and
	// returns the normalized identity.
	Exchange(ctx context.Context, code, codeVerifier string) (*OIDCIdentity, error)
	// VerifyIDToken validates an ID token and returns the normalized identity.
	VerifyIDToken(ctx context.Context, rawToken string) (*OIDCIdentity, error)
	// Name returns the provider identifier (e.g. "google").
	Name() string
}
