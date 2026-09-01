package capability

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	Command string
	Args    []string
	Env     []string
}

// MCPConnectionTester performs only MCP initialize/handshake. It deliberately
// never calls tools/list or any provider business tool.
type MCPConnectionTester struct {
	profiles map[string]TrustedStdioProfile
	resolver CredentialResolver
	timeout  time.Duration
}

func NewMCPConnectionTester(profiles map[string]TrustedStdioProfile, resolver CredentialResolver, timeout time.Duration) *MCPConnectionTester {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	if profiles == nil {
		profiles = map[string]TrustedStdioProfile{}
	}
	return &MCPConnectionTester{profiles: profiles, resolver: resolver, timeout: timeout}
}

func (t *MCPConnectionTester) Handshake(ctx context.Context, connection *entities.MCPConnection) (domainsvc.MCPHandshakeResult, error) {
	if connection == nil {
		return domainsvc.MCPHandshakeResult{ErrorCode: "mcp_connection_missing"}, errors.New("MCP connection is missing")
	}
	if connection.AuthType == entities.MCPAuthOAuth {
		return domainsvc.MCPHandshakeResult{ErrorCode: "oauth_authorization_required"}, errors.New("OAuth authorization is not completed")
	}
	headers, err := t.authHeaders(connection)
	if err != nil {
		return domainsvc.MCPHandshakeResult{ErrorCode: "credential_unavailable"}, err
	}
	callCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	c, err, errorCode := t.newClient(connection, headers)
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
	if connection.AuthType == entities.MCPAuthOAuth {
		return domainsvc.MCPToolDiscoveryResult{ErrorCode: "oauth_authorization_required"}, errors.New("OAuth authorization is not completed")
	}
	headers, err := t.authHeaders(connection)
	if err != nil {
		return domainsvc.MCPToolDiscoveryResult{ErrorCode: "credential_unavailable"}, safeDiscoveryError(err)
	}
	callCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	c, err, errorCode := t.newClient(connection, headers)
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
	if connection.AuthType == entities.MCPAuthOAuth {
		return domainsvc.MCPToolCallResult{}, domaincap.NewCapabilityError(domaincap.ErrorKindUnavailable, "capability.oauth_authorization_required", "MCP authorization is required")
	}
	headers, err := t.authHeaders(connection)
	if err != nil {
		return domainsvc.MCPToolCallResult{}, domaincap.NewCapabilityError(domaincap.ErrorKindUnauthorized, "capability.credential_unavailable", "MCP credentials are unavailable")
	}
	if timeout <= 0 {
		timeout = t.timeout
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	started := time.Now()
	c, err, errorCode := t.newClient(connection, headers)
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

func (t *MCPConnectionTester) newClient(connection *entities.MCPConnection, headers map[string]string) (*client.Client, error, string) {
	var (
		c   *client.Client
		err error
	)
	switch connection.Transport {
	case entities.MCPTransportStreamableHTTP:
		if connection.Endpoint == nil {
			return nil, errors.New("endpoint is missing"), "endpoint_missing"
		}
		c, err = client.NewStreamableHttpClient(*connection.Endpoint, transport.WithHTTPHeaders(headers))
	case entities.MCPTransportSSE:
		if connection.Endpoint == nil {
			return nil, errors.New("endpoint is missing"), "endpoint_missing"
		}
		c, err = client.NewSSEMCPClient(*connection.Endpoint, client.WithHeaders(headers))
	case entities.MCPTransportSTDIO:
		profile, ok := t.profiles[valueOrEmpty(connection.StdioProfile)]
		if !ok || profile.Command == "" {
			return nil, errors.New("stdio profile is not trusted by this deployment"), "stdio_profile_not_trusted"
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

func (t *MCPConnectionTester) authHeaders(connection *entities.MCPConnection) (map[string]string, error) {
	if connection.AuthType != entities.MCPAuthBearer {
		return map[string]string{}, nil
	}
	if connection.CredentialReference == nil || *connection.CredentialReference == "" {
		return nil, errors.New("bearer credential is not configured")
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
