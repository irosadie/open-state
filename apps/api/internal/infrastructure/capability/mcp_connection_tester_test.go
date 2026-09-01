package capability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestMCPConnectionTesterRequiresOAuthAuthorization(t *testing.T) {
	tester := NewMCPConnectionTester(nil, nil, 0)
	result, err := tester.Handshake(context.Background(), &entities.MCPConnection{AuthType: entities.MCPAuthOAuth})
	if err == nil || result.Ready || result.ErrorCode != "oauth_authorization_required" {
		t.Fatalf("result = %#v, err = %v", result, err)
	}
}

func TestMCPConnectionTesterRejectsUntrustedSTDIOProfile(t *testing.T) {
	tester := NewMCPConnectionTester(nil, nil, 0)
	profile := "unapproved"
	result, err := tester.Handshake(context.Background(), &entities.MCPConnection{Transport: entities.MCPTransportSTDIO, StdioProfile: &profile, AuthType: entities.MCPAuthNone})
	if err == nil || result.Ready || result.ErrorCode != "stdio_profile_not_trusted" {
		t.Fatalf("result = %#v, err = %v", result, err)
	}
}

func TestMCPConnectionTesterInitializesStreamableHTTPWithoutDiscovery(t *testing.T) {
	mcpServer := server.NewMCPServer("provider", "1.0.0")
	testServer := server.NewTestStreamableHTTPServer(mcpServer)
	defer testServer.Close()

	tester := NewMCPConnectionTester(nil, nil, 0)
	result, err := tester.Handshake(context.Background(), &entities.MCPConnection{Transport: entities.MCPTransportStreamableHTTP, Endpoint: stringPtr(testServer.URL), AuthType: entities.MCPAuthNone})
	if err != nil || !result.Ready {
		t.Fatalf("result = %#v, err = %v", result, err)
	}
}

func TestMCPConnectionTesterClassifiesFailedHTTPHandshake(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer testServer.Close()

	tester := NewMCPConnectionTester(nil, nil, 0)
	result, err := tester.Handshake(context.Background(), &entities.MCPConnection{Transport: entities.MCPTransportStreamableHTTP, Endpoint: stringPtr(testServer.URL), AuthType: entities.MCPAuthNone})
	if err == nil || result.Ready || result.ErrorCode == "" {
		t.Fatalf("result = %#v, err = %v", result, err)
	}
}

func TestMCPConnectionTesterDiscoversToolsWithoutCallingBusinessTools(t *testing.T) {
	var businessCalls int32
	mcpServer := server.NewMCPServer("provider", "1.0.0")
	mcpServer.AddTool(mcp.NewTool("padel.court.availability", mcp.WithString("venue_id")), func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		atomic.AddInt32(&businessCalls, 1)
		return mcp.NewToolResultText("must not be called during discovery"), nil
	})
	testServer := server.NewTestStreamableHTTPServer(mcpServer)
	defer testServer.Close()

	tester := NewMCPConnectionTester(nil, nil, 0)
	result, err := tester.DiscoverTools(context.Background(), &entities.MCPConnection{
		Transport: entities.MCPTransportStreamableHTTP, Endpoint: stringPtr(testServer.URL), AuthType: entities.MCPAuthNone,
	})
	if err != nil || result.ErrorCode != "" {
		t.Fatalf("result = %#v, err = %v", result, err)
	}
	if len(result.Tools) != 1 || result.Tools[0].Name != "padel.court.availability" {
		t.Fatalf("discovered tools = %#v", result.Tools)
	}
	if atomic.LoadInt32(&businessCalls) != 0 {
		t.Fatalf("business tool calls = %d", businessCalls)
	}
}

func TestMCPConnectionTesterInvokesOnlyResolvedTool(t *testing.T) {
	mcpServer := server.NewMCPServer("provider", "1.0.0")
	mcpServer.AddTool(mcp.NewTool("padel.cek_available", mcp.WithString("venue_id")), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		return mcp.NewToolResultJSON(map[string]any{"available": args["venue_id"] == "padel-senayan"})
	})
	testServer := server.NewTestStreamableHTTPServer(mcpServer)
	defer testServer.Close()

	tester := NewMCPConnectionTester(nil, nil, 0)
	result, err := tester.InvokeTool(context.Background(), &entities.MCPConnection{
		Transport: entities.MCPTransportStreamableHTTP, Endpoint: stringPtr(testServer.URL), AuthType: entities.MCPAuthNone, Status: entities.MCPConnectionEnabled,
	}, &entities.MCPDiscoveredTool{
		Name: "padel.cek_available", Enabled: true, Availability: entities.MCPToolPresent,
	}, map[string]any{"venue_id": "padel-senayan"}, time.Second)
	if err != nil {
		t.Fatalf("invoke resolved tool: %v", err)
	}
	if result.Data["available"] != true {
		t.Fatalf("unexpected normalized result: %#v", result.Data)
	}
}

func stringPtr(value string) *string { return &value }
