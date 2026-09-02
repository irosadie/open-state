package capability

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"

	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
)

// TrustedStdioProfile describes a server executable explicitly approved by the
// deployment. Project operators can select its name, but cannot submit a shell
// command or alter its environment through the registry API.
type TrustedStdioProfile struct {
	Command            string
	Args               []string
	AllowedArgPrefixes []string
	Env                []string
	MaxArgs            int
	MaxRuntime         time.Duration
	MaxOutputBytes     int64
}

// ParseTrustedStdioProfiles parses deployment-owned JSON profiles. The profile
// file is configuration, never user input; commands are still validated before
// being handed to the MCP SDK.
func ParseTrustedStdioProfiles(raw string) (map[string]TrustedStdioProfile, error) {
	profiles := make(map[string]TrustedStdioProfile)
	if strings.TrimSpace(raw) == "" {
		return profiles, nil
	}
	var input map[string]struct {
		Command            string   `json:"command"`
		Args               []string `json:"args"`
		AllowedArgPrefixes []string `json:"allowedArgPrefixes"`
		Env                []string `json:"env"`
		MaxArgs            int      `json:"maxArgs"`
		MaxRuntimeMS       int      `json:"maxRuntimeMs"`
		MaxOutputBytes     int64    `json:"maxOutputBytes"`
	}
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		return nil, errors.New("MCP_STDIO_PROFILES_JSON is invalid")
	}
	for name, value := range input {
		name = strings.TrimSpace(name)
		if name == "" || len(name) > 128 || strings.ContainsAny(name, "\x00\n\r") {
			return nil, errors.New("STDIO profile name is invalid")
		}
		if _, err := exec.LookPath(value.Command); err != nil && !strings.HasPrefix(value.Command, "/") {
			return nil, errors.New("STDIO profile executable is not resolvable")
		}
		maxArgs := value.MaxArgs
		if maxArgs <= 0 {
			maxArgs = 32
		}
		maxRuntime := time.Duration(value.MaxRuntimeMS) * time.Millisecond
		if maxRuntime <= 0 {
			maxRuntime = 30 * time.Second
		}
		maxOutputBytes := value.MaxOutputBytes
		if maxOutputBytes <= 0 {
			maxOutputBytes = 8 * 1024 * 1024
		}
		if maxArgs > 256 || maxRuntime > 10*time.Minute || maxOutputBytes > 64*1024*1024 {
			return nil, errors.New("STDIO profile resource limit is too large")
		}
		profiles[name] = TrustedStdioProfile{Command: value.Command, Args: append([]string{}, value.Args...), AllowedArgPrefixes: append([]string{}, value.AllowedArgPrefixes...), Env: append([]string{}, value.Env...), MaxArgs: maxArgs, MaxRuntime: maxRuntime, MaxOutputBytes: maxOutputBytes}
	}
	return profiles, nil
}

// MCPConnectionTester performs only MCP initialize/handshake. It deliberately
// never calls tools/list or any provider business tool.
type MCPConnectionTester struct {
	profiles    map[string]TrustedStdioProfile
	resolver    CredentialResolver
	secretStore domainsvc.SecretStore
	oauthTokens domainsvc.OAuthAccessTokenProvider
	egress      *EgressPolicy
	timeout     time.Duration
}

type MCPConnectionTesterOption func(*MCPConnectionTester)

func WithSecretStore(store domainsvc.SecretStore) MCPConnectionTesterOption {
	return func(t *MCPConnectionTester) { t.secretStore = store }
}

func WithOAuthAccessTokenProvider(provider domainsvc.OAuthAccessTokenProvider) MCPConnectionTesterOption {
	return func(t *MCPConnectionTester) { t.oauthTokens = provider }
}

func WithEgressPolicy(policy *EgressPolicy) MCPConnectionTesterOption {
	return func(t *MCPConnectionTester) { t.egress = policy }
}

