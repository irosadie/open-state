package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

func requirementsForCapabilities(ctx context.Context, deps Dependencies, tenantID, projectID string, names []string, before []string, evidence []entities.CapabilityExecutionEvidence) ([]entities.ProviderRequirement, error) {
	out := make([]entities.ProviderRequirement, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		var cap *entities.Capability
		if deps.CapabilityRegistry != nil {
			found, err := deps.CapabilityRegistry.FindByName(ctx, tenantID, name)
			if err == nil {
				cap = found
			}
		}
		// Provider requirements describe the direct third-party MCP path only.
		// Internal capabilities remain part of the state definition but do not
		// cause the LLM to call an unrelated provider MCP server.
		if cap != nil && cap.ProviderType != entities.ProviderTypeMCP {
			continue
		}

		req := entities.ProviderRequirement{
			Capability:        name,
			Purpose:           name,
			Required:          true,
			BeforeTransitions: append([]string(nil), before...),
			Status:            "PENDING",
		}
		bindingResolved := false
		if cap != nil {
			if cap.Description.Valid && strings.TrimSpace(cap.Description.String) != "" {
				req.Purpose = cap.Description.String
			}
			if deps.ProjectCapabilityBindings == nil {
				req.Status = "MISSING_MAPPING"
			} else {
				binding, bindingErr := deps.ProjectCapabilityBindings.FindByCapability(ctx, tenantID, projectID, cap.ID)
				if bindingErr == nil {
					if binding.Health == entities.ProjectCapabilityMCPBindingActive {
						bindingResolved = true
						if deps.GatewayMode != appservices.MCPGatewayModeSecure {
							req.ProviderServer = binding.ConnectionAlias
							req.Tool = binding.ToolName
						}
						if strings.TrimSpace(binding.ToolDescription) != "" && req.Purpose == name {
							req.Purpose = binding.ToolDescription
						}
					} else {
						req.Status = "UNAVAILABLE"
						req.Error = binding.HealthReason
					}
				} else {
					var notFound *domain.DomainError
					if errors.As(bindingErr, &notFound) && notFound.Code == domain.ErrNotFound {
						req.Status = "MISSING_MAPPING"
					} else {
						return nil, bindingErr
					}
				}
			}
		}
		for _, item := range evidence {
			if req.Status == "UNAVAILABLE" || req.Status == "MISSING_MAPPING" {
				break
			}
			if item.CapabilityID == "" || cap == nil || item.CapabilityID != cap.ID {
				continue
			}
			switch item.Status {
			case entities.CapabilityEvidenceSucceeded:
				req.Status = "SATISFIED"
			case entities.CapabilityEvidenceFailed:
				req.Status = "FAILED"
				if len(item.Error) > 0 {
					var failure map[string]any
					if json.Unmarshal(item.Error, &failure) == nil {
						if message, ok := failure["message"].(string); ok {
							req.Error = message
						}
					}
				}
			}
		}
		if !bindingResolved && req.Status == "PENDING" {
			req.Status = "MISSING_MAPPING"
		}
		out = append(out, req)
	}
	return out, nil
}

func entryRequirements(ctx context.Context, deps Dependencies, tenantID, projectID string, workflow *entities.Workflow) ([]entities.ProviderRequirement, error) {
	if workflow == nil || len(workflow.DraftDefinition) == 0 {
		return []entities.ProviderRequirement{}, nil
	}
	var def engine.WorkflowDefinition
	if err := json.Unmarshal(workflow.DraftDefinition, &def); err != nil {
		return nil, fmt.Errorf("invalid workflow definition: %w", err)
	}
	entry := def.EntryNodeID
	if entry == "" {
		for _, node := range def.Nodes {
			if node.Kind == engine.NodeKindStart {
				entry = node.ID
				break
			}
		}
	}
	for _, node := range def.Nodes {
		if node.ID != entry {
			continue
		}
		var transitions []string
		for _, transition := range def.Transitions {
			if transition.SourceStateID == node.ID {
				transitions = append(transitions, transition.Event)
			}
		}
		return requirementsForCapabilities(ctx, deps, tenantID, projectID, node.Capabilities, transitions, nil)
	}
	return []entities.ProviderRequirement{}, nil
}
