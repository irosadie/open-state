package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
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

func testPrincipal(tenantID string) entities.APIKeyPrincipal {
	defaultProjectID := "project-padel"
	return entities.APIKeyPrincipal{
		KeyID: "key-test", TenantID: tenantID, KeyPrefix: "osk_test",
		ProjectIDs: []string{"project-padel", "project-food", "retail"}, DefaultProjectID: &defaultProjectID,
		Scopes: []entities.MCPAPIScope{
			entities.MCPAPIScopeStateRead,
			entities.MCPAPIScopeStateWrite,
			entities.MCPAPIScopeCapabilityInvoke,
		},
	}
}

func newAuthenticatedTestServer(srv *server.MCPServer, principal entities.APIKeyPrincipal) *httptest.Server {
	mux := http.NewServeMux()
	streamable := server.NewStreamableHTTPServer(srv)
	mux.Handle("/mcp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		streamable.ServeHTTP(w, r.WithContext(WithAPIKeyPrincipal(r.Context(), principal)))
	}))
	return httptest.NewServer(mux)
}

func (fakeIntentResolver) ListIntents(_ context.Context, _, projectID string) ([]entities.Intent, error) {
	if projectID != "project-padel" {
		return []entities.Intent{}, nil
	}
	return []entities.Intent{{
		ProjectID: "project-padel", Key: "BOOKING_PADEL", Name: "Booking Lapangan Padel",
		Description: "Booking lapangan padel", Examples: []string{"saya mau order lapangan"}, WorkflowSlug: "padel-booking",
	}}, nil
}

func (fakeIntentResolver) ResolveIntent(_ context.Context, _ string, projectID, intentID string) (*entities.Workflow, error) {
	if intentID != "BOOKING_PADEL" {
		return nil, domain.NewNotFound("intent not found")
	}
	definition, _ := json.Marshal(engine.WorkflowDefinition{
		Slug: "padel-booking", ProjectID: projectID, EntryNodeID: "start",
		Nodes:       []engine.WorkflowNode{{ID: "start", Kind: engine.NodeKindStart, Capabilities: []string{"padel.availability.read"}}},
		Transitions: []engine.TransitionDefinition{{Event: "availability.confirmed", SourceStateID: "start", TargetStateID: "done"}},
	})
	return &entities.Workflow{ID: intentID, Slug: "padel-booking", Name: "Padel", Status: entities.WorkflowPublished, DraftDefinition: definition}, nil
}

type fakeCapabilityRegistry struct {
	capability entities.Capability
}

type fakeMCPBindingRepository struct {
	binding entities.ProjectCapabilityMCPBinding
}

func (f *fakeMCPBindingRepository) ListEligibleToolOptions(context.Context, string, string) ([]entities.ProjectMCPToolOption, error) {
	return nil, nil
}

func (f *fakeMCPBindingRepository) ListByProject(context.Context, string, string) ([]entities.ProjectCapabilityMCPBinding, error) {
	return []entities.ProjectCapabilityMCPBinding{f.binding}, nil
}

func (f *fakeMCPBindingRepository) FindByCapability(_ context.Context, _, _, capabilityID string) (*entities.ProjectCapabilityMCPBinding, error) {
	if capabilityID != f.binding.CapabilityID {
		return nil, domain.NewNotFound("binding not found")
	}
	return &f.binding, nil
}

func (f *fakeMCPBindingRepository) Upsert(context.Context, repositories.ProjectCapabilityMCPBindingUpsertInput) error {
	return nil
}

func (f *fakeMCPBindingRepository) Delete(context.Context, string, string, string) error {
	return nil
}

