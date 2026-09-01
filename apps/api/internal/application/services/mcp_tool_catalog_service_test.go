package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

type fakeMCPToolCatalogRepository struct {
	catalog      *entities.MCPToolCatalog
	failedRun    *entities.MCPDiscoveryRun
	setTool      *entities.MCPDiscoveredTool
	setEnabled   bool
	reconcile    repositories.MCPToolCatalogReconcileInput
	failureInput repositories.MCPToolCatalogFailureInput
}

func (f *fakeMCPToolCatalogRepository) Get(context.Context, string, string, string) (*entities.MCPToolCatalog, error) {
	if f.catalog == nil {
		return &entities.MCPToolCatalog{Tools: []entities.MCPDiscoveredTool{}}, nil
	}
	return f.catalog, nil
}

func (f *fakeMCPToolCatalogRepository) Reconcile(_ context.Context, input repositories.MCPToolCatalogReconcileInput) (*entities.MCPDiscoveryRun, error) {
	f.reconcile = input
	return &entities.MCPDiscoveryRun{Status: entities.MCPDiscoverySucceeded}, nil
}

func (f *fakeMCPToolCatalogRepository) RecordFailure(_ context.Context, input repositories.MCPToolCatalogFailureInput) (*entities.MCPDiscoveryRun, error) {
	f.failureInput = input
	f.failedRun = &entities.MCPDiscoveryRun{Status: entities.MCPDiscoveryFailed, ErrorCode: &input.ErrorCode}
	return f.failedRun, nil
}

func (f *fakeMCPToolCatalogRepository) SetEnabled(context.Context, string, string, string, string, bool) (*entities.MCPDiscoveredTool, error) {
	f.setEnabled = true
	return f.setTool, nil
}

type fakeMCPToolDiscoverer struct {
	result domainsvc.MCPToolDiscoveryResult
	err    error
}

func (f fakeMCPToolDiscoverer) DiscoverTools(context.Context, *entities.MCPConnection) (domainsvc.MCPToolDiscoveryResult, error) {
	return f.result, f.err
}

func TestSanitizeDiscoveredToolsRedactsProviderSecrets(t *testing.T) {
	tools, _, err := sanitizeDiscoveredTools([]domainsvc.MCPToolDefinition{{
		Name: "padel.court.availability", Description: "Use token=super-secret-value to access the provider.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{"token":{"type":"string","default":"secret-value"}}}`),
		Annotations: json.RawMessage(`{"readOnlyHint":true}`),
	}})
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(tools)
	if strings.Contains(string(encoded), "super-secret-value") || strings.Contains(string(encoded), "secret-value") {
		t.Fatalf("provider secret was retained: %s", encoded)
	}
}

func TestMCPToolCatalogRefreshRecordsFailureWithoutMutatingCatalog(t *testing.T) {
	previous := &entities.MCPToolCatalog{ConnectionID: testID, Tools: []entities.MCPDiscoveredTool{{Name: "existing", Availability: entities.MCPToolPresent}}}
	catalogRepo := &fakeMCPToolCatalogRepository{catalog: previous}
	connectionRepo := &fakeMCPConnectionRepository{item: &entities.MCPConnection{ID: testID, TenantID: testTenant, ProjectID: testProject, Status: entities.MCPConnectionEnabled}}
	service := NewMCPToolCatalogService(connectionRepo, catalogRepo, fakeMCPToolDiscoverer{
		result: domainsvc.MCPToolDiscoveryResult{ErrorCode: "mcp_discovery_timeout"}, err: errors.New("safe failure"),
	}, nil)

	_, err := service.Refresh(context.Background(), testTenant, testProject, testID, "actor")
	if err == nil {
		t.Fatal("expected discovery error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Details == nil || !strings.Contains(string(domainErr.Details), "mcp_discovery_timeout") {
		t.Fatalf("error = %#v", err)
	}
	if catalogRepo.failureInput.ErrorCode != "mcp_discovery_timeout" || len(previous.Tools) != 1 {
		t.Fatalf("failure = %#v, catalog = %#v", catalogRepo.failureInput, previous)
	}
}

func TestMCPToolCatalogRefreshSanitizesAndReconciles(t *testing.T) {
	catalogRepo := &fakeMCPToolCatalogRepository{}
	connectionRepo := &fakeMCPConnectionRepository{item: &entities.MCPConnection{ID: testID, TenantID: testTenant, ProjectID: testProject, Status: entities.MCPConnectionEnabled}}
	service := NewMCPToolCatalogService(connectionRepo, catalogRepo, fakeMCPToolDiscoverer{result: domainsvc.MCPToolDiscoveryResult{Tools: []domainsvc.MCPToolDefinition{{
		Name: "padel.court.availability", Description: "Check availability", InputSchema: json.RawMessage(`{"type":"object","properties":{}}`), Annotations: json.RawMessage(`{}`),
	}}}}, nil)

	_, err := service.Refresh(context.Background(), testTenant, testProject, testID, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if len(catalogRepo.reconcile.Tools) != 1 || catalogRepo.reconcile.Tools[0].Fingerprint == "" || catalogRepo.reconcile.CatalogFingerprint == "" {
		t.Fatalf("reconcile = %#v", catalogRepo.reconcile)
	}
}

func TestMCPToolCatalogRefreshRejectsDisabledConnection(t *testing.T) {
	connectionRepo := &fakeMCPConnectionRepository{item: &entities.MCPConnection{ID: testID, TenantID: testTenant, ProjectID: testProject, Status: entities.MCPConnectionDisabled}}
	service := NewMCPToolCatalogService(connectionRepo, &fakeMCPToolCatalogRepository{}, fakeMCPToolDiscoverer{}, nil)
	_, err := service.Refresh(context.Background(), testTenant, testProject, testID, "actor")
	if err == nil {
		t.Fatal("expected disabled connection error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrConflict {
		t.Fatalf("error = %#v", err)
	}
}
