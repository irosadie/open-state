package services

import (
	"context"
	"errors"
	"testing"
	"time"

	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	domainservices "github.com/irosadie/open-state/api/internal/domain/services"
)

type gatewayRuntimeFake struct {
	instance *entities.WorkflowInstance
	info     *engine.StateInfo
}

func (f gatewayRuntimeFake) GetCurrentState(context.Context, string, string) (*entities.WorkflowInstance, *entities.StateInstance, error) {
	return f.instance, nil, nil
}

func (f gatewayRuntimeFake) CurrentStateInfo(context.Context, string, string) (*engine.StateInfo, error) {
	return f.info, nil
}

type gatewayProviderFake struct {
	calls    int
	result   domainservices.MCPToolCallResult
	err      error
	toolName string
}

func (f *gatewayProviderFake) InvokeTool(_ context.Context, _ *entities.MCPConnection, tool *entities.MCPDiscoveredTool, _ map[string]any, _ time.Duration) (domainservices.MCPToolCallResult, error) {
	f.calls++
	if tool != nil {
		f.toolName = tool.Name
	}
	return f.result, f.err
}

type gatewayValidatorFake struct{ err error }

func (f gatewayValidatorFake) Validate(map[string]any, []byte) error { return f.err }

func newGatewayServiceForTest(provider *gatewayProviderFake, evidence *fakeCapabilityEvidenceRepo, bindingHealth entities.ProjectCapabilityMCPBindingHealth) *MCPGatewayService {
	const capabilityName = "padel.availability.read"
	caps := &fakeCapRepo{caps: map[string]*entities.Capability{
		capabilityName: {
			ID: bindingCapabilityID, TenantID: testTenant, Name: capabilityName,
			ProviderType: entities.ProviderTypeMCP, Status: entities.CapabilityActive,
		},
	}}
	bindings := &fakeProjectCapabilityMCPBindingRepository{items: []entities.ProjectCapabilityMCPBinding{{
		CapabilityID: bindingCapabilityID, ConnectionID: bindingConnectionID, ToolID: bindingToolID,
		ConnectionAlias: "padel-provider", ToolName: "padel.cek_available", BoundToolFingerprint: "fp-1", Health: bindingHealth,
	}}}
	connections := &fakeMCPConnectionRepository{item: &entities.MCPConnection{
		ID: bindingConnectionID, TenantID: testTenant, ProjectID: testProject,
		Status: entities.MCPConnectionEnabled,
	}}
	catalog := &fakeMCPToolCatalogRepository{catalog: &entities.MCPToolCatalog{Tools: []entities.MCPDiscoveredTool{{
		ID: bindingToolID, TenantID: testTenant, ProjectID: testProject, ConnectionID: bindingConnectionID,
		Name: "padel.cek_available", Fingerprint: "fp-1", Enabled: true, Availability: entities.MCPToolPresent,
	}}}}
	runtime := gatewayRuntimeFake{
		instance: &entities.WorkflowInstance{ID: "instance-1", TenantID: testTenant, WorkflowID: "workflow-1", Status: entities.WorkflowInstanceRunning},
		info:     &engine.StateInfo{ProjectID: testProject, StateID: "state-availability", Capabilities: []string{capabilityName}},
	}
	return NewMCPGatewayService(runtime, caps, bindings, connections, catalog, evidence, nil, nil, provider, nil, nil, time.Second)
}

func gatewayRequest() GatewayInvocationRequest {
	return GatewayInvocationRequest{
		TenantID: testTenant, InstanceID: "instance-1", CapabilityName: "padel.availability.read",
		Payload:       map[string]any{"venue_id": "padel-senayan", "date": "2026-09-01"},
		CorrelationID: "corr-1", IdempotencyKey: "availability-1",
	}
}

func TestMCPGatewayAuthorizesCurrentStateAndReusesSuccessfulEvidence(t *testing.T) {
	provider := &gatewayProviderFake{result: domainservices.MCPToolCallResult{Data: map[string]any{"available": true}}}
	evidence := &fakeCapabilityEvidenceRepo{}
	service := newGatewayServiceForTest(provider, evidence, entities.ProjectCapabilityMCPBindingActive)

	first, err := service.Execute(context.Background(), gatewayRequest())
	if err != nil {
		t.Fatalf("first gateway execution: %v", err)
	}
	second, err := service.Execute(context.Background(), gatewayRequest())
	if err != nil {
		t.Fatalf("second gateway execution: %v", err)
	}
	if provider.calls != 1 || !second.Reused || first.Data["available"] != true {
		t.Fatalf("calls=%d first=%#v second=%#v", provider.calls, first, second)
	}
	if provider.toolName != "padel.cek_available" {
		t.Fatalf("resolved tool = %q", provider.toolName)
	}
}

