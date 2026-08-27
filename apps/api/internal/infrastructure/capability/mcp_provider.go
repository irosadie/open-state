package capability

import (
	"context"
	"errors"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
)

// MCPProviderConfig configures the MCP client adapter.
type MCPProviderConfig struct {
	// Endpoint is the MCP server base URL (Streamable HTTP).
	Endpoint string
	// ToolName is the MCP tool to call for the capability.
	ToolName string
	// Timeout is the per-call timeout.
	Timeout time.Duration
}

// MCPProvider implements domaincap.CapabilityProvider by calling a remote MCP
// tool over Streamable HTTP. It is the primary production integration
// (PRD §2201). The core engine remains MCP-agnostic (PRD §172, §2559).
type MCPProvider struct {
	cfg     MCPProviderConfig
	headers map[string]string
}

// NewMCPProvider builds an adapter for a remote MCP endpoint.
func NewMCPProvider(cfg MCPProviderConfig, headers map[string]string) *MCPProvider {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.ToolName == "" {
		cfg.ToolName = "invoke_capability"
	}
	return &MCPProvider{cfg: cfg, headers: headers}
}

// Invoke implements domaincap.CapabilityProvider.
func (p *MCPProvider) Invoke(ctx context.Context, inv domaincap.Invocation) (domaincap.InvocationResult, error) {
	c, err := client.NewStreamableHttpClient(p.cfg.Endpoint)
	if err != nil {
		return domaincap.InvocationResult{}, domaincap.NewCapabilityError(
			domaincap.ErrorKindUnavailable, "capability.unavailable", "failed to create MCP client: "+err.Error())
	}

	callCtx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	if _, err := c.Initialize(callCtx, mcp.InitializeRequest{}); err != nil {
		return domaincap.InvocationResult{}, mapMCPError(err)
	}

	req := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name:      p.cfg.ToolName,
		Arguments: inv.Payload,
	}}
	res, err := c.CallTool(callCtx, req)
	if err != nil {
		return domaincap.InvocationResult{}, mapMCPError(err)
	}
	if res.IsError {
		return domaincap.InvocationResult{}, domaincap.NewCapabilityError(
			domaincap.ErrorKindBusiness, "capability.business_error", "MCP tool returned error")
	}

	data := extractResultData(res)
	event := "capability.success"
	return domaincap.InvocationResult{
		Data:            data,
		FromMock:        false,
		Duration:        0,
		CapabilityEvent: &event,
	}, nil
}

// mapMCPError maps transport/timeout errors to classified CapabilityError.
func mapMCPError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return domaincap.NewCapabilityError(domaincap.ErrorKindTimeout, "capability.timeout", "MCP call timed out")
	}
	return domaincap.NewCapabilityError(domaincap.ErrorKindExternal, "capability.failed", "MCP call failed: "+err.Error())
}

// extractResultData converts the MCP call result content to a data map.
func extractResultData(res *mcp.CallToolResult) map[string]any {
	data := map[string]any{}
	if res == nil {
		return data
	}
	for _, c := range res.Content {
		if txt, ok := c.(mcp.TextContent); ok {
			data["text"] = txt.Text
		}
	}
	return data
}
