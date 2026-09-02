package capability

import (
	"context"
	"strings"
	"testing"

	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
)

func TestMemorySecretStoreLifecycleUsesOpaqueScopedReferences(t *testing.T) {
	store := NewMemorySecretStore()
	ctx := context.Background()
	ref, err := store.Put(ctx, "tenant-a", "project-a", "mcp_bearer", "super-secret-token")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(ref, "super-secret-token") || !strings.HasPrefix(ref, "secret://memory/") {
		t.Fatalf("reference leaked or has wrong format: %q", ref)
	}
	if _, err := store.Resolve(ctx, "tenant-b", "project-a", ref); err == nil {
		t.Fatal("cross-tenant secret resolution unexpectedly succeeded")
	}
	rotated, err := store.Rotate(ctx, "tenant-a", "project-a", ref, "mcp_bearer", "replacement-token")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Resolve(ctx, "tenant-a", "project-a", ref); err == nil {
		t.Fatal("old secret remained usable after rotation")
	}
	if got, err := store.Resolve(ctx, "tenant-a", "project-a", rotated); err != nil || got != "replacement-token" {
		t.Fatalf("rotated secret = %q, err = %v", got, err)
	}
	status, err := store.Status(ctx, "tenant-a", "project-a", ref)
	if err != nil || status != domainsvc.SecretStatusRevoked {
		t.Fatalf("old status = %q, err = %v", status, err)
	}
}

func TestCompositeSecretStoreKeepsEnvironmentReferencesReadOnly(t *testing.T) {
	store := NewCompositeSecretStore("TEST_CRED_")
	if _, err := store.Put(context.Background(), "tenant", "project", "mcp_bearer", "token"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Rotate(context.Background(), "tenant", "project", "ref:vault/provider", "mcp_bearer", "token"); err != nil {
		t.Fatal(err)
	}
}