func NewMCPConnectionTester(profiles map[string]TrustedStdioProfile, resolver CredentialResolver, timeout time.Duration, options ...MCPConnectionTesterOption) *MCPConnectionTester {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	if profiles == nil {
		profiles = map[string]TrustedStdioProfile{}
	}
	tester := &MCPConnectionTester{profiles: profiles, resolver: resolver, timeout: timeout}
	for _, option := range options {
		if option != nil {
			option(tester)
		}
	}
	return tester
}

// SetOAuthAccessTokenProvider wires the application OAuth lifecycle after the
// transport adapter has been constructed in the composition root.
func (t *MCPConnectionTester) SetOAuthAccessTokenProvider(provider domainsvc.OAuthAccessTokenProvider) {
	t.oauthTokens = provider
}

func (t *MCPConnectionTester) Handshake(ctx context.Context, connection *entities.MCPConnection) (domainsvc.MCPHandshakeResult, error) {
	if connection == nil {
		return domainsvc.MCPHandshakeResult{ErrorCode: "mcp_connection_missing"}, errors.New("MCP connection is missing")
	}
	headers, err := t.authHeaders(ctx, connection)
	if err != nil {
		return domainsvc.MCPHandshakeResult{ErrorCode: authErrorCode(err)}, err
	}
	callCtx, cancel := context.WithTimeout(ctx, t.operationTimeout(connection))
	defer cancel()

	c, err, errorCode := t.newClient(callCtx, connection, headers)
	if err != nil {
		return domainsvc.MCPHandshakeResult{ErrorCode: errorCode}, safeHandshakeError(err)
	}
	defer c.Close()
	if _, err := c.Initialize(callCtx, mcp.InitializeRequest{}); err != nil {
		return domainsvc.MCPHandshakeResult{ErrorCode: classifyHandshakeError(err)}, safeHandshakeError(err)
	}
	return domainsvc.MCPHandshakeResult{Ready: true}, nil
}

// DiscoverTools initializes a new MCP session and requests tools/list. This is
// the only adapter path that reads provider tools; it intentionally has no
// tools/call invocation path.
func (t *MCPConnectionTester) DiscoverTools(ctx context.Context, connection *entities.MCPConnection) (domainsvc.MCPToolDiscoveryResult, error) {
	if connection == nil {
		return domainsvc.MCPToolDiscoveryResult{ErrorCode: "mcp_connection_missing"}, errors.New("MCP connection is missing")
	}
	headers, err := t.authHeaders(ctx, connection)
	if err != nil {
		return domainsvc.MCPToolDiscoveryResult{ErrorCode: authErrorCode(err)}, safeDiscoveryError(err)
	}
	callCtx, cancel := context.WithTimeout(ctx, t.operationTimeout(connection))
	defer cancel()
	c, err, errorCode := t.newClient(callCtx, connection, headers)
	if err != nil {
		return domainsvc.MCPToolDiscoveryResult{ErrorCode: errorCode}, safeDiscoveryError(err)
	}
	defer c.Close()
	if _, err := c.Initialize(callCtx, mcp.InitializeRequest{}); err != nil {
		return domainsvc.MCPToolDiscoveryResult{ErrorCode: classifyDiscoveryError(err)}, safeDiscoveryError(err)
	}
	result, err := c.ListTools(callCtx, mcp.ListToolsRequest{})
	if err != nil {
		return domainsvc.MCPToolDiscoveryResult{ErrorCode: classifyDiscoveryError(err)}, safeDiscoveryError(err)
	}
	tools := make([]domainsvc.MCPToolDefinition, 0, len(result.Tools))
	for _, tool := range result.Tools {
		inputSchema, err := toolInputSchema(tool)
		if err != nil {
			return domainsvc.MCPToolDiscoveryResult{ErrorCode: "mcp_discovery_invalid_response"}, safeDiscoveryError(err)
		}
		annotations, err := json.Marshal(tool.Annotations)
		if err != nil {
			return domainsvc.MCPToolDiscoveryResult{ErrorCode: "mcp_discovery_invalid_response"}, safeDiscoveryError(err)
		}
		tools = append(tools, domainsvc.MCPToolDefinition{
			Name: tool.Name, Title: tool.Title, Description: tool.Description,
			InputSchema: inputSchema, Annotations: annotations,
		})
	}
	return domainsvc.MCPToolDiscoveryResult{Tools: tools}, nil
}

