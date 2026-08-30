package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

func testDeps() Dependencies {
	return Dependencies{
		IntentResolver: fakeIntentResolver{},
		Orchestrator:   fakeOrchestrator{},
	}
}

type fakeOrchestrator struct{}

type mcpTraceRepo struct {
	entries []entities.RuntimeTraceEntry
}

func (r *mcpTraceRepo) Append(_ context.Context, tenantID string, input repositories.AppendRuntimeTraceInput) (*entities.RuntimeTraceEntry, error) {
	entry := entities.RuntimeTraceEntry{
		ID:                 "trace-1",
		TenantID:           tenantID,
		WorkflowInstanceID: input.WorkflowInstanceID,
		Sequence:           int64(len(r.entries) + 1),
		Stage:              input.Stage,
		Source:             input.Source,
		Status:             input.Status,
		Attributes:         input.Attributes,
	}
	r.entries = append(r.entries, entry)
	return &r.entries[len(r.entries)-1], nil
}

func (r *mcpTraceRepo) ListByInstance(context.Context, string, string) ([]entities.RuntimeTraceEntry, error) {
	return r.entries, nil
}

func (r *mcpTraceRepo) ListByTurn(context.Context, string, string, string) ([]entities.RuntimeTraceEntry, error) {
	return r.entries, nil
}

func (fakeOrchestrator) StartWorkflow(_ context.Context, _ string, workflowID, versionID, correlation string) (*entities.WorkflowInstance, error) {
	return &entities.WorkflowInstance{ID: "inst-1", WorkflowID: workflowID, WorkflowVersionID: versionID, Status: entities.WorkflowInstanceRunning}, nil
}
func (fakeOrchestrator) SuspendWorkflow(context.Context, string, string) (*entities.WorkflowInstance, error) {
	return nil, nil
}
func (fakeOrchestrator) ResumeWorkflow(context.Context, string, string) (*entities.WorkflowInstance, error) {
	return nil, nil
}
func (fakeOrchestrator) CancelWorkflow(context.Context, string, string) (*entities.WorkflowInstance, error) {
	return nil, nil
}
func (fakeOrchestrator) ListInstances(context.Context, string) ([]entities.WorkflowInstance, error) {
	return nil, nil
}
func (fakeOrchestrator) GetCurrentState(context.Context, string, string) (*entities.WorkflowInstance, *entities.StateInstance, error) {
	return nil, nil, nil
}
func (fakeOrchestrator) GetActiveWorkflow(context.Context, string, string) (*entities.WorkflowInstance, error) {
	return nil, nil
}
func (fakeOrchestrator) ListHistory(context.Context, string, string) ([]entities.Event, error) {
	return nil, nil
}
func (fakeOrchestrator) ReplayWorkflow(context.Context, string, string) (map[string]any, *entities.Event, error) {
	return map[string]any{}, nil, nil
}
func (fakeOrchestrator) ProposeEvent(context.Context, string, string, string, map[string]any, string) (*entities.Event, error) {
	return nil, nil
}
func (fakeOrchestrator) ListAllowedCapabilities(context.Context, string, entities.BindingScopeType, string) ([]entities.Capability, error) {
	return nil, nil
}

func (fakeOrchestrator) GetAllowedTransitions(context.Context, string, string) ([]engine.TransitionDefinition, error) {
	return nil, nil
}

func (fakeOrchestrator) CurrentStateInfo(context.Context, string, string) (*engine.StateInfo, error) {
	return &engine.StateInfo{}, nil
}

func (fakeOrchestrator) ReplayState(context.Context, string, string) (map[string]any, string, error) {
	return map[string]any{}, "", nil
}

type fakeIntentResolver struct{}

func (fakeIntentResolver) ListIntents(context.Context, string, string) ([]entities.Workflow, error) {
	return []entities.Workflow{{ID: "BOOKING_PADEL", Slug: "padel-booking", Name: "Padel"}}, nil
}

func (fakeIntentResolver) ResolveIntent(_ context.Context, _ string, projectID, intentID string) (*entities.Workflow, error) {
	return &entities.Workflow{ID: intentID, Slug: "padel-booking", Name: "Padel", Status: entities.WorkflowPublished}, nil
}

func TestServerRegistersTools(t *testing.T) {
	srv := NewServer(testDeps())
	tools := srv.ListTools()
	expected := []string{
		"resolve_intent",
		"get_active_workflow",
		"get_context",
		"invoke_capability",
		"get_current_state",
		"get_allowed_capabilities",
		"propose_event",
		"start_workflow",
		"suspend_workflow",
		"resume_workflow",
		"cancel_workflow",
		"get_workflow_instances",
		"get_history",
		"replay_workflow",
	}
	if len(tools) != len(expected) {
		t.Errorf("expected %d tools, got %d: %v", len(expected), len(tools), toolNames(tools))
	}
	for _, name := range expected {
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
			"tenant":  "tenant-1",
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

func TestStartWorkflowToolDispatch(t *testing.T) {
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
		Name: "start_workflow",
		Arguments: map[string]any{
			"tenant":   "tenant-1",
			"workflow": "wf-1",
			"version":  "wv-1",
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

func TestMCPRuntimeBoundaryRecordsOnlyApplicationMetadata(t *testing.T) {
	repo := &mcpTraceRepo{}
	deps := testDeps()
	deps.TraceRecorder = appservices.NewRuntimeTraceRecorder(repo)

	if _, err := handleStartWorkflow(context.Background(), deps, "tenant-1", "workflow-1", "version-1", "conversation-1"); err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	if len(repo.entries) != 1 {
		t.Fatalf("expected one trace entry, got %d", len(repo.entries))
	}
	entry := repo.entries[0]
	if entry.Stage != entities.RuntimeTraceStageWorkflowLookup || entry.Source != entities.RuntimeTraceSourceOpenState {
		t.Fatalf("unexpected trace boundary: %+v", entry)
	}
	if _, ok := entry.Attributes["raw_response"]; ok {
		t.Fatal("raw provider data should not be part of an application boundary")
	}
}
