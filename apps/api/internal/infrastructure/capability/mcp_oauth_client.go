package capability

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
)

type HTTPMCPOAuthClient struct {
	client *http.Client
	policy *EgressPolicy
}

func NewHTTPMCPOAuthClient(policy *EgressPolicy) *HTTPMCPOAuthClient {
	if policy == nil {
		return &HTTPMCPOAuthClient{client: &http.Client{Timeout: 15 * time.Second}}
	}
	return &HTTPMCPOAuthClient{client: policy.HTTPClient(context.Background()), policy: policy}
}

func (c *HTTPMCPOAuthClient) Exchange(ctx context.Context, endpoint, clientID, clientSecret, code, redirectURI, verifier string) (domainsvc.OAuthTokenResponse, error) {
	return c.post(ctx, endpoint, url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	})
}

func (c *HTTPMCPOAuthClient) Refresh(ctx context.Context, endpoint, clientID, clientSecret, refreshToken string) (domainsvc.OAuthTokenResponse, error) {
	return c.post(ctx, endpoint, url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {refreshToken},
	})
}

func (c *HTTPMCPOAuthClient) post(ctx context.Context, endpoint string, values url.Values) (domainsvc.OAuthTokenResponse, error) {
	if strings.TrimSpace(endpoint) == "" {
		return domainsvc.OAuthTokenResponse{}, errors.New("OAuth token endpoint is missing")
	}
	if c.policy != nil {
		if err := c.policy.ValidateURL(ctx, endpoint); err != nil {
			return domainsvc.OAuthTokenResponse{}, errors.New("OAuth token endpoint is blocked by egress policy")
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return domainsvc.OAuthTokenResponse{}, errors.New("OAuth token request is invalid")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := c.client.Do(req)
	if err != nil {
		return domainsvc.OAuthTokenResponse{}, errors.New("OAuth token exchange failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return domainsvc.OAuthTokenResponse{}, errors.New("OAuth provider rejected the token request")
	}
	var payload struct {
		AccessToken  string      `json:"access_token"`
		RefreshToken string      `json:"refresh_token"`
		TokenType    string      `json:"token_type"`
		ExpiresIn    json.Number `json:"expires_in"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 64*1024))
	if err := decoder.Decode(&payload); err != nil || strings.TrimSpace(payload.AccessToken) == "" {
		return domainsvc.OAuthTokenResponse{}, errors.New("OAuth provider returned an invalid token response")
	}
	expiresIn := 0
	if payload.ExpiresIn != "" {
		expiresIn, _ = strconv.Atoi(payload.ExpiresIn.String())
	}
	if expiresIn <= 0 || expiresIn > 365*24*60*60 {
		expiresIn = 3600
	}
	return domainsvc.OAuthTokenResponse{AccessToken: payload.AccessToken, RefreshToken: payload.RefreshToken, TokenType: payload.TokenType, ExpiresIn: time.Duration(expiresIn) * time.Second}, nil
}
