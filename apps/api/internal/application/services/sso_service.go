package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// SSOService orchestrates the external OIDC login flow (PRD §79): building the
// authorization URL (PKCE + state) and completing the callback by resolving or
// auto-provisioning a local user linked to the external identity.
type SSOService struct {
	identities repositories.IUserIdentityRepository
	auth       repositories.IAuthRepository
	roles      repositories.IRoleAssignmentRepository
	token      domainsvc.TokenService
	providers  map[string]domainsvc.OIDCProvider
	audit      *AuditWriter
}

// NewSSOService builds an SSOService over the enabled providers.
func NewSSOService(
	identities repositories.IUserIdentityRepository,
	auth repositories.IAuthRepository,
	roles repositories.IRoleAssignmentRepository,
	token domainsvc.TokenService,
	providers map[string]domainsvc.OIDCProvider,
	audit *AuditWriter,
) *SSOService {
	return &SSOService{
		identities: identities,
		auth:       auth,
		roles:      roles,
		token:      token,
		providers:  providers,
		audit:      audit,
	}
}

// EnabledProviders returns the identifiers of configured providers.
func (s *SSOService) EnabledProviders() []string {
	out := make([]string, 0, len(s.providers))
	for name := range s.providers {
		out = append(out, name)
	}
	return out
}

// StartAuth builds the authorization URL for a provider with a fresh state and
// PKCE code verifier. The caller persists state + verifier (e.g. in a cookie).
func (s *SSOService) StartAuth(provider string) (authURL, state, codeVerifier string, err error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", "", "", domain.NewNotFound("sso provider not enabled")
	}
	state, err = randomString(32)
	if err != nil {
		return "", "", "", domain.NewInternal("failed to generate state")
	}
	codeVerifier, err = randomString(48)
	if err != nil {
		return "", "", "", domain.NewInternal("failed to generate PKCE verifier")
	}
	return p.AuthURL(state, codeVerifier), state, codeVerifier, nil
}

// CompleteLogin handles the provider callback: verifies the state, exchanges the
// code with PKCE, resolves or auto-provisions the user, and returns a login
// session (PRD §79). codeVerifier comes from the persisted PKCE state.
func (s *SSOService) CompleteLogin(ctx context.Context, provider, code, codeVerifier string) (*dtos.LoginDTO, error) {
	p, ok := s.providers[provider]
	if !ok {
		return nil, domain.NewNotFound("sso provider not enabled")
	}

	identity, err := p.Exchange(ctx, code, codeVerifier)
	if err != nil {
		return nil, domain.NewUnauthorized("sso authentication failed")
	}
	if identity.Email == "" {
		return nil, domain.NewUnauthorized("sso provider returned no email")
	}

	user, err := s.resolveOrProvision(ctx, identity)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.token.GenerateToken(user.ID)
	if err != nil {
		return nil, domain.NewInternal("failed to issue token")
	}

	if s.audit != nil {
		s.audit.Write(ctx, user.ID, user.ID, entities.AuditActionSSOLogin, "auth", user.ID, nil, nil, nil)
	}

	return &dtos.LoginDTO{
		AccessToken: accessToken,
		User:        *toUserDTO(user),
	}, nil
}

// resolveOrProvision returns the local user linked to an external identity,
// creating both when the identity is seen for the first time (auto-provision,
// PRD §79). New users default to the least-privilege VIEWER role.
func (s *SSOService) resolveOrProvision(ctx context.Context, identity *domainsvc.OIDCIdentity) (*entities.User, error) {
	existing, err := s.identities.FindByProviderSubject(ctx, identity.Provider, identity.Subject)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return s.auth.FindUserByID(ctx, existing.UserID)
	}

	// Auto-provision: create user + default tenant role + identity link.
	user, err := s.auth.CreateUser(ctx, identity.Email, "", identity.Name, entities.UserRoleLegacy, entities.UserStatusActive)
	if err != nil {
		return nil, domain.NewInternal("failed to provision user")
	}

	// Default tenant role (VIEWER, least privilege). Uses the fixed demo tenant
	// for now; full tenant resolution is out of scope for this slice.
	if _, err := s.roles.Assign(ctx, user.ID, demoTenantID, entities.UserRoleViewer); err != nil {
		return nil, domain.NewInternal("failed to assign default role")
	}

	if _, err := s.identities.Create(ctx, user.ID, identity.Provider, identity.Subject, true); err != nil {
		return nil, domain.NewInternal("failed to link identity")
	}

	return user, nil
}

// demoTenantID is the default tenant for auto-provisioned SSO users until
// per-user tenant selection lands.
const demoTenantID = "00000000-0000-0000-0000-000000000001"

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
