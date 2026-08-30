package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

type adminRuntimeFake struct {
	instance *entities.WorkflowInstance
	called   string
}

func (f *adminRuntimeFake) SuspendWorkflow(_ context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error) {
	f.called = "suspend:" + tenantID + ":" + instanceID
	return f.instance, nil
}
func (f *adminRuntimeFake) ResumeWorkflow(_ context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error) {
	f.called = "resume:" + tenantID + ":" + instanceID
	return f.instance, nil
}
func (f *adminRuntimeFake) RetryWorkflow(_ context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error) {
	f.called = "retry:" + tenantID + ":" + instanceID
	return f.instance, nil
}
func (f *adminRuntimeFake) ListInstances(context.Context, string) ([]entities.WorkflowInstance, error) {
	return []entities.WorkflowInstance{*f.instance}, nil
}

type adminEventBrowserFake struct {
	filter repositories.EventFilter
	tenant string
}

func (f *adminEventBrowserFake) FindEventByID(_ context.Context, tenantID, _ string) (*entities.Event, error) {
	f.tenant = tenantID
	return &entities.Event{ID: "event-1", TenantID: tenantID, EventID: "event-1", Payload: json.RawMessage(`{"safe":true}`), Timestamp: time.Now(), CreatedAt: time.Now()}, nil
}
func (f *adminEventBrowserFake) ListEventsFiltered(_ context.Context, tenantID string, filter repositories.EventFilter) ([]entities.Event, error) {
	f.tenant, f.filter = tenantID, filter
	return []entities.Event{}, nil
}
func (f *adminEventBrowserFake) CountEventsFiltered(_ context.Context, tenantID string, filter repositories.EventFilter) (int64, error) {
	f.tenant, f.filter = tenantID, filter
	return 0, nil
}

func TestAdminRuntimeServicePassesTenantScopedCommands(t *testing.T) {
	now := time.Date(2026, 8, 29, 10, 0, 0, 0, time.UTC)
	runtime := &adminRuntimeFake{instance: &entities.WorkflowInstance{ID: "instance-1", TenantID: "tenant-a", WorkflowID: "workflow-1", WorkflowVersionID: "version-1", Status: entities.WorkflowInstanceFailed, Version: 4, CreatedAt: now, UpdatedAt: now}}
	auditRepo := &adminAuditRepoFake{}
	service := NewAdminRuntimeService(runtime, &adminEventBrowserFake{}, NewAuditWriter(auditRepo, nil, nil))

	correlationID := stringPtr("correlation-1")
	result, err := service.Retry(context.Background(), "tenant-a", "actor", "instance-1", correlationID)
	if err != nil {
		t.Fatalf("retry instance: %v", err)
	}
	if result.Status != "FAILED" || runtime.called != "retry:tenant-a:instance-1" {
		t.Fatalf("unexpected retry result/call: %#v %q", result, runtime.called)
	}
	if len(auditRepo.entries) != 1 || auditRepo.entries[0].Action != entities.AuditActionWorkflowRetried || auditRepo.entries[0].CorrelationID == nil || *auditRepo.entries[0].CorrelationID != *correlationID {
		t.Fatalf("expected correlated retry audit entry, got %#v", auditRepo.entries)
	}
}

func TestAdminRuntimeServiceFiltersEventsAndUsesSafePagination(t *testing.T) {
	events := &adminEventBrowserFake{}
	service := NewAdminRuntimeService(&adminRuntimeFake{instance: &entities.WorkflowInstance{}}, events, nil)

	page, err := service.ListEvents(context.Background(), "tenant-a", EventBrowserQuery{
		WorkflowInstanceID: "instance-1", Type: "state.entered", Source: "engine", CorrelationID: "corr-1", Page: 0, PageSize: 500,
	})
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if page.Page != 1 || page.PageSize != 100 || events.tenant != "tenant-a" {
		t.Fatalf("unexpected page or tenant: %#v tenant=%q", page, events.tenant)
	}
	if events.filter.WorkflowInstanceID == nil || *events.filter.WorkflowInstanceID != "instance-1" || events.filter.Limit != 100 {
		t.Fatalf("unexpected event filter: %#v", events.filter)
	}
}
