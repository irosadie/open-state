package config

import "os"

// SSOConfig holds the OIDC providers configuration (PRD §79). A provider is
// enabled only when its client id is set.
type SSOConfig struct {
	Google  OIDCProviderConfig
	GitHub  OIDCProviderConfig
	Entra   OIDCProviderConfig
	BaseURL string
}

// OIDCProviderConfig is the per-provider OIDC settings.
type OIDCProviderConfig struct {
	ClientID     string
	ClientSecret string
	IssuerURL    string
	RedirectURI  string
	Scopes       []string
}

// Enabled reports whether the provider has a client id configured.
func (p OIDCProviderConfig) Enabled() bool { return p.ClientID != "" }

// loadSSO reads SSO/OIDC env vars.
func loadSSO() SSOConfig {
	return SSOConfig{
		BaseURL: os.Getenv("SSO_BASE_URL"),
		Google: OIDCProviderConfig{
			ClientID:     os.Getenv("SSO_GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("SSO_GOOGLE_CLIENT_SECRET"),
			IssuerURL:    envDefault("SSO_GOOGLE_ISSUER", "https://accounts.google.com"),
			RedirectURI:  os.Getenv("SSO_GOOGLE_REDIRECT_URI"),
			Scopes:       []string{"openid", "email", "profile"},
		},
		GitHub: OIDCProviderConfig{
			ClientID:     os.Getenv("SSO_GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("SSO_GITHUB_CLIENT_SECRET"),
			IssuerURL:    envDefault("SSO_GITHUB_ISSUER", "https://token.actions.githubusercontent.com"),
			RedirectURI:  os.Getenv("SSO_GITHUB_REDIRECT_URI"),
			Scopes:       []string{"openid", "email", "profile"},
		},
		Entra: OIDCProviderConfig{
			ClientID:     os.Getenv("SSO_ENTRA_CLIENT_ID"),
			ClientSecret: os.Getenv("SSO_ENTRA_CLIENT_SECRET"),
			IssuerURL:    os.Getenv("SSO_ENTRA_ISSUER"),
			RedirectURI:  os.Getenv("SSO_ENTRA_REDIRECT_URI"),
			Scopes:       []string{"openid", "email", "profile"},
		},
	}
}