// InvokeTool executes only the tool selected by an already verified project
// binding. It deliberately does not call tools/list and does not accept a raw
// endpoint, alias, credential, or caller-provided tool name.
func (t *MCPConnectionTester) InvokeTool(ctx context.Context, connection *entities.MCPConnection, tool *entities.MCPDiscoveredTool, payload map[string]any, timeout time.Duration) (domainsvc.MCPToolCallResult, error) {
	if connection == nil {
		return domainsvc.MCPToolCallResult{}, domaincap.NewCapabilityError(domaincap.ErrorKindUnavailable, "capability.mcp_connection_missing", "MCP connection is unavailable")
	}
	if tool == nil || strings.TrimSpace(tool.Name) == "" {
		return domainsvc.MCPToolCallResult{}, domaincap.NewCapabilityError(domaincap.ErrorKindUnavailable, "capability.mcp_tool_missing", "MCP tool binding is unavailable")
	}
	if connection.Status != entities.MCPConnectionEnabled {
		return domainsvc.MCPToolCallResult{}, domaincap.NewCapabilityError(domaincap.ErrorKindUnavailable, "capability.mcp_connection_disabled", "MCP connection is disabled")
	}
	if tool.Availability != entities.MCPToolPresent || !tool.Enabled {
		return domainsvc.MCPToolCallResult{}, domaincap.NewCapabilityError(domaincap.ErrorKindUnavailable, "capability.mcp_tool_unavailable", "MCP tool is unavailable")
	}
	headers, err := t.authHeaders(ctx, connection)
	if err != nil {
		if authErrorCode(err) == "oauth_authorization_required" {
			return domainsvc.MCPToolCallResult{}, domaincap.NewCapabilityError(domaincap.ErrorKindUnauthorized, "capability.oauth_authorization_required", "MCP authorization is required")
		}
		return domainsvc.MCPToolCallResult{}, domaincap.NewCapabilityError(domaincap.ErrorKindUnauthorized, "capability.credential_unavailable", "MCP credentials are unavailable")
	}
	if timeout <= 0 {
		timeout = t.timeout
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	started := time.Now()
	c, err, errorCode := t.newClient(callCtx, connection, headers)
	if err != nil {
		return domainsvc.MCPToolCallResult{}, mapResolvedMCPError(errorCode, err)
	}
	defer c.Close()
	if _, err := c.Initialize(callCtx, mcp.InitializeRequest{}); err != nil {
		return domainsvc.MCPToolCallResult{}, mapResolvedMCPError(classifyHandshakeError(err), err)
	}
	result, err := c.CallTool(callCtx, mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: tool.Name, Arguments: payload,
	}})
	if err != nil {
		return domainsvc.MCPToolCallResult{}, mapResolvedMCPError(classifyInvocationError(err), err)
	}
	if result == nil {
		return domainsvc.MCPToolCallResult{}, domaincap.NewCapabilityError(domaincap.ErrorKindExternal, "capability.invalid_provider_response", "MCP provider returned an invalid response")
	}
	if result.IsError {
		return domainsvc.MCPToolCallResult{}, domaincap.NewCapabilityError(domaincap.ErrorKindBusiness, "capability.business_error", "MCP provider rejected the capability request")
	}
	data, err := normalizeResolvedMCPResult(result)
	if err != nil {
		return domainsvc.MCPToolCallResult{}, domaincap.NewCapabilityError(domaincap.ErrorKindExternal, "capability.invalid_provider_response", "MCP provider returned an invalid response")
	}
	return domainsvc.MCPToolCallResult{Data: data, Duration: time.Since(started)}, nil
}

