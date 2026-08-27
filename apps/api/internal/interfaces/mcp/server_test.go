package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func testDeps() Dependencies {
	return Dependencies{
		IntentResolver: fakeIntentResolver{},
	}
}

type fakeIntentResolver struct{}

func (fakeIntentResolver) ListIntents() []IntentInfo {
	return []IntentInfo{
		{ID: "BOOKING_PADEL", ProjectID: "project-padel", Name: "Padel", WorkflowSlug: "padel-booking"},
	}
}

func TestServerRegistersFourTools(t *testing.T) {
	srv := NewServer(testDeps())
	tools := srv.ListTools()
	if len(tools) != 4 {
		t.Errorf("expected 4 tools, got %d: %v", len(tools), toolNames(tools))
	}
	for _, name := range []string{"resolve_intent", "get_active_workflow", "get_context", "invoke_capability"} {
		if _, ok := tools[name]; !ok {
			t.Errorf("missing tool %q", name)
		}
	}
}

func TestResolveIntentTool(t *testing.T) {
	ts := server.NewTestStreamableHTTPServer(NewServer(testDeps()))
	defer ts.Close()

	c, err := client.NewStreamableHttpClient(ts.URL + "/mcp")
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if _, err := c.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	call := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: "resolve_intent",
		Arguments: map[string]any{
			"intent":  "BOOKING_PADEL",
			"project": "project-padel",
		},
	}}
	res, err := c.CallTool(context.Background(), call)
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %+v", res.Content)
	}
}

func toolNames(tools map[string]*server.ServerTool) []string {
	var names []string
	for k := range tools {
		names = append(names, k)
	}
	return names
}
