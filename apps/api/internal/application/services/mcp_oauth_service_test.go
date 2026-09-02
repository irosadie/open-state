package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	capinfra "github.com/irosadie/open-state/api/internal/infrastructure/capability"
)

type oauthConnectionRepository struct {
	*fakeMCPConnectionRepository
	lastOAuth repositories.MCPConnectionOAuthUpdateInput
}

func (r *oauthConnectionRepository) UpdateCredential(_ context.Context, input repositories.MCPConnectionCredentialUpdateInput) (*entities.MCPConnection, error) {
	r.item.CredentialReference = input.CredentialReference
	r.item.CredentialStatus = input.CredentialStatus
	return r.item, nil
}

func (r *oauthConnectionRepository) UpdateOAuth(_ context.Context, input repositories.MCPConnectionOAuthUpdateInput) (*entities.MCPConnection, error) {
	r.lastOAuth = input
	r.item.CredentialReference = input.CredentialReference
	r.item.OAuthAccessTokenReference = input.CredentialReference
	r.item.OAuthRefreshTokenReference = input.RefreshTokenReference
	r.item.CredentialStatus = input.CredentialStatus
	r.item.OAuthStatus = input.OAuthStatus
	r.item.OAuthExpiresAt = input.OAuthExpiresAt
	return r.item, nil
}

func (r *oauthConnectionRepository) DisconnectOAuth(context.Context, string, string, string, string) (*entities.MCPConnection, error) {
	r.item.CredentialReference = nil
	r.item.OAuthAccessTokenReference = nil
	r.item.OAuthRefreshTokenReference = nil
	r.item.OAuthStatus = entities.MCPOAuthDisconnected
	r.item.CredentialStatus = entities.MCPCredentialActionRequired
	return r.item, nil
}

func (r *oauthConnectionRepository) RecordHealth(context.Context, repositories.MCPConnectionHealthUpdateInput) (*entities.MCPConnection, error) {
	return r.item, nil
}

func (r *oauthConnectionRepository) ResetHealth(context.Context, string, string, string, string) (*entities.MCPConnection, error) {
	return r.item, nil
}

type oauthTransactionRepository struct {
	transaction *entities.MCPAuthorizationTransaction
}

func (r *oauthTransactionRepository) Create(_ context.Context, input repositories.MCPAuthorizationTransactionCreateInput) (*entities.MCPAuthorizationTransaction, error) {
	r.transaction = &entities.MCPAuthorizationTransaction{ID: "tx-1", TenantID: input.TenantID, ProjectID: input.ProjectID, ConnectionID: input.ConnectionID, StateHash: input.StateHash, VerifierReference: input.VerifierReference, RedirectURI: input.RedirectURI, ExpiresAt: input.ExpiresAt, Status: entities.MCPAuthorizationPending}
	return r.transaction, nil
}

func (r *oauthTransactionRepository) FindPendingByState(_ context.Context, _, _, _ string, hash []byte) (*entities.MCPAuthorizationTransaction, error) {
	if r.transaction == nil || r.transaction.Status != entities.MCPAuthorizationPending || string(hash) != string(r.transaction.StateHash) {
		return nil, errors.New("transaction not found")
	}
	return r.transaction, nil
}

func (r *oauthTransactionRepository) Consume(context.Context, string, string, string) error {
	if r.transaction == nil {
		return errors.New("transaction not found")
	}
	r.transaction.Status = entities.MCPAuthorizationConsumed
	return nil
}

type oauthClientFake struct {
	exchanged bool
	refreshed bool
	err       error
}

func (c *oauthClientFake) Exchange(context.Context, string, string, string, string, string, string) (domainsvc.OAuthTokenResponse, error) {
	c.exchanged = true
	if c.err != nil {
		return domainsvc.OAuthTokenResponse{}, c.err
	}
	return domainsvc.OAuthTokenResponse{AccessToken: "access-secret", RefreshToken: "refresh-secret", ExpiresIn: time.Hour}, nil
}

func (c *oauthClientFake) Refresh(context.Context, string, string, string, string) (domainsvc.OAuthTokenResponse, error) {
	c.refreshed = true
	if c.err != nil {
		return domainsvc.OAuthTokenResponse{}, c.err
	}
	return domainsvc.OAuthTokenResponse{AccessToken: "refreshed-secret", ExpiresIn: time.Hour}, nil
}