func (t *MCPConnectionTester) operationTimeout(connection *entities.MCPConnection) time.Duration {
	timeout := t.timeout
	if connection != nil && connection.StdioProfile != nil {
		if profile, ok := t.profiles[*connection.StdioProfile]; ok && profile.MaxRuntime > 0 && profile.MaxRuntime < timeout {
			return profile.MaxRuntime
		}
	}
	return timeout
}

func (t *MCPConnectionTester) newClient(ctx context.Context, connection *entities.MCPConnection, headers map[string]string) (*client.Client, error, string) {
	var (
		c   *client.Client
		err error
	)
	switch connection.Transport {
	case entities.MCPTransportStreamableHTTP:
		if connection.Endpoint == nil {
			return nil, errors.New("endpoint is missing"), "endpoint_missing"
		}
		if t.egress != nil {
			if err := t.egress.ValidateURL(ctx, *connection.Endpoint); err != nil {
				return nil, err, "mcp_egress_blocked"
			}
			c, err = client.NewStreamableHttpClient(*connection.Endpoint, transport.WithHTTPHeaders(headers), transport.WithHTTPBasicClient(t.egress.HTTPClient(ctx)))
		} else {
			c, err = client.NewStreamableHttpClient(*connection.Endpoint, transport.WithHTTPHeaders(headers))
		}
	case entities.MCPTransportSSE:
		if connection.Endpoint == nil {
			return nil, errors.New("endpoint is missing"), "endpoint_missing"
		}
		if t.egress != nil {
			if err := t.egress.ValidateURL(ctx, *connection.Endpoint); err != nil {
				return nil, err, "mcp_egress_blocked"
			}
			c, err = client.NewSSEMCPClient(*connection.Endpoint, client.WithHeaders(headers), client.WithHTTPClient(t.egress.HTTPClient(ctx)))
		} else {
			c, err = client.NewSSEMCPClient(*connection.Endpoint, client.WithHeaders(headers))
		}
	case entities.MCPTransportSTDIO:
		profile, ok := t.profiles[valueOrEmpty(connection.StdioProfile)]
		if !ok || profile.Command == "" {
			return nil, errors.New("stdio profile is not trusted by this deployment"), "stdio_profile_not_trusted"
		}
		if err := validateStdioProfile(profile, connection.StdioArgs); err != nil {
			return nil, err, "stdio_profile_args_not_allowed"
		}
		c, err = client.NewStdioMCPClient(profile.Command, profile.Env, append(append([]string{}, profile.Args...), connection.StdioArgs...)...)
	default:
		return nil, errors.New("unsupported MCP transport"), "transport_unsupported"
	}
	if err != nil {
		return nil, err, "mcp_client_unavailable"
	}
	return c, nil, ""
}

func toolInputSchema(tool mcp.Tool) (json.RawMessage, error) {
	if len(tool.RawInputSchema) > 0 {
		if !json.Valid(tool.RawInputSchema) {
			return nil, errors.New("tool input schema is invalid JSON")
		}
		return append(json.RawMessage(nil), tool.RawInputSchema...), nil
	}
	value, err := json.Marshal(tool.InputSchema)
	if err != nil || !json.Valid(value) {
		return nil, errors.New("tool input schema is invalid JSON")
	}
	return value, nil
}

