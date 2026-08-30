// Command seed upserts the demo example workflows (padel-court-booking,
// order-food, order-doctor) and their projects under a fixed demo tenant, so
// they can be resolved and executed end-to-end (PRD §40.1, epic #7).
//
// The seed is idempotent: re-running it does not duplicate rows (upsert by
// project slug / workflow slug). Demo scope is a dedicated tenant, so seeded
// data never pollutes other tenants (PRD §4).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/config"
	infradb "github.com/irosadie/open-state/api/internal/infrastructure/database"
	"github.com/irosadie/open-state/go-shared/domain"
)

// Demo tenant/project identifiers. A fixed tenant UUID isolates seeded data.
const demoTenantID = "00000000-0000-0000-0000-000000000001"

// seedProject maps a demo project to its canonical workflow JSON file and the
// intent id that resolves to it (PRD §40.1).
type seedProject struct {
	slug         string
	name         string
	workflowFile string
	intentID     string
}

var seeds = []seedProject{
	{slug: "padel", name: "Padel", workflowFile: "padel-booking.workflow.json", intentID: "BOOKING_PADEL"},
	{slug: "retail", name: "Retail", workflowFile: "order-food.workflow.json", intentID: "ORDER_FOOD"},
	{slug: "health", name: "Health", workflowFile: "order-doctor.workflow.json", intentID: "ORDER_DOCTOR"},
}

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err.Error())
		return
	}
	slog.Info("seed complete")
}

func run() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := config.NewPool(ctx, dbURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	adapter := infradb.NewPostgresAdapter(pool)
	for _, s := range seeds {
		if err := seedOne(ctx, adapter.Projects(), adapter.Workflows(), s); err != nil {
			return err
		}
	}
	return nil
}

// seedOne upserts a project and its workflow definition + intent mapping.
func seedOne(ctx context.Context, projects repositories.IProjectRepository, workflows repositories.IWorkflowRepository, s seedProject) error {
	proj, err := projects.FindBySlug(ctx, demoTenantID, s.slug)
	if err != nil {
		if !isNotFound(err) {
			return fmt.Errorf("find project %s: %w", s.slug, err)
		}
		proj, err = projects.Create(ctx, demoTenantID, s.name, s.slug, entities.ProjectActive)
		if err != nil {
			return fmt.Errorf("create project %s: %w", s.slug, err)
		}
	}

	def, err := loadWorkflowDefinition(s.workflowFile, proj.ID)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(def)
	if err != nil {
		return fmt.Errorf("marshal workflow %s: %w", def.Slug, err)
	}

	wf, err := workflows.FindBySlug(ctx, demoTenantID, proj.ID, def.Slug)
	if err != nil {
		if !isNotFound(err) {
			return fmt.Errorf("find workflow %s: %w", def.Slug, err)
		}
		desc := def.Description
		wf, err = workflows.Create(ctx, demoTenantID, proj.ID, def.Slug, def.Name, &desc, raw)
		if err != nil {
			return fmt.Errorf("create workflow %s: %w", def.Slug, err)
		}
	}

	versionNo := wf.CurrentVersion + 1
	if _, err := workflows.Publish(ctx, demoTenantID, proj.ID, wf.ID, versionNo, raw, entities.VersionStatusPublished, wf.Version); err != nil {
		return fmt.Errorf("publish workflow %s: %w", def.Slug, err)
	}

	slog.Info("seeded intent", "intent", s.intentID, "workflow", def.Slug, "project", s.slug)
	return nil
}

// loadWorkflowDefinition reads a canonical *.workflow.json, extracts the
// `workflow` envelope, and injects the projectId (engine format, PRD §161).
func loadWorkflowDefinition(file, projectID string) (*engine.WorkflowDefinition, error) {
	_, baseDir := repoRoot()
	raw, err := os.ReadFile(filepath.Join(baseDir, "docs", file))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}
	var envelope struct {
		Workflow engine.WorkflowDefinition `json:"workflow"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", file, err)
	}
	def := envelope.Workflow
	def.ProjectID = projectID
	return &def, nil
}

// repoRoot returns the directory containing go.work (the repo root), by walking
// up from the current working directory.
func repoRoot() (string, string) {
	dir, err := os.Getwd()
	if err != nil {
		return "", ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", dir
		}
		dir = parent
	}
}

func isNotFound(err error) bool {
	var de *domain.DomainError
	return errors.As(err, &de) && de.Code == domain.ErrNotFound
}