func (f *fakeCapabilityRegistry) Create(context.Context, string, string, *string, entities.ProviderType, *string, *string, []byte, []byte, int, *string) (*entities.Capability, error) {
	return &f.capability, nil
}
func (f *fakeCapabilityRegistry) FindByID(context.Context, string, string) (*entities.Capability, error) {
	return &f.capability, nil
}
func (f *fakeCapabilityRegistry) FindByName(_ context.Context, tenantID, name string) (*entities.Capability, error) {
	if tenantID != f.capability.TenantID || name != f.capability.Name {
		return nil, domain.NewNotFound("capability not found")
	}
	return &f.capability, nil
}
func (f *fakeCapabilityRegistry) ListByTenant(context.Context, string) ([]entities.Capability, error) {
	return []entities.Capability{f.capability}, nil
}
func (f *fakeCapabilityRegistry) ListByTenantFiltered(context.Context, string, entities.ProviderType, entities.CapabilityStatus) ([]entities.Capability, error) {
	return []entities.Capability{f.capability}, nil
}
func (f *fakeCapabilityRegistry) Update(context.Context, string, string, *string, entities.ProviderType, *string, *string, []byte, []byte, entities.CapabilityStatus, int, *string) (*entities.Capability, error) {
	return &f.capability, nil
}
func (f *fakeCapabilityRegistry) UpdateStatus(context.Context, string, string, entities.CapabilityStatus) (*entities.Capability, error) {
	return &f.capability, nil
}
func (f *fakeCapabilityRegistry) Disable(context.Context, string, string) (*entities.Capability, error) {
	return &f.capability, nil
}
func (f *fakeCapabilityRegistry) Bind(context.Context, string, string, entities.BindingScopeType, string, entities.BindingPermission) (*entities.CapabilityBinding, error) {
	return nil, nil
}
func (f *fakeCapabilityRegistry) ListBindingsByCapability(context.Context, string, string) ([]entities.CapabilityBinding, error) {
	return nil, nil
}
func (f *fakeCapabilityRegistry) ListBindingsByScope(context.Context, string, entities.BindingScopeType, string) ([]entities.CapabilityBinding, error) {
	return nil, nil
}
func (f *fakeCapabilityRegistry) Unbind(context.Context, string, string) error { return nil }
func (f *fakeCapabilityRegistry) UpsertPolicy(context.Context, string, entities.PolicyScopeType, string, string, []byte) (*entities.Policy, error) {
	return nil, nil
}
func (f *fakeCapabilityRegistry) FindPolicyByType(context.Context, string, entities.PolicyScopeType, string, string) (*entities.Policy, error) {
	return nil, nil
}
func (f *fakeCapabilityRegistry) ListPoliciesByScope(context.Context, string, entities.PolicyScopeType, string) ([]entities.Policy, error) {
	return nil, nil
}

func TestResolveIntentProjectsProviderRequirementWithoutSecrets(t *testing.T) {
	secret := "provider-secret-must-not-leak"
	deps := testDeps()
	deps.CapabilityRegistry = &fakeCapabilityRegistry{capability: entities.Capability{
		ID: "cap-padel", TenantID: "tenant-1", Name: "padel.availability.read",
		ProviderType:        entities.ProviderTypeMCP,
		ProviderID:          sql.NullString{String: "padel-provider-mock", Valid: true},
		ProviderTool:        sql.NullString{String: "padel.cek_available", Valid: true},
		CredentialReference: sql.NullString{String: secret, Valid: true},
	}}
	deps.ProjectCapabilityBindings = &fakeMCPBindingRepository{binding: entities.ProjectCapabilityMCPBinding{
		CapabilityID:    "cap-padel",
		ConnectionAlias: "padel-provider-mock",
		ToolName:        "padel.cek_available",
		Health:          entities.ProjectCapabilityMCPBindingActive,
	}}
	result, err := handleResolveIntent(context.Background(), deps, "tenant-1", "project-padel", "BOOKING_PADEL")
	if err != nil {
		t.Fatalf("resolve intent: %v", err)
	}
	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}
	if strings.Contains(content.Text, secret) || strings.Contains(strings.ToLower(content.Text), "credential") {
		t.Fatalf("provider requirement leaked credential metadata: %s", content.Text)
	}
	var payload struct {
		Required []entities.ProviderRequirement `json:"requiredCapabilities"`
	}
	if err := json.Unmarshal([]byte(content.Text), &payload); err != nil {
		t.Fatalf("decode requirement response: %v", err)
	}
	if len(payload.Required) != 1 || payload.Required[0].ProviderServer != "padel-provider-mock" || payload.Required[0].Tool != "padel.cek_available" {
		t.Fatalf("unexpected provider requirement: %+v", payload.Required)
	}
}