func (t *MCPConnectionTester) authHeaders(ctx context.Context, connection *entities.MCPConnection) (map[string]string, error) {
	if connection.AuthType != entities.MCPAuthBearer && connection.AuthType != entities.MCPAuthOAuth {
		return map[string]string{}, nil
	}
	if connection.AuthType == entities.MCPAuthOAuth && t.oauthTokens != nil {
		secret, err := t.oauthTokens.AccessToken(ctx, connection.ID, connection.TenantID, connection.ProjectID)
		if err != nil {
			return nil, err
		}
		return map[string]string{"Authorization": "Bearer " + secret}, nil
	}
	if connection.AuthType == entities.MCPAuthOAuth && (connection.CredentialReference == nil || *connection.CredentialReference == "") {
		return nil, errors.New("OAuth authorization is not completed")
	}
	if connection.CredentialReference == nil || *connection.CredentialReference == "" {
		return nil, errors.New("MCP credential is not configured")
	}
	if t.secretStore != nil {
		secret, err := t.secretStore.Resolve(ctx, connection.TenantID, connection.ProjectID, *connection.CredentialReference)
		if err != nil {
			return nil, errors.New("MCP credential is unavailable")
		}
		return map[string]string{"Authorization": "Bearer " + secret}, nil
	}
	if t.resolver == nil {
		return nil, errors.New("credential resolver is not configured")
	}
	secret, ok := t.resolver.Resolve(*connection.CredentialReference)
	if !ok {
		return nil, errors.New("bearer credential is unavailable")
	}
	return map[string]string{"Authorization": "Bearer " + secret}, nil
}

func validateStdioProfile(profile TrustedStdioProfile, extra []string) error {
	if strings.TrimSpace(profile.Command) == "" || strings.ContainsAny(profile.Command, ";&|$`\n\r") {
		return errors.New("stdio executable is not allowed")
	}
	if profile.MaxArgs > 0 && len(profile.Args)+len(extra) > profile.MaxArgs {
		return errors.New("stdio argument limit exceeded")
	}
	for _, arg := range extra {
		if strings.TrimSpace(arg) == "" || strings.ContainsAny(arg, "\x00\n\r") {
			return errors.New("stdio argument is not allowed")
		}
		if len(profile.AllowedArgPrefixes) == 0 {
			return errors.New("stdio profile does not allow connection arguments")
		}
		allowed := false
		for _, prefix := range profile.AllowedArgPrefixes {
			if strings.HasPrefix(arg, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return errors.New("stdio argument is outside the reviewed profile")
		}
	}
	return nil
}

func authErrorCode(err error) string {
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "oauth") {
		return "oauth_authorization_required"
	}
	return "credential_unavailable"
}

func classifyHandshakeError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "mcp_handshake_timeout"
	}
	return "mcp_handshake_failed"
}

func safeHandshakeError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("MCP handshake failed")
}

func classifyDiscoveryError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "mcp_discovery_timeout"
	}
	return "mcp_discovery_failed"
}

func safeDiscoveryError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("MCP tool discovery failed")
}

func classifyInvocationError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "mcp_invocation_timeout"
	}
	return "mcp_invocation_failed"
}

func mapResolvedMCPError(code string, _ error) error {
	if code == "mcp_invocation_timeout" || code == "mcp_handshake_timeout" {
		return domaincap.NewCapabilityError(domaincap.ErrorKindTimeout, "capability.timeout", "MCP capability invocation timed out")
	}
	if code == "credential_unavailable" || code == "oauth_authorization_required" {
		return domaincap.NewCapabilityError(domaincap.ErrorKindUnauthorized, "capability.credential_unavailable", "MCP credentials are unavailable")
	}
	return domaincap.NewCapabilityError(domaincap.ErrorKindUnavailable, "capability.unavailable", "MCP capability provider is unavailable")
}

func normalizeResolvedMCPResult(result *mcp.CallToolResult) (map[string]any, error) {
	if result.StructuredContent != nil {
		if structured, ok := result.StructuredContent.(map[string]any); ok {
			return structured, nil
		}
		if len(result.RawStructuredContent) > 0 {
			var structured map[string]any
			if err := json.Unmarshal(result.RawStructuredContent, &structured); err == nil {
				return structured, nil
			}
		}
	}
	for _, content := range result.Content {
		textContent, ok := content.(mcp.TextContent)
		if !ok {
			continue
		}
		var data map[string]any
		if err := json.Unmarshal([]byte(textContent.Text), &data); err == nil && data != nil {
			return data, nil
		}
		return map[string]any{"text": textContent.Text}, nil
	}
	return map[string]any{}, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
