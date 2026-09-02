package dtos

// MCP connection DTOs intentionally omit credential values and protected
// credential references. credentialValue is write-only and is converted to an
// opaque secret-store reference before persistence.
type CreateMCPConnectionRequest struct {
	Name                       string   `json:"name"`
	Alias                      string   `json:"alias"`
	Transport                  string   `json:"transport"`
	Endpoint                   string   `json:"endpoint"`
	StdioProfile               string   `json:"stdioProfile"`
	StdioArgs                  []string `json:"stdioArgs"`
	AuthType                   string   `json:"authType"`
	CredentialReference        string   `json:"credentialReference"`
	CredentialValue            string   `json:"credentialValue"`
	OAuthAuthorizationEndpoint string   `json:"oauthAuthorizationEndpoint"`
	OAuthTokenEndpoint         string   `json:"oauthTokenEndpoint"`
	OAuthClientID              string   `json:"oauthClientId"`
	OAuthClientSecretValue     string   `json:"oauthClientSecretValue"`
	OAuthScopes                []string `json:"oauthScopes"`
	OAuthRedirectURI           string   `json:"oauthRedirectUri"`
	TimeoutMS                  int      `json:"timeoutMs"`
	MaxConcurrency             int      `json:"maxConcurrency"`
	RateLimitPerSecond         float64  `json:"rateLimitPerSecond"`
	RateLimitBurst             int      `json:"rateLimitBurst"`
	RetryMax                   int      `json:"retryMax"`
	CircuitFailureThreshold    int      `json:"circuitFailureThreshold"`
	CircuitRecoverySeconds     int      `json:"circuitRecoverySeconds"`
}

type UpdateMCPConnectionRequest = CreateMCPConnectionRequest

type MCPConnectionDTO struct {
	ID                         string   `json:"id"`
	TenantID                   string   `json:"tenantId"`
	ProjectID                  string   `json:"projectId"`
	Name                       string   `json:"name"`
	Alias                      string   `json:"alias"`
	Transport                  string   `json:"transport"`
	Endpoint                   *string  `json:"endpoint"`
	StdioProfile               *string  `json:"stdioProfile"`
	StdioArgs                  []string `json:"stdioArgs"`
	AuthType                   string   `json:"authType"`
	CredentialStatus           string   `json:"credentialStatus"`
	OAuthAuthorizationEndpoint *string  `json:"oauthAuthorizationEndpoint"`
	OAuthTokenEndpoint         *string  `json:"oauthTokenEndpoint"`
	OAuthClientID              *string  `json:"oauthClientId"`
	OAuthScopes                []string `json:"oauthScopes"`
	OAuthRedirectURI           *string  `json:"oauthRedirectUri"`
	OAuthStatus                string   `json:"oauthStatus"`
	Status                     string   `json:"status"`
	LastTestStatus             string   `json:"lastTestStatus"`
	LastTestErrorCode          *string  `json:"lastTestErrorCode"`
	LastTestedAt               *string  `json:"lastTestedAt"`
	HealthStatus               string   `json:"healthStatus"`
	HealthReason               *string  `json:"healthReason"`
	LastSuccessAt              *string  `json:"lastSuccessAt"`
	ConsecutiveFailures        int      `json:"consecutiveFailures"`
	CircuitOpenedAt            *string  `json:"circuitOpenedAt"`
	TimeoutMS                  int      `json:"timeoutMs"`
	MaxConcurrency             int      `json:"maxConcurrency"`
	RateLimitPerSecond         float64  `json:"rateLimitPerSecond"`
	RateLimitBurst             int      `json:"rateLimitBurst"`
	RetryMax                   int      `json:"retryMax"`
	CircuitFailureThreshold    int      `json:"circuitFailureThreshold"`
	CircuitRecoverySeconds     int      `json:"circuitRecoverySeconds"`
	CreatedAt                  string   `json:"createdAt"`
	UpdatedAt                  string   `json:"updatedAt"`
}

type MCPConnectionListDTO struct {
	Data []MCPConnectionDTO `json:"data"`
}

type MCPOAuthStartDTO struct {
	AuthorizationURL string `json:"authorizationUrl"`
	Status           string `json:"status"`
	ExpiresAt        string `json:"expiresAt"`
}

type MCPOAuthStatusDTO struct {
	Status           string  `json:"status"`
	ExpiresAt        *string `json:"expiresAt"`
	CredentialStatus string  `json:"credentialStatus"`
}

type MCPCredentialStatusDTO struct {
	Status           string `json:"status"`
	CredentialStatus string `json:"credentialStatus"`
	CanRotate        bool   `json:"canRotate"`
	CanRevoke        bool   `json:"canRevoke"`
}
