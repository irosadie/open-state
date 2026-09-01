package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

const (
	testTenant  = "00000000-0000-0000-0000-000000000001"
	testProject = "00000000-0000-0000-0000-000000000002"
	testID      = "00000000-0000-0000-0000-000000000003"
)

type fakeMCPConnectionRepository struct {
	item       *entities.MCPConnection
	findErr    error
	createErr  error
	updateErr  error
	testCalled bool
}

func (f *fakeMCPConnectionRepository) Create(_ context.Context, input repositories.MCPConnectionCreateInput) (*entities.MCPConnection, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.item = &entities.MCPConnection{ID: testID, TenantID: input.TenantID, ProjectID: input.ProjectID, Name: input.Name, Alias: input.Alias, Transport: input.Transport, Endpoint: input.Endpoint, AuthType: input.AuthType, CredentialReference: input.CredentialReference, CredentialStatus: input.CredentialStatus, Status: input.Status, LastTestStatus: entities.MCPTestNever}
	return f.item, nil
}
func (f *fakeMCPConnectionRepository) FindByID(_ context.Context, _, _, _ string) (*entities.MCPConnection, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.item, nil
}
func (f *fakeMCPConnectionRepository) ListByProject(context.Context, string, string) ([]entities.MCPConnection, error) {
	if f.item == nil {
		return []entities.MCPConnection{}, nil
	}
	return []entities.MCPConnection{*f.item}, nil
}
func (f *fakeMCPConnectionRepository) Update(_ context.Context, input repositories.MCPConnectionUpdateInput) (*entities.MCPConnection, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	f.item.Name = input.Name
	f.item.AuthType = input.AuthType
	return f.item, nil
}
func (f *fakeMCPConnectionRepository) Delete(context.Context, string, string, string) error {
	return nil
}
func (f *fakeMCPConnectionRepository) UpdateStatus(context.Context, string, string, string, entities.MCPConnectionStatus, string) (*entities.MCPConnection, error) {
	return f.item, nil
}
func (f *fakeMCPConnectionRepository) RecordTest(_ context.Context, _, _, _ string, status entities.MCPConnectionTestStatus, errorCode, _ string) (*entities.MCPConnection, error) {
	f.testCalled = true
	f.item.LastTestStatus = status
	f.item.LastTestErrorCode = optionalString(errorCode)
	return f.item, nil
}

type fakeProjectRepository struct{}

func (fakeProjectRepository) Create(context.Context, string, string, string, entities.ProjectStatus) (*entities.Project, error) {
	return nil, nil
}
func (fakeProjectRepository) FindByID(context.Context, string, string) (*entities.Project, error) {
	return &entities.Project{ID: testProject, TenantID: testTenant}, nil
}
func (fakeProjectRepository) FindBySlug(context.Context, string, string) (*entities.Project, error) {
	return nil, nil
}
func (fakeProjectRepository) ListByTenant(context.Context, string) ([]entities.Project, error) {
	return nil, nil
}

type fakeMCPTester struct{ called bool }

func (f *fakeMCPTester) Handshake(context.Context, *entities.MCPConnection) (domainsvc.MCPHandshakeResult, error) {
	f.called = true
	return domainsvc.MCPHandshakeResult{Ready: true}, nil
}

func TestMCPConnectionServiceCreateReturnsSafeMetadata(t *testing.T) {
	repo := &fakeMCPConnectionRepository{}
	service := NewMCPConnectionService(repo, fakeProjectRepository{}, &fakeMCPTester{}, nil)
	result, err := service.Create(context.Background(), testTenant, testProject, "actor", dtos.CreateMCPConnectionRequest{Name: "Padel", Alias: "padel", Transport: "streamable_http", Endpoint: "http://localhost:8031/mcp", AuthType: "bearer", CredentialReference: "ref:vault/padel"})
	if err != nil {
		t.Fatal(err)
	}
	if result.CredentialStatus != "configured" {
		t.Fatalf("credential status = %q", result.CredentialStatus)
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "credentialReference") || strings.Contains(string(encoded), "ref:vault") {
		t.Fatal("safe DTO exposed credential reference")
	}
}

func TestMCPConnectionServiceRejectsInvalidTransportConfiguration(t *testing.T) {
	service := NewMCPConnectionService(&fakeMCPConnectionRepository{}, fakeProjectRepository{}, &fakeMCPTester{}, nil)
	_, err := service.Create(context.Background(), testTenant, testProject, "actor", dtos.CreateMCPConnectionRequest{Name: "Padel", Alias: "padel", Transport: "sse", AuthType: "none"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var de *domain.DomainError
	if !errors.As(err, &de) || de.Code != domain.ErrValidation {
		t.Fatalf("error = %#v", err)
	}
}

func TestMCPConnectionServiceDoesNotTestDisabledConnection(t *testing.T) {
	repo := &fakeMCPConnectionRepository{item: &entities.MCPConnection{ID: testID, TenantID: testTenant, ProjectID: testProject, Status: entities.MCPConnectionDisabled}}
	tester := &fakeMCPTester{}
	service := NewMCPConnectionService(repo, fakeProjectRepository{}, tester, nil)
	_, err := service.Test(context.Background(), testTenant, testProject, testID, "actor")
	if err == nil {
		t.Fatal("expected disabled connection error")
	}
	if tester.called || repo.testCalled {
		t.Fatal("disabled connection contacted or recorded a test")
	}
}