func oauthConnection() *entities.MCPConnection {
	return &entities.MCPConnection{ID: testID, TenantID: testTenant, ProjectID: testProject, Name: "Provider", Alias: "provider", AuthType: entities.MCPAuthOAuth, OAuthAuthorizationEndpoint: stringPointer("https://provider.example/authorize"), OAuthTokenEndpoint: stringPointer("https://provider.example/token"), OAuthClientID: stringPointer("client-id"), OAuthRedirectURI: stringPointer("https://app.example/callback"), OAuthStatus: entities.MCPOAuthDisconnected, CredentialStatus: entities.MCPCredentialMissing, Status: entities.MCPConnectionEnabled, CreatedAt: time.Now(), UpdatedAt: time.Now()}
}

func stringPointer(value string) *string { return &value }

func capabilityTestSecretStore() domainsvc.SecretStore { return capinfra.NewMemorySecretStore() }

func newOAuthService(repo *oauthConnectionRepository, transactions *oauthTransactionRepository, secrets domainsvc.SecretStore, client domainsvc.OAuthClient) *MCPOAuthService {
	return NewMCPOAuthService(repo, transactions, secrets, client, nil)
}

func TestMCPOAuthLifecycleProtectsStateAndTokenArtifacts(t *testing.T) {
	repo := &oauthConnectionRepository{fakeMCPConnectionRepository: &fakeMCPConnectionRepository{item: oauthConnection()}}
	transactions := &oauthTransactionRepository{}
	secrets := capabilityTestSecretStore()
	client := &oauthClientFake{}
	service := newOAuthService(repo, transactions, secrets, client)
	start, err := service.Start(context.Background(), testTenant, testProject, testID, "actor")
	if err != nil {
		t.Fatal(err)
	}
	authorization, err := url.Parse(start.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	state := authorization.Query().Get("state")
	if state == "" || authorization.Query().Get("code_challenge") == "" {
		t.Fatal("OAuth state or PKCE challenge missing")
	}
	result, err := service.Callback(context.Background(), testTenant, testProject, testID, "actor", state, "authorization-code")
	if err != nil {
		t.Fatal(err)
	}
	if !client.exchanged || transactions.transaction.Status != entities.MCPAuthorizationConsumed || repo.item.OAuthStatus != entities.MCPOAuthConnected {
		t.Fatal("OAuth lifecycle did not complete")
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "access-secret") || strings.Contains(string(encoded), "secret://") {
		t.Fatal("OAuth token artifact leaked in response")
	}
	if _, err := service.Callback(context.Background(), testTenant, testProject, testID, "actor", state, "authorization-code"); err == nil {
		t.Fatal("OAuth state was reusable")
	}
}

func TestMCPOAuthRefreshFailureBecomesActionRequired(t *testing.T) {
	secrets := capabilityTestSecretStore()
	refreshRef, err := secrets.Put(context.Background(), testTenant, testProject, "refresh", "refresh-secret")
	if err != nil {
		t.Fatal(err)
	}
	expired := time.Now().UTC().Add(-time.Hour)
	connection := oauthConnection()
	connection.OAuthRefreshTokenReference = &refreshRef
	connection.OAuthExpiresAt = &expired
	connection.OAuthStatus = entities.MCPOAuthExpired
	repo := &oauthConnectionRepository{fakeMCPConnectionRepository: &fakeMCPConnectionRepository{item: connection}}
	client := &oauthClientFake{err: errors.New("provider rejected refresh")}
	service := newOAuthService(repo, &oauthTransactionRepository{}, secrets, client)
	if _, err := service.AccessToken(context.Background(), testID, testTenant, testProject); err == nil {
		t.Fatal("refresh failure was hidden")
	}
	if repo.lastOAuth.OAuthStatus != entities.MCPOAuthActionRequired || repo.lastOAuth.CredentialStatus != entities.MCPCredentialActionRequired {
		t.Fatal("refresh failure did not require operator action")
	}
}

func TestMCPOAuthDisconnectRemovesReferencesWithoutReturningSecrets(t *testing.T) {
	secrets := capabilityTestSecretStore()
	accessRef, _ := secrets.Put(context.Background(), testTenant, testProject, "access", "access-secret")
	connection := oauthConnection()
	connection.OAuthAccessTokenReference = &accessRef
	connection.CredentialReference = &accessRef
	connection.OAuthStatus = entities.MCPOAuthConnected
	repo := &oauthConnectionRepository{fakeMCPConnectionRepository: &fakeMCPConnectionRepository{item: connection}}
	service := newOAuthService(repo, &oauthTransactionRepository{}, secrets, &oauthClientFake{})
	result, err := service.Disconnect(context.Background(), testTenant, testProject, testID, "actor")
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "access-secret") || repo.item.OAuthAccessTokenReference != nil {
		t.Fatal("OAuth disconnect retained or exposed token material")
	}
}
