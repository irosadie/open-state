package engine

import (
	"context"

	"github.com/irosadie/open-state/go-shared/domain"
)

// SuspendWorkflow pauses a running workflow, preserving state, context,
// history, and version (PRD §43).
func (e *Engine) SuspendWorkflow(ctx context.Context, tenantID, instanceID string) (*WorkflowInstance, error) {
	instance, err := e.repos.Instances.Get(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	if instance.Status != InstanceRunning && instance.Status != InstanceWaiting {
		return nil, domain.NewConflict("only running/waiting instances can be suspended (status: " + string(instance.Status) + ")")
	}

	instance.Status = InstanceSuspended
	instance.Version++
	if err := e.repos.Instances.UpdateWithVersion(ctx, instance, instance.Version-1); err != nil {
		return nil, err
	}
	return instance, nil
}

// ResumeWorkflow restores a suspended instance to running, continuing from its
// saved state (PRD §43).
func (e *Engine) ResumeWorkflow(ctx context.Context, tenantID, instanceID string) (*WorkflowInstance, error) {
	instance, err := e.repos.Instances.Get(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	if instance.Status != InstanceSuspended {
		return nil, domain.NewConflict("only suspended instances can be resumed (status: " + string(instance.Status) + ")")
	}

	instance.Status = InstanceRunning
	instance.Version++
	if err := e.repos.Instances.UpdateWithVersion(ctx, instance, instance.Version-1); err != nil {
		return nil, err
	}
	return instance, nil
}

// CancelWorkflow cancels a workflow instance (PRD §137).
func (e *Engine) CancelWorkflow(ctx context.Context, tenantID, instanceID string) (*WorkflowInstance, error) {
	instance, err := e.repos.Instances.Get(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	if instance.Status == InstanceCompleted || instance.Status == InstanceCancelled {
		return nil, domain.NewConflict("workflow already terminal")
	}

	instance.Status = InstanceCancelled
	instance.Version++
	if err := e.repos.Instances.UpdateWithVersion(ctx, instance, instance.Version-1); err != nil {
		return nil, err
	}
	return instance, nil
}
