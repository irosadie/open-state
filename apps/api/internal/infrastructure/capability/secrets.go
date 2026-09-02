package capability

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
)

// CredentialResolver resolves a credential_reference to the actual credential
// from secure infrastructure (env / secret manager / Vault) (PRD §61).
// It never stores secrets in workflow definitions and never returns them for
// logging (PRD §91).
type CredentialResolver interface {
	Resolve(ref string) (string, bool)
}

// EnvCredentialResolver resolves credential references from environment
// variables using an optional prefix (e.g. "CRED_").
type EnvCredentialResolver struct {
	Prefix string
}

// Resolve implements CredentialResolver. It returns the env value for the
// variable named by prefix+ref (uppercased, non-alphanumeric → underscore).
func (r EnvCredentialResolver) Resolve(ref string) (string, bool) {
	if ref == "" {
		return "", false
	}
	key := r.Prefix + sanitizeEnvKey(ref)
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return "", false
	}
	return v, true
}

// sanitizeEnvKey converts a credential reference to a safe env var name.
func sanitizeEnvKey(ref string) string {
	var b strings.Builder
	for _, r := range ref {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return strings.ToUpper(b.String())
}

// MemorySecretStore is intentionally a development/test adapter. It provides
// the complete lifecycle without ever serializing a secret into application
// records. Production deployments should bind SecretStore to a vault/KMS
// implementation instead.
type MemorySecretStore struct {
	mu     sync.RWMutex
	values map[string]memorySecret
}

type memorySecret struct {
	tenantID  string
	projectID string
	kind      string
	value     string
	revoked   bool
}

func NewMemorySecretStore() *MemorySecretStore {
	return &MemorySecretStore{values: make(map[string]memorySecret)}
}

func (s *MemorySecretStore) Put(_ context.Context, tenantID, projectID, kind, value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", errors.New("secret value is required")
	}
	ref := "secret://memory/" + uuid.NewString()
	s.mu.Lock()
	s.values[ref] = memorySecret{tenantID: tenantID, projectID: projectID, kind: kind, value: value}
	s.mu.Unlock()
	return ref, nil
}

func (s *MemorySecretStore) Resolve(_ context.Context, tenantID, projectID, reference string) (string, error) {
	s.mu.RLock()
	secret, ok := s.values[reference]
	s.mu.RUnlock()
	if !ok || secret.tenantID != tenantID || secret.projectID != projectID || secret.revoked || secret.value == "" {
		return "", errors.New("secret unavailable")
	}
	return secret.value, nil
}

func (s *MemorySecretStore) Rotate(ctx context.Context, tenantID, projectID, reference, kind, value string) (string, error) {
	if _, err := s.Resolve(ctx, tenantID, projectID, reference); err != nil {
		return "", err
	}
	newRef, err := s.Put(ctx, tenantID, projectID, kind, value)
	if err != nil {
		return "", err
	}
	if err := s.Revoke(ctx, tenantID, projectID, reference); err != nil {
		return "", err
	}
	return newRef, nil
}

func (s *MemorySecretStore) Revoke(_ context.Context, tenantID, projectID, reference string) error {
	s.mu.Lock()
	secret, ok := s.values[reference]
	if !ok || secret.tenantID != tenantID || secret.projectID != projectID {
		s.mu.Unlock()
		return errors.New("secret unavailable")
	}
	secret.value = ""
	secret.revoked = true
	s.values[reference] = secret
	s.mu.Unlock()
	return nil
}

func (s *MemorySecretStore) Status(_ context.Context, tenantID, projectID, reference string) (domainsvc.SecretStatus, error) {
	s.mu.RLock()
	secret, ok := s.values[reference]
	s.mu.RUnlock()
	if !ok || secret.tenantID != tenantID || secret.projectID != projectID {
		return domainsvc.SecretStatusMissing, nil
	}
	if secret.revoked || secret.value == "" {
		return domainsvc.SecretStatusRevoked, nil
	}
	return domainsvc.SecretStatusConfigured, nil
}

// EnvironmentSecretStore resolves deployment-provisioned references. Writes
// are deliberately rejected: environment-backed secrets must be rotated by
// the deployment secret manager, not by the HTTP API.
type EnvironmentSecretStore struct {
	Resolver EnvCredentialResolver
}

func (s EnvironmentSecretStore) Put(context.Context, string, string, string, string) (string, error) {
	return "", errors.New("environment secret store is read-only")
}

func (s EnvironmentSecretStore) Resolve(_ context.Context, _, _, reference string) (string, error) {
	value, ok := s.Resolver.Resolve(reference)
	if !ok {
		return "", errors.New("secret unavailable")
	}
	return value, nil
}

func (s EnvironmentSecretStore) Rotate(context.Context, string, string, string, string, string) (string, error) {
	return "", errors.New("environment secret store is read-only")
}