func TestMCPGatewayRejectsCapabilityNotDeclaredByCurrentState(t *testing.T) {
	provider := &gatewayProviderFake{result: domainservices.MCPToolCallResult{Data: map[string]any{"available": true}}}
	service := newGatewayServiceForTest(provider, &fakeCapabilityEvidenceRepo{}, entities.ProjectCapabilityMCPBindingActive)
	req := gatewayRequest()
	req.CapabilityName = "padel.create_booking"
	_, err := service.Execute(context.Background(), req)
	var capabilityErr *domaincap.CapabilityError
	if !errors.As(err, &capabilityErr) || capabilityErr.Kind != domaincap.ErrorKindUnauthorized || provider.calls != 0 {
		t.Fatalf("err=%v calls=%d", err, provider.calls)
	}
}

func TestMCPGatewayRejectsUnhealthyBindingBeforeProvider(t *testing.T) {
	provider := &gatewayProviderFake{result: domainservices.MCPToolCallResult{Data: map[string]any{"available": true}}}
	service := newGatewayServiceForTest(provider, &fakeCapabilityEvidenceRepo{}, entities.ProjectCapabilityMCPBindingToolDisabled)
	_, err := service.Execute(context.Background(), gatewayRequest())
	var capabilityErr *domaincap.CapabilityError
	if !errors.As(err, &capabilityErr) || capabilityErr.Kind != domaincap.ErrorKindUnavailable || provider.calls != 0 {
		t.Fatalf("err=%v calls=%d", err, provider.calls)
	}
}

func TestMCPGatewayProviderFailureRecordsFailedEvidenceAndCannotBeReused(t *testing.T) {
	provider := &gatewayProviderFake{err: domaincap.NewCapabilityError(domaincap.ErrorKindTimeout, "capability.timeout", "MCP capability invocation timed out")}
	evidence := &fakeCapabilityEvidenceRepo{}
	service := newGatewayServiceForTest(provider, evidence, entities.ProjectCapabilityMCPBindingActive)
	_, err := service.Execute(context.Background(), gatewayRequest())
	if err == nil || len(evidence.records) != 1 || evidence.records[0].Status != entities.CapabilityEvidenceFailed {
		t.Fatalf("err=%v evidence=%#v", err, evidence.records)
	}
	_, err = service.Execute(context.Background(), gatewayRequest())
	var capabilityErr *domaincap.CapabilityError
	if !errors.As(err, &capabilityErr) || capabilityErr.Code != "capability.idempotency_conflict" || provider.calls != 1 {
		t.Fatalf("err=%v calls=%d", err, provider.calls)
	}
}

func TestMCPGatewayRejectsInvalidNormalizedOutput(t *testing.T) {
	provider := &gatewayProviderFake{result: domainservices.MCPToolCallResult{Data: map[string]any{"available": true}}}
	evidence := &fakeCapabilityEvidenceRepo{}
	service := newGatewayServiceForTest(provider, evidence, entities.ProjectCapabilityMCPBindingActive)
	service.capabilities.(*fakeCapRepo).caps["padel.availability.read"].OutputSchema = []byte(`{"type":"object","required":["available"]}`)
	service.validator = gatewayValidatorFake{err: errors.New("output does not match schema")}
	_, err := service.Execute(context.Background(), gatewayRequest())
	var capabilityErr *domaincap.CapabilityError
	if !errors.As(err, &capabilityErr) || capabilityErr.Code != "capability.output_invalid" || len(evidence.records) != 1 || evidence.records[0].Status != entities.CapabilityEvidenceFailed {
		t.Fatalf("err=%v evidence=%#v", err, evidence.records)
	}
}

func TestMCPGatewayRequiresCorrelationAndIdempotency(t *testing.T) {
	service := newGatewayServiceForTest(&gatewayProviderFake{}, &fakeCapabilityEvidenceRepo{}, entities.ProjectCapabilityMCPBindingActive)
	req := gatewayRequest()
	req.IdempotencyKey = ""
	_, err := service.Execute(context.Background(), req)
	var capabilityErr *domaincap.CapabilityError
	if !errors.As(err, &capabilityErr) || capabilityErr.Kind != domaincap.ErrorKindValidation {
		t.Fatalf("err=%v", err)
	}
}
