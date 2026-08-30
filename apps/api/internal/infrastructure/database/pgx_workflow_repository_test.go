package database

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/infrastructure/config"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// This integration test is opt-in because repository tests require a migrated
// PostgreSQL instance. It exercises the durable draft column against the real
// sqlc/pgx adapter when DATABASE_URL is available.
func TestPgxWorkflowRepositoryDraftLifecycle(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping PostgreSQL workflow repository test")
	}

	ctx := context.Background()
	pool, err := config.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	tenantID := uuid.NewString()
	projectRepo := NewPgxProjectRepository(pool)
	project, err := projectRepo.Create(ctx, tenantID, "Draft test", "draft-test-"+uuid.NewString(), entities.ProjectActive)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM projects WHERE id = $1", project.ID)
	}()

	workflowRepo := NewPgxWorkflowRepository(pool)
	initial := []byte(`{"nodes":[{"id":"start","kind":"START"}],"transitions":[]}`)
	wf, err := workflowRepo.Create(ctx, tenantID, project.ID, "draft-"+uuid.NewString(), "Draft", nil, initial)
	if err != nil {
		t.Fatalf("create workflow: %v", err)
	}

	read, err := workflowRepo.FindByID(ctx, tenantID, project.ID, wf.ID)
	if err != nil || !bytes.Equal(read.DraftDefinition, initial) {
		t.Fatalf("draft round-trip failed: workflow=%+v err=%v", read, err)
	}
	if _, err := workflowRepo.FindByID(ctx, uuid.NewString(), project.ID, wf.ID); err == nil {
		t.Fatal("expected tenant-scoped read to reject another tenant")
	}

	updatedDefinition := []byte(`{"nodes":[{"id":"start","kind":"START"},{"id":"end","kind":"END"}],"transitions":[]}`)
	updated, err := workflowRepo.UpdateDraft(ctx, tenantID, project.ID, wf.ID, "Draft", nil, updatedDefinition, wf.Version)
	if err != nil {
		t.Fatalf("update workflow draft: %v", err)
	}
	if updated.Version != wf.Version+1 || !bytes.Equal(updated.DraftDefinition, updatedDefinition) {
		t.Fatalf("unexpected updated draft: %+v", updated)
	}
	if _, err := workflowRepo.UpdateDraft(ctx, tenantID, project.ID, wf.ID, "stale", nil, initial, wf.Version); err == nil {
		t.Fatal("expected optimistic conflict")
	} else {
		var domainErr *domain.DomainError
		if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrConflict {
			t.Fatalf("expected conflict error, got %v", err)
		}
	}

	if _, err := workflowRepo.Publish(ctx, tenantID, project.ID, wf.ID, 1, updatedDefinition, entities.VersionStatusPublished, updated.Version); err != nil {
		t.Fatalf("publish workflow: %v", err)
	}
	root, err := workflowRepo.FindByID(ctx, tenantID, project.ID, wf.ID)
	if err != nil {
		t.Fatalf("read published root: %v", err)
	}
	if _, err := workflowRepo.UpdateDraft(ctx, tenantID, project.ID, wf.ID, "Draft", nil, initial, root.Version); err != nil {
		t.Fatalf("update after publish: %v", err)
	}
	version, err := workflowRepo.FindVersionByNumber(ctx, tenantID, project.ID, wf.ID, 1)
	if err != nil || !bytes.Equal(version.Definition, updatedDefinition) {
		t.Fatalf("published snapshot was not preserved: version=%+v err=%v", version, err)
	}
}
