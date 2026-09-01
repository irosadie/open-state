package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/entities"
)

func TestSecureServerExposesOnlyGatewayCapabilitySurface(t *testing.T) {
	deps := testDeps()
	deps.GatewayMode = appservices.MCPGatewayModeSecure
	server := NewServer(deps)
	tools := server.ListTools()
	if _, ok := tools["invoke_capability"]; !ok {
		t.Fatal("secure server must expose invoke_capability")
	}
	if _, ok := tools["report_capability_result"]; ok {
		t.Fatal("secure server must not expose advisory provider reporting")
	}

	client, err := client.NewInProcessClient(server)
	if err != nil {
		t.Fatalf("create in-process client: %v", err)
	}
	defer client.Close()
	result, err := client.Initialize(context.Background(), mcp.InitializeRequest{})
	if err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if !strings.Contains(result.Instructions, "enforced gateway") || strings.Contains(result.Instructions, "report_capability_result") {
		t.Fatalf("unexpected secure instructions: %q", result.Instructions)
	}
}

func TestSecureResolveIntentDoesNotProjectProviderRouting(t *testing.T) {
	deps := testDeps()
	deps.GatewayMode = appservices.MCPGatewayModeSecure
	deps.CapabilityRegistry = &fakeCapabilityRegistry{capability: entities.Capability{
		ID: "cap-padel", TenantID: "tenant-1", Name: "padel.availability.read", ProviderType: entities.ProviderTypeMCP,
	}}
	deps.ProjectCapabilityBindings = &fakeMCPBindingRepository{binding: entities.ProjectCapabilityMCPBinding{
		CapabilityID: "cap-padel", ConnectionAlias: "provider-secret-alias", ToolName: "provider.secret.tool", Health: entities.ProjectCapabilityMCPBindingActive,
	}}
	result, err := handleResolveIntent(context.Background(), deps, "tenant-1", "project-padel", "BOOKING_PADEL")
	if err != nil {
		t.Fatalf("resolve intent: %v", err)
	}
	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}
	var payload struct {
		Required []entities.ProviderRequirement `json:"requiredCapabilities"`
	}
	if err := json.Unmarshal([]byte(content.Text), &payload); err != nil {
		t.Fatalf("decode requirement response: %v", err)
	}
	if len(payload.Required) != 1 || payload.Required[0].ProviderServer != "" || payload.Required[0].Tool != "" || payload.Required[0].Status != "PENDING" {
		t.Fatalf("secure provider projection = %#v", payload.Required)
	}
}
