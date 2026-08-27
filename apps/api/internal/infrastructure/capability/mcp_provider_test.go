package capability

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
)

func TestEnvCredentialResolver(t *testing.T) {
	_ = os.Setenv("CRED_PAYMENT_PROD", "sekret-value")
	defer os.Unsetenv("CRED_PAYMENT_PROD")

	r := EnvCredentialResolver{Prefix: "CRED_"}
	v, ok := r.Resolve("payment-prod")
	if !ok || v != "sekret-value" {
		t.Fatalf("resolve failed: %q %v", v, ok)
	}
	if _, ok := r.Resolve("missing-key"); ok {
		t.Error("missing key should resolve false")
	}
}

func TestSanitizeEnvKey(t *testing.T) {
	if got := sanitizeEnvKey("payment.prod"); got != "PAYMENT_PROD" {
		t.Errorf("got %q", got)
	}
}

// buildEchoMCPServer returns a test MCP server exposing invoke_capability that
// echoes arguments back (used to validate the client adapter).
func buildEchoMCPServer() *server.MCPServer {
	srv := server.NewMCPServer("echo", "0.0.1")
	tool := mcp.NewTool("invoke_capability",
		mcp.WithString("capability", mcp.Required(), mcp.Description("capability name")),
	)
	srv.AddTool(tool, func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]any)
		return mcp.NewToolResultJSON(map[string]any{"echo": args})
	})
	return srv
}

func TestMCPProviderInvoke(t *testing.T) {
	ts := server.NewTestStreamableHTTPServer(buildEchoMCPServer())
	defer ts.Close()

	p := NewMCPProvider(MCPProviderConfig{
		Endpoint: ts.URL + "/mcp",
		ToolName: "invoke_capability",
		Timeout:  3 * time.Second,
	}, nil)

	inv := domaincap.Invocation{
		Name:    "payment.create",
		Payload: map[string]any{"capability": "payment.create"},
	}
	res, err := p.Invoke(context.Background(), inv)
	if err != nil {
		t.Fatalf("invoke error: %v", err)
	}
	if res.FromMock {
		t.Error("real provider must not be FromMock")
	}
	if res.CapabilityEvent == nil || *res.CapabilityEvent != "capability.success" {
		t.Errorf("expected success event, got %v", res.CapabilityEvent)
	}
}

func TestMCPProviderTimeout(t *testing.T) {
	// invalid endpoint → connection error → classified EXTERNAL/unavailable
	p := NewMCPProvider(MCPProviderConfig{Endpoint: "http://127.0.0.1:1/nope", Timeout: 200 * time.Millisecond}, nil)
	_, err := p.Invoke(context.Background(), domaincap.Invocation{Name: "x"})
	if err == nil {
		t.Fatal("expected error on unreachable endpoint")
	}
	var ce *domaincap.CapabilityError
	if !asCapErr(err, &ce) {
		t.Fatalf("expected CapabilityError, got %T", err)
	}
	if ce.Kind != domaincap.ErrorKindUnavailable && ce.Kind != domaincap.ErrorKindExternal {
		t.Errorf("expected unavailable/external, got %s", ce.Kind)
	}
}

func asCapErr(err error, target **domaincap.CapabilityError) bool {
	ce, ok := err.(*domaincap.CapabilityError)
	if !ok {
		return false
	}
	*target = ce
	return true
}
