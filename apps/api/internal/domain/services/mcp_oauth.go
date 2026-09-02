package services

import (
	"context"
	"time"
)

type OAuthTokenResponse struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    time.Duration
}

// OAuthClient is the infrastructure boundary for provider token exchange. It
// must use the same outbound egress policy as MCP transports.
type OAuthClient interface {
	Exchange(ctx context.Context, tokenEndpoint, clientID, clientSecret, code, redirectURI, verifier string) (OAuthTokenResponse, error)
	Refresh(ctx context.Context, tokenEndpoint, clientID, clientSecret, refreshToken string) (OAuthTokenResponse, error)
}

type MCPCallOptions struct {
	CorrelationID  string
	IdempotencyKey string
	Idempotent     bool
}

type mcpCallOptionsContextKey struct{}

func WithMCPCallOptions(ctx context.Context, options MCPCallOptions) context.Context {
	return context.WithValue(ctx, mcpCallOptionsContextKey{}, options)
}

func MCPCallOptionsFromContext(ctx context.Context) MCPCallOptions {
	options, _ := ctx.Value(mcpCallOptionsContextKey{}).(MCPCallOptions)
	return options
}
