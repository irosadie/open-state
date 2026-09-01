package capability

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// ValidateMCPProviderContract checks a host-configured provider alias against
// the provider MCP identity and tools/list response. The State MCP does not
// call this function: the LLM host owns the connection and endpoint mapping.
func ValidateMCPProviderContract(ctx context.Context, endpoint, expectedServer, expectedTool string) error {
	if strings.TrimSpace(endpoint) == "" || strings.TrimSpace(expectedServer) == "" || strings.TrimSpace(expectedTool) == "" {
		return errors.New("provider endpoint, server alias, and tool are required")
	}
	mcpClient, err := client.NewStreamableHttpClient(endpoint)
	if err != nil {
		return fmt.Errorf("create provider MCP client: %w", err)
	}
	defer mcpClient.Close()
	initialized, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{})
	if err != nil {
		return fmt.Errorf("initialize provider MCP server: %w", err)
	}
	if initialized.ServerInfo.Name != expectedServer {
		return fmt.Errorf("provider server identity mismatch: expected %q, got %q", expectedServer, initialized.ServerInfo.Name)
	}
	tools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("discover provider MCP tools: %w", err)
	}
	for _, tool := range tools.Tools {
		if tool.Name == expectedTool {
			return nil
		}
	}
	return fmt.Errorf("provider tool %q is not exposed by server %q", expectedTool, expectedServer)
}
