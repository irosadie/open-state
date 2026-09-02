package services

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
)

func TestMCPBearerCredentialRotateAndRevokeAreSafe(t *testing.T) {
	repo := &oauthConnectionRepository{fakeMCPConnectionRepository: &fakeMCPConnectionRepository{item: &entities.MCPConnection{ID: testID, TenantID: testTenant, ProjectID: testProject, Name: "Provider", Alias: "provider", AuthType: entities.MCPAuthBearer, CredentialStatus: entities.MCPCredentialMissing, Status: entities.MCPConnectionEnabled}}}
	secrets := capabilityTestSecretStore()
	service := NewMCPConnectionService(repo, fakeProjectRepository{}, nil, nil, secrets)
	rotated, err := service.RotateCredential(context.Background(), testTenant, testProject, testID, "actor", "bearer-secret")
	if err != nil {
		t.Fatal(err)
	}
	if rotated.CredentialStatus != string(entities.MCPCredentialConfigured) {
		t.Fatalf("credential status = %q", rotated.CredentialStatus)
	}
	encoded, _ := json.Marshal(rotated)
	if strings.Contains(string(encoded), "bearer-secret") || strings.Contains(string(encoded), "secret://") {
		t.Fatal("bearer credential leaked in response")
	}
	if _, err := service.RevokeCredential(context.Background(), testTenant, testProject, testID, "actor"); err != nil {
		t.Fatal(err)
	}
	if repo.item.CredentialReference != nil || repo.item.CredentialStatus != entities.MCPCredentialMissing {
		t.Fatal("bearer credential was not revoked")
	}
}

func TestMCPConnectionWriteDTOKeepsCredentialValuesWriteOnly(t *testing.T) {
	request := dtos.CreateMCPConnectionRequest{CredentialValue: "secret-value", OAuthClientSecretValue: "oauth-secret"}
	encoded, _ := json.Marshal(request)
	if !strings.Contains(string(encoded), "secret-value") {
		t.Fatal("write request should contain value before transport")
	}
	// The response DTO is intentionally a different type and has no write-only fields.
	response := dtos.MCPConnectionDTO{ID: testID, CredentialStatus: "configured"}
	encoded, _ = json.Marshal(response)
	if strings.Contains(string(encoded), "CredentialValue") || strings.Contains(string(encoded), "credentialReference") {
		t.Fatal("response DTO contains credential fields")
	}
}
