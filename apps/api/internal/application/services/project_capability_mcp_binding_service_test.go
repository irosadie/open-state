package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

const (
	bindingCapabilityID = "00000000-0000-0000-0000-000000000004"
	bindingConnectionID = "00000000-0000-0000-0000-000000000005"
	bindingToolID       = "00000000-0000-0000-0000-000000000006"
	otherConnectionID   = "00000000-0000-0000-0000-000000000007"
)

type fakeProjectCapabilityMCPBindingRepository struct {
	items       []entities.ProjectCapabilityMCPBinding
	findErr     error
	upsertInput repositories.ProjectCapabilityMCPBindingUpsertInput
}

func (f *fakeProjectCapabilityMCPBindingRepository) ListEligibleToolOptions(context.Context, string, string) ([]entities.ProjectMCPToolOption, error) {
	return nil, nil
}

func (f *fakeProjectCapabilityMCPBindingRepository) ListByProject(context.Context, string, string) ([]entities.ProjectCapabilityMCPBinding, error) {
	return f.items, nil
}

func (f *fakeProjectCapabilityMCPBindingRepository) FindByCapability(_ context.Context, _, _, capabilityID string) (*entities.ProjectCapabilityMCPBinding, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	for i := range f.items {
		if f.items[i].CapabilityID == capabilityID {
			return &f.items[i], nil
		}
	}
	return nil, domain.NewNotFound("binding not found")
}

func (f *fakeProjectCapabilityMCPBindingRepository) Upsert(_ context.Context, input repositories.ProjectCapabilityMCPBindingUpsertInput) error {
	f.upsertInput = input
	return nil
}

func (f *fakeProjectCapabilityMCPBindingRepository) Delete(context.Context, string, string, string) error {
	return nil
}

func newBindingCapability() *entities.Capability {
	return &entities.Capability{
		ID:           bindingCapabilityID,
		TenantID:     testTenant,
		Name:         "padel.availability.read",
		ProviderType: entities.ProviderTypeMCP,
		Status:       entities.CapabilityActive,
	}
}

func newBindingServiceForTest(
	caps repositories.ICapabilityRepository,
	connections repositories.IMCPConnectionRepository,
	catalog repositories.IMCPToolCatalogRepository,
	bindings repositories.IProjectCapabilityMCPBindingRepository,
	projects repositories.IProjectRepository,
) *ProjectCapabilityMCPBindingService {
	return NewProjectCapabilityMCPBindingService(bindings, caps, connections, catalog, projects, nil)
}