func (s EnvironmentSecretStore) Revoke(context.Context, string, string, string) error {
	return errors.New("environment secret store is read-only")
}

func (s EnvironmentSecretStore) Status(ctx context.Context, tenantID, projectID, reference string) (domainsvc.SecretStatus, error) {
	if _, err := s.Resolve(ctx, tenantID, projectID, reference); err != nil {
		return domainsvc.SecretStatusMissing, nil
	}
	return domainsvc.SecretStatusConfigured, nil
}

// CompositeSecretStore supports local rotation while keeping existing
// environment-backed references compatible during migration.
type CompositeSecretStore struct {
	Memory *MemorySecretStore
	Env    EnvironmentSecretStore
}

func NewCompositeSecretStore(prefix string) *CompositeSecretStore {
	return &CompositeSecretStore{Memory: NewMemorySecretStore(), Env: EnvironmentSecretStore{Resolver: EnvCredentialResolver{Prefix: prefix}}}
}

func (s *CompositeSecretStore) Put(ctx context.Context, tenantID, projectID, kind, value string) (string, error) {
	return s.Memory.Put(ctx, tenantID, projectID, kind, value)
}

func (s *CompositeSecretStore) Resolve(ctx context.Context, tenantID, projectID, reference string) (string, error) {
	if strings.HasPrefix(reference, "secret://memory/") {
		return s.Memory.Resolve(ctx, tenantID, projectID, reference)
	}
	return s.Env.Resolve(ctx, tenantID, projectID, reference)
}

func (s *CompositeSecretStore) Rotate(ctx context.Context, tenantID, projectID, reference, kind, value string) (string, error) {
	if strings.HasPrefix(reference, "secret://memory/") {
		return s.Memory.Rotate(ctx, tenantID, projectID, reference, kind, value)
	}
	return s.Memory.Put(ctx, tenantID, projectID, kind, value)
}

func (s *CompositeSecretStore) Revoke(ctx context.Context, tenantID, projectID, reference string) error {
	if strings.HasPrefix(reference, "secret://memory/") {
		return s.Memory.Revoke(ctx, tenantID, projectID, reference)
	}
	return s.Env.Revoke(ctx, tenantID, projectID, reference)
}

func (s *CompositeSecretStore) Status(ctx context.Context, tenantID, projectID, reference string) (domainsvc.SecretStatus, error) {
	if strings.HasPrefix(reference, "secret://memory/") {
		return s.Memory.Status(ctx, tenantID, projectID, reference)
	}
	return s.Env.Status(ctx, tenantID, projectID, reference)
}

// SecretBackend is the production integration seam. The deployment can wrap a
// Vault/KMS SDK without pulling that vendor dependency into the application.
type SecretBackend interface {
	Put(ctx context.Context, tenantID, projectID, kind, value string) (string, error)
	Resolve(ctx context.Context, tenantID, projectID, reference string) (string, error)
	Rotate(ctx context.Context, tenantID, projectID, reference, kind, value string) (string, error)
	Revoke(ctx context.Context, tenantID, projectID, reference string) error
	Status(ctx context.Context, tenantID, projectID, reference string) (domainsvc.SecretStatus, error)
}

// ProductionSecretStore is the adapter used by production composition roots.
// It intentionally has no fallback to environment variables or memory.
type ProductionSecretStore struct{ Backend SecretBackend }

func (s ProductionSecretStore) Put(ctx context.Context, tenantID, projectID, kind, value string) (string, error) {
	if s.Backend == nil {
		return "", errors.New("production secret backend is not configured")
	}
	return s.Backend.Put(ctx, tenantID, projectID, kind, value)
}

func (s ProductionSecretStore) Resolve(ctx context.Context, tenantID, projectID, reference string) (string, error) {
	if s.Backend == nil {
		return "", errors.New("production secret backend is not configured")
	}
	return s.Backend.Resolve(ctx, tenantID, projectID, reference)
}

func (s ProductionSecretStore) Rotate(ctx context.Context, tenantID, projectID, reference, kind, value string) (string, error) {
	if s.Backend == nil {
		return "", errors.New("production secret backend is not configured")
	}
	return s.Backend.Rotate(ctx, tenantID, projectID, reference, kind, value)
}

func (s ProductionSecretStore) Revoke(ctx context.Context, tenantID, projectID, reference string) error {
	if s.Backend == nil {
		return errors.New("production secret backend is not configured")
	}
	return s.Backend.Revoke(ctx, tenantID, projectID, reference)
}

func (s ProductionSecretStore) Status(ctx context.Context, tenantID, projectID, reference string) (domainsvc.SecretStatus, error) {
	if s.Backend == nil {
		return domainsvc.SecretStatusMissing, errors.New("production secret backend is not configured")
	}
	return s.Backend.Status(ctx, tenantID, projectID, reference)
}