func TestServerRegistersTools(t *testing.T) {
	srv := NewServer(testDeps())
	tools := srv.ListTools()
	expected := []string{
		"list_intents",
		"resolve_intent",
		"get_active_workflow",
		"get_context",
		"invoke_capability",
		"get_current_state",
		"get_allowed_capabilities",
		"propose_event",
		"report_capability_result",
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

func TestServerInitializeDescribesStateGatekeeperProtocol(t *testing.T) {
	server := NewServer(testDeps())
	client, err := client.NewInProcessClient(server)
	if err != nil {
		t.Fatalf("create in-process client: %v", err)
	}
	defer client.Close()
	result, err := client.Initialize(context.Background(), mcp.InitializeRequest{})
	if err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if !strings.Contains(result.Instructions, "mandatory state controller") || !strings.Contains(result.Instructions, "report_capability_result") {
		t.Fatalf("unexpected State MCP instructions: %q", result.Instructions)
	}
}

func TestListIntentsTool(t *testing.T) {
	ts := newAuthenticatedTestServer(NewServer(testDeps()), testPrincipal("tenant-1"))
	defer ts.Close()

	c, err := client.NewStreamableHttpClient(ts.URL + "/mcp")
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if _, err := c.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	res, err := c.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name:      "list_intents",
		Arguments: map[string]any{"tenant": "tenant-1", "project": "project-padel"},
	}})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %+v", res.Content)
	}
	content, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", res.Content[0])
	}
	var payload struct {
		Intents []IntentInfo `json:"intents"`
	}
	if err := json.Unmarshal([]byte(content.Text), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Intents) != 1 || payload.Intents[0].ID != "BOOKING_PADEL" {
		t.Fatalf("unexpected intents: %+v", payload.Intents)
	}
	if len(payload.Intents[0].Examples) != 1 || payload.Intents[0].Examples[0] != "saya mau order lapangan" {
		t.Fatalf("unexpected examples: %+v", payload.Intents[0].Examples)
	}

	otherRes, err := c.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name:      "list_intents",
		Arguments: map[string]any{"tenant": "tenant-1", "project": "project-food"},
	}})
	if err != nil {
		t.Fatalf("call other project: %v", err)
	}
	otherContent, ok := otherRes.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content for other project, got %T", otherRes.Content[0])
	}
	var otherPayload struct {
		Intents []IntentInfo `json:"intents"`
	}
	if err := json.Unmarshal([]byte(otherContent.Text), &otherPayload); err != nil {
		t.Fatalf("decode other project response: %v", err)
	}
	if len(otherPayload.Intents) != 0 {
		t.Fatalf("expected project isolation, got %+v", otherPayload.Intents)
	}
}

func TestListIntentsRequiresScope(t *testing.T) {
	ts := newAuthenticatedTestServer(NewServer(testDeps()), testPrincipal("tenant-1"))
	defer ts.Close()

	c, err := client.NewStreamableHttpClient(ts.URL + "/mcp")
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if _, err := c.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	res, err := c.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: "list_intents", Arguments: map[string]any{"tenant": "tenant-1"},
	}})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	content, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", res.Content[0])
	}
	if strings.Contains(content.Text, `"ok":false`) {
		t.Fatalf("expected the API key default project to be used, got %s", content.Text)
	}
}

func TestResolveIntentRejectsWorkflowSlug(t *testing.T) {
	ts := newAuthenticatedTestServer(NewServer(testDeps()), testPrincipal("tenant-1"))
	defer ts.Close()

	c, err := client.NewStreamableHttpClient(ts.URL + "/mcp")
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if _, err := c.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	res, err := c.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name:      "resolve_intent",
		Arguments: map[string]any{"tenant": "tenant-1", "project": "project-padel", "intent": "padel-booking"},
	}})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	content, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", res.Content[0])
	}
	if !strings.Contains(content.Text, "intent not found") {
		t.Fatalf("expected not-found response, got %s", content.Text)
	}
}

func TestResolveIntentTool(t *testing.T) {
	ts := newAuthenticatedTestServer(NewServer(testDeps()), testPrincipal("tenant-1"))
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
	ts := newAuthenticatedTestServer(NewServer(testDeps()), testPrincipal("tenant-1"))
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