func TestProjectCapabilityMCPBindingUpsertRejectsToolFromAnotherConnection(t *testing.T) {
	caps := &fakeCapRepo{caps: map[string]*entities.Capability{
		"padel.availability.read": newBindingCapability(),
	}}
	connections := &fakeMCPConnectionRepository{item: &entities.MCPConnection{
		ID: bindingConnectionID, TenantID: testTenant, ProjectID: testProject,
		Status: entities.MCPConnectionEnabled,
	}}
	catalog := &fakeMCPToolCatalogRepository{catalog: &entities.MCPToolCatalog{Tools: []entities.MCPDiscoveredTool{{
		ID: bindingToolID, TenantID: testTenant, ProjectID: testProject, ConnectionID: otherConnectionID,
		Name: "padel.cek_available", Fingerprint: "fp-1", Enabled: true, Availability: entities.MCPToolPresent,
	}}}}
	service := newBindingServiceForTest(caps, connections, catalog, &fakeProjectCapabilityMCPBindingRepository{}, fakeProjectRepository{})

	_, err := service.Upsert(context.Background(), testTenant, testProject, bindingCapabilityID, "actor", dtos.UpsertProjectCapabilityMCPBindingRequest{
		ConnectionID: bindingConnectionID,
		ToolID:       bindingToolID,
	})
	if err == nil {
		t.Fatal("expected cross-connection tool to be rejected")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrNotFound {
		t.Fatalf("error = %#v", err)
	}
}

func TestProjectCapabilityMCPBindingUpsertRejectsConnectionFromAnotherProject(t *testing.T) {
	caps := &fakeCapRepo{caps: map[string]*entities.Capability{
		"padel.availability.read": newBindingCapability(),
	}}
	connections := &fakeMCPConnectionRepository{item: &entities.MCPConnection{
		ID: bindingConnectionID, TenantID: testTenant, ProjectID: "00000000-0000-0000-0000-000000000099",
		Status: entities.MCPConnectionEnabled,
	}}
	service := newBindingServiceForTest(caps, connections, &fakeMCPToolCatalogRepository{}, &fakeProjectCapabilityMCPBindingRepository{}, fakeProjectRepository{})

	_, err := service.Upsert(context.Background(), testTenant, testProject, bindingCapabilityID, "actor", dtos.UpsertProjectCapabilityMCPBindingRequest{
		ConnectionID: bindingConnectionID,
		ToolID:       bindingToolID,
	})
	if err == nil {
		t.Fatal("expected cross-project connection to be rejected")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrNotFound {
		t.Fatalf("error = %#v", err)
	}
}

func TestProjectCapabilityMCPBindingUpsertRejectsUnhealthyTool(t *testing.T) {
	caps := &fakeCapRepo{caps: map[string]*entities.Capability{
		"padel.availability.read": newBindingCapability(),
	}}
	connections := &fakeMCPConnectionRepository{item: &entities.MCPConnection{
		ID: bindingConnectionID, TenantID: testTenant, ProjectID: testProject,
		Status: entities.MCPConnectionEnabled,
	}}
	catalog := &fakeMCPToolCatalogRepository{catalog: &entities.MCPToolCatalog{Tools: []entities.MCPDiscoveredTool{{
		ID: bindingToolID, TenantID: testTenant, ProjectID: testProject, ConnectionID: bindingConnectionID,
		Name: "padel.cek_available", Fingerprint: "fp-1", Enabled: false, Availability: entities.MCPToolPresent,
	}}}}
	service := newBindingServiceForTest(caps, connections, catalog, &fakeProjectCapabilityMCPBindingRepository{}, fakeProjectRepository{})

	_, err := service.Upsert(context.Background(), testTenant, testProject, bindingCapabilityID, "actor", dtos.UpsertProjectCapabilityMCPBindingRequest{
		ConnectionID: bindingConnectionID,
		ToolID:       bindingToolID,
	})
	if err == nil {
		t.Fatal("expected disabled tool to be rejected")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrConflict {
		t.Fatalf("error = %#v", err)
	}
}

func TestProjectCapabilityMCPBindingListIncludesMissingMapping(t *testing.T) {
	caps := &fakeCapRepo{caps: map[string]*entities.Capability{
		"padel.availability.read": newBindingCapability(),
	}}
	service := newBindingServiceForTest(caps, nil, nil, &fakeProjectCapabilityMCPBindingRepository{}, fakeProjectRepository{})

	result, err := service.List(context.Background(), testTenant, testProject)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 1 || result.Data[0].Health != string(entities.ProjectCapabilityMCPBindingMissingMapping) {
		t.Fatalf("unexpected bindings = %#v", result.Data)
	}
}

func TestValidateWorkflowReportsUnhealthyBinding(t *testing.T) {
	caps := &fakeCapRepo{caps: map[string]*entities.Capability{
		"padel.availability.read": newBindingCapability(),
	}}
	bindings := &fakeProjectCapabilityMCPBindingRepository{items: []entities.ProjectCapabilityMCPBinding{{
		CapabilityID: bindingCapabilityID,
		Health:       entities.ProjectCapabilityMCPBindingToolRemoved,
		HealthReason: "Tool is no longer present in the latest catalog",
	}}}
	service := newBindingServiceForTest(caps, nil, nil, bindings, fakeProjectRepository{})
	definition, err := json.Marshal(engine.WorkflowDefinition{
		Nodes: []engine.WorkflowNode{{ID: "check-availability", Kind: engine.NodeKindState, Capabilities: []string{"padel.availability.read"}}},
	})
	if err != nil {
		t.Fatal(err)
	}

	validationErr, err := service.ValidateWorkflow(context.Background(), testTenant, testProject, definition)
	if err != nil {
		t.Fatal(err)
	}
	if validationErr == nil || validationErr.Code != domain.ErrValidation || !strings.Contains(string(validationErr.Details), "MCP_BINDING_UNAVAILABLE") {
		t.Fatalf("unexpected validation error = %#v", validationErr)
	}
}

func TestBuilderPublishRejectsMissingMCPBinding(t *testing.T) {
	caps := &fakeCapRepo{caps: map[string]*entities.Capability{
		"padel.availability.read": newBindingCapability(),
	}}
	bindingService := newBindingServiceForTest(caps, nil, nil, &fakeProjectCapabilityMCPBindingRepository{}, &fakeProjectRepo{
		projects: map[string]*entities.Project{testProject: {ID: testProject, TenantID: testTenant}},
	})
	wfRepo := &fakeWorkflowRepo{workflows: map[string]*entities.Workflow{
		"workflow-1": {
			ID: "workflow-1", TenantID: testTenant, ProjectID: testProject, Version: 1,
			DraftDefinition: workflowWithCapabilityDefinition(t),
		},
	}}
	service := NewBuilderService(wfRepo, &fakeProjectRepo{
		projects: map[string]*entities.Project{testProject: {ID: testProject, TenantID: testTenant}},
	}, nil, bindingService)

	_, err := service.Publish(context.Background(), testTenant, testProject, "workflow-1", "actor", dtos.PublishWorkflowRequest{Version: 1})
	if err == nil {
		t.Fatal("expected publish to reject missing MCP binding")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrValidation || len(wfRepo.versions) != 0 {
		t.Fatalf("error = %#v, versions = %#v", err, wfRepo.versions)
	}
}

func workflowWithCapabilityDefinition(t *testing.T) []byte {
	t.Helper()
	definition, err := json.Marshal(engine.WorkflowDefinition{
		Nodes: []engine.WorkflowNode{
			{ID: "start", Kind: engine.NodeKindStart, Capabilities: []string{"padel.availability.read"}},
			{ID: "done", Kind: engine.NodeKindEnd},
		},
		Transitions: []engine.TransitionDefinition{{ID: "to-done", SourceStateID: "start", TargetStateID: "done", Event: "done"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return definition
}

var _ repositories.IProjectCapabilityMCPBindingRepository = (*fakeProjectCapabilityMCPBindingRepository)(nil)
