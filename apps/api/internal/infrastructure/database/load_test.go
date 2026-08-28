package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ---------------------------------------------------------------------------
// Postgres-backed load test (PRD §123, task 5.2). Measures state-transition
// throughput including persistence cost, so operators get an end-to-end
// performance signal alongside the in-memory baseline.
//
// This test is gated by DATABASE_URL: if it is not set, the test is skipped so
// CI and machines without a database stay green. It requires the seed to have
// run (a published order-food workflow in the demo tenant/project) and a
// deterministic, loose throughput bound to avoid flakiness (PRD §123).
// ---------------------------------------------------------------------------

const (
	// seedTenant matches the seed's demo tenant so the seeded workflows are
	// available for the run (read-only lookup).
	seedTenant = "00000000-0000-0000-0000-000000000001"
	// loadTestTenant is a dedicated tenant for the instances the load test
	// creates, so cleanup never touches the demo/seed tenant's data.
	loadTestTenant  = "00000000-0000-0000-0000-0000000000ff"
	loadTestProject = "retail"
	loadTestSlug    = "order-food"
	loadTestStateKey = "n_check_stock"
)

// runPostgresTransitions measures the throughput of N atomic state transitions
// against the real Postgres instance repository. It returns the measured
// transitions/second.
func runPostgresTransitions(ctx context.Context, pool *pgxpool.Pool, n int) (float64, error) {
	instRepo := NewPgxInstanceRepository(pool)
	wfRepo := NewPgxWorkflowRepository(pool)
	projRepo := NewPgxProjectRepository(pool)

	// Resolve the seeded project slug to its UUID (repos require UUID project id).
	proj, err := projRepo.FindBySlug(ctx, seedTenant, loadTestProject)
	if err != nil {
		return 0, fmt.Errorf("find project %s: %w (run the seed first)", loadTestProject, err)
	}

	// Find the seeded order-food workflow + its current version (requires seed).
	wf, err := wfRepo.FindBySlug(ctx, seedTenant, proj.ID, loadTestSlug)
	if err != nil {
		return 0, fmt.Errorf("find seeded workflow %s: %w (run the seed first)", loadTestSlug, err)
	}
	version, err := wfRepo.FindCurrentVersion(ctx, seedTenant, proj.ID, wf.ID)
	if err != nil {
		return 0, fmt.Errorf("find current version: %w", err)
	}

	// Create a workflow instance pinned to the seeded version under the isolated
	// load-test tenant (instances have no FK on tenant_id; workflow/version FKs
	// reference the seeded workflow which exists).
	inst, err := instRepo.Create(ctx, loadTestTenant, repositories.CreateWorkflowInstanceInput{
		WorkflowID:        wf.ID,
		WorkflowVersionID: version.ID,
	})
	if err != nil {
		return 0, fmt.Errorf("create instance: %w", err)
	}

	// Create an initial state instance.
	state, err := instRepo.CreateStateInstance(ctx, loadTestTenant, repositories.CreateStateInstanceInput{
		WorkflowInstanceID: inst.ID,
		WorkflowVersionID:  version.ID,
		StateKey:           loadTestStateKey,
	})
	if err != nil {
		return 0, fmt.Errorf("create state instance: %w", err)
	}

	// Measure atomic transitions (exit -> enter -> repoint -> version bump).
	start := time.Now()
	for i := 0; i < n; i++ {
		next, err := instRepo.Transition(ctx, loadTestTenant, repositories.TransitionInput{
			WorkflowInstanceID:      inst.ID,
			ExpectedWorkflowVersion: inst.Version + i,
			ExitStateInstanceID:     state.ID,
			ExpectedExitVersion:     state.Version,
			NewWorkflowVersionID:    version.ID,
			NewStateKey:             loadTestStateKey,
		})
		if err != nil {
			return 0, fmt.Errorf("transition %d: %w", i, err)
		}
		state = next // next iteration exits the state we just entered
	}
	elapsed := time.Since(start)

	return float64(n) / elapsed.Seconds(), nil
}

// cleanupLoadTest removes rows created by the load test for its tenant.
func cleanupLoadTest(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	// state_instances cascade-delete with workflow_instances (ON DELETE CASCADE).
	if _, err := pool.Exec(ctx, `DELETE FROM workflow_instances WHERE tenant_id = $1`, loadTestTenant); err != nil {
		t.Logf("cleanup: %v", err)
	}
}

func TestPostgresLoadThroughput(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set; skipping Postgres-backed load test")
	}

	ctx := context.Background()
	pool, err := config.NewPool(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	defer cleanupLoadTest(t, pool)

	const events = 1000
	tps, err := runPostgresTransitions(ctx, pool, events)
	if err != nil {
		t.Fatalf("postgres throughput run: %v", err)
	}
	t.Logf("postgres transitions/sec: %.0f (over %d events)", tps, events)

	// Loose lower bound — guards against gross regressions, tolerant of slow CI.
	const looseBound = 50.0
	if tps < looseBound {
		t.Errorf("postgres throughput %.0f tps below loose bound %.0f tps", tps, looseBound)
	}
}

func TestPostgresLoadThroughputSkippedWithoutDB(t *testing.T) {
	// Ensure the test is skipped (not failed) when DATABASE_URL is absent, so
	// plain `go test ./...` stays green without a database.
	if os.Getenv("DATABASE_URL") != "" {
		t.Log("DATABASE_URL set; gated test active")
	}
	// No assertions; exists to document the gating contract.
}
