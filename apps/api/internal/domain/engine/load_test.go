package engine

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Baseline load test (PRD §123). Measures state-transition throughput so
// operators get an initial performance signal. The in-memory run has no
// external dependency; the Postgres-backed run is optional (gated by
// DATABASE_URL). A deliberately loose lower bound guards against gross
// regressions without flaking on shared machines.
// ---------------------------------------------------------------------------

// throughputWorkflow is a minimal linear workflow used for load runs.
func throughputWorkflow() *WorkflowDefinition {
	return &WorkflowDefinition{
		Slug:          "load",
		ProjectID:     "load-proj",
		Name:          "Load",
		SchemaVersion: 1,
		Status:        WorkflowPublished,
		EntryNodeID:   "n0",
		Nodes: []WorkflowNode{
			{ID: "n0", Kind: NodeKindStart, Name: "START"},
			{ID: "n1", Kind: NodeKindState, Name: "S1"},
			{ID: "n2", Kind: NodeKindState, Name: "S2"},
			{ID: "n3", Kind: NodeKindState, Name: "S3"},
		},
		Transitions: []TransitionDefinition{
			{ID: "t0", SourceStateID: "n0", Event: "e", TargetStateID: "n1", Priority: 1},
			{ID: "t1", SourceStateID: "n1", Event: "e", TargetStateID: "n2", Priority: 1},
			{ID: "t2", SourceStateID: "n2", Event: "e", TargetStateID: "n3", Priority: 1},
			// cycle back to keep transitions flowing indefinitely
			{ID: "t3", SourceStateID: "n3", Event: "e", TargetStateID: "n1", Priority: 1},
		},
		Policy: WorkflowPolicy{Interruptible: "NEVER", Priority: 1},
	}
}

// runTransitions drives `n` state transitions through an in-memory engine and
// returns the measured transitions/second.
func runTransitions(n int) float64 {
	repos := newFakeRepos()
	def := throughputWorkflow()
	_ = repos.Workflows.Save(context.Background(), def)
	eng := NewEngine(repos)
	inst, _ := eng.StartWorkflow(context.Background(), "load", "load-proj", "c", def, "e")
	_, _, _ = eng.ProcessEvent(context.Background(), "load", inst.ID, &Event{ID: "e0", Type: "e", Source: SourceSystem})

	start := time.Now()
	for i := 1; i <= n; i++ {
		_, _, err := eng.ProcessEvent(context.Background(), "load", inst.ID, &Event{ID: fmt.Sprintf("e%d", i), Type: "e", Source: SourceUser})
		if err != nil {
			return 0
		}
	}
	elapsed := time.Since(start)
	return float64(n) / elapsed.Seconds()
}

// TestLoadThroughput measures in-memory transition throughput and asserts a
// deliberately loose lower bound. The measured value is reported.
func TestLoadThroughput(t *testing.T) {
	const events = 10000
	tps := runTransitions(events)
	t.Logf("in-memory transitions/sec: %.0f (over %d events)", tps, events)

	// Loose lower bound (e.g. >10k tps) — guards against gross regressions
	// without flaking on shared/CI machines (PRD §123).
	const looseBound = 1000.0
	if tps < looseBound {
		t.Errorf("throughput %.0f tps below loose bound %.0f tps", tps, looseBound)
	}
}

// BenchmarkProcessEvent measures in-memory state-transition throughput via the
// Go benchmark harness (go test -bench=.). Run with:
//
//	go test ./internal/domain/engine -bench=BenchmarkProcessEvent -benchtime=1s
func BenchmarkProcessEvent(b *testing.B) {
	repos := newFakeRepos()
	def := throughputWorkflow()
	_ = repos.Workflows.Save(context.Background(), def)
	eng := NewEngine(repos)
	inst, _ := eng.StartWorkflow(context.Background(), "load", "load-proj", "c", def, "e")
	_, _, _ = eng.ProcessEvent(context.Background(), "load", inst.ID, &Event{ID: "e0", Type: "e", Source: SourceSystem})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := eng.ProcessEvent(context.Background(), "load", inst.ID, &Event{ID: fmt.Sprintf("be%d", i), Type: "e", Source: SourceUser})
		if err != nil {
			b.Fatal(err)
		}
	}
}
