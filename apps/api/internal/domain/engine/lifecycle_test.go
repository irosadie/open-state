package engine

import (
	"context"
	"testing"
)

func TestSuspendResume(t *testing.T) {
	repos := newFakeRepos()
	def := padelDef()
	_ = repos.Workflows.Save(context.Background(), def)
	eng := NewEngine(repos)
	inst, _ := eng.StartWorkflow(context.Background(), "t", "c", def, "workflow.started")

	sus, err := eng.SuspendWorkflow(context.Background(), "t", inst.ID)
	if err != nil {
		t.Fatalf("suspend: %v", err)
	}
	if sus.Status != InstanceSuspended {
		t.Errorf("expected SUSPENDED, got %q", sus.Status)
	}
	if sus.CurrentStateID != "start" {
		t.Errorf("state not preserved on suspend: %q", sus.CurrentStateID)
	}

	// cannot process event while suspended
	_, _, err = eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{ID: "e0", Type: "workflow.started", Source: SourceSystem})
	if err == nil {
		t.Error("expected error processing event while suspended")
	}

	res, err := eng.ResumeWorkflow(context.Background(), "t", inst.ID)
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if res.Status != InstanceRunning {
		t.Errorf("expected RUNNING after resume, got %q", res.Status)
	}
}

func TestCancelWorkflow(t *testing.T) {
	repos := newFakeRepos()
	def := padelDef()
	_ = repos.Workflows.Save(context.Background(), def)
	eng := NewEngine(repos)
	inst, _ := eng.StartWorkflow(context.Background(), "t", "c", def, "workflow.started")

	cancel, err := eng.CancelWorkflow(context.Background(), "t", inst.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if cancel.Status != InstanceCancelled {
		t.Errorf("expected CANCELLED, got %q", cancel.Status)
	}
}

func TestOptimisticConcurrencyConflict(t *testing.T) {
	repos := newFakeRepos()
	def := padelDef()
	_ = repos.Workflows.Save(context.Background(), def)
	eng := NewEngine(repos)
	inst, _ := eng.StartWorkflow(context.Background(), "t", "c", def, "workflow.started")

	// simulate two concurrent updates both at version 1
	instA, _ := repos.Instances.Get(context.Background(), "t", inst.ID) // version 1
	instB, _ := repos.Instances.Get(context.Background(), "t", inst.ID) // version 1

	instA.Version = 2
	if err := repos.Instances.UpdateWithVersion(context.Background(), instA, 1); err != nil {
		t.Fatalf("first update should succeed: %v", err)
	}

	instB.Version = 2
	err := repos.Instances.UpdateWithVersion(context.Background(), instB, 1) // stale version
	if err == nil {
		t.Error("expected version conflict on stale update")
	}
}
