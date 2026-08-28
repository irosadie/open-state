package oidc

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/irosadie/open-state/api/internal/domain/services"
)

// ProviderConfig configures a single OIDC provider adapter.
type ProviderConfig struct {
	// Provider identifier (e.g. "google").
	Name string
	// ClientID and ClientSecret from the provider app.
	ClientID     string
	ClientSecret string
	// IssuerURL for OIDC discovery (well-known config + JWKS).
	IssuerURL string
	// RedirectURI registered with the provider.
	RedirectURI string
	// Scopes requested (openid, email, profile).
	Scopes []string
}

// Provider is an infrastructure OIDC adapter implementing the domain
// OIDCProvider port (PRD §79). It performs discovery (JWKS), the authorization
// code flow with PKCE, and ID token verification.
type Provider struct {
	name        string
	oauthConfig *oauth2.Config
	verifier    *oidc.IDTokenVerifier
	provider    *oidc.Provider
}

// New builds a Provider by performing OIDC discovery for the issuer.
func New(ctx context.Context, cfg ProviderConfig) (*Provider, error) {
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery for %s: %w", cfg.Name, err)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Endpoint:     provider.Endpoint(),
		Scopes:       cfg.Scopes,
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})

	return &Provider{
		name:        cfg.Name,
		oauthConfig: oauthConfig,
		verifier:    verifier,
		provider:    provider,
	}, nil
}

// Name returns the provider identifier.
func (p *Provider) Name() string { return p.name }

// AuthURL returns the authorization URL with PKCE (S256) and the given state.
// The code_challenge is the base64url(SHA-256) of the code_verifier (RFC 7636).
func (p *Provider) AuthURL(state, codeVerifier string) string {
	return p.oauthConfig.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge(codeVerifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", state))
}

// pkceChallenge computes the S256 code challenge from a code verifier.
func pkceChallenge(codeVerifier string) string {
	sum := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// Exchange exchanges the authorization code (with PKCE) and returns the identity
// verified from the ID token (PRD §79).
func (p *Provider) Exchange(ctx context.Context, code, codeVerifier string) (*services.OIDCIdentity, error) {
	token, err := p.oauthConfig.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return nil, fmt.Errorf("oidc code exchange: %w", err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, errors.New("oidc exchange: missing id_token")
	}
	return p.verify(ctx, rawIDToken)
}

// VerifyIDToken validates a raw ID token and returns the identity.
func (p *Provider) VerifyIDToken(ctx context.Context, rawToken string) (*services.OIDCIdentity, error) {
	return p.verify(ctx, rawToken)
}

func (p *Provider) verify(ctx context.Context, rawToken string) (*services.OIDCIdentity, error) {
	idToken, err := p.verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("oidc id_token verification failed: %w", err)
	}

	var claims struct {
		Subject string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("oidc claims parse: %w", err)
	}

	return &services.OIDCIdentity{
		Provider: p.name,
		Subject:  claims.Subject,
		Email:    claims.Email,
		Name:     claims.Name,
		Photo:    claims.Picture,
	}, nil
}
