package services

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	domainengine "github.com/irosadie/open-state/api/internal/domain/engine"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// SimulationService adapts the HTTP Builder contract to the domain sandbox. It
// intentionally has no persistence or capability-provider dependencies.
type SimulationService struct {
	engine *domainengine.Engine
}

// NewSimulationService builds the isolated workflow simulation service.
func NewSimulationService() *SimulationService {
	return &SimulationService{engine: domainengine.NewEngine(domainengine.EngineRepositories{})}
}

// Simulate runs a supplied, potentially unsaved definition and maps the
// deterministic domain trace to a JSON-safe DTO.
func (s *SimulationService) Simulate(ctx context.Context, tenantID string, req dtos.SimulateWorkflowRequest) (*dtos.SimulationResultDTO, error) {
	if len(req.Definition) == 0 {
		return nil, domain.NewValidation("definition is required")
	}

	var definition domainengine.WorkflowDefinition
	if err := json.Unmarshal(req.Definition, &definition); err != nil {
		return nil, domain.NewValidation("definition must be valid JSON")
	}
	initialContext, err := decodeJSONMap(req.InitialContext)
	if err != nil {
		return nil, domain.NewValidation("initialContext must contain valid JSON values")
	}
	events := make([]domainengine.SimulationEvent, 0, len(req.Events))
	for index, input := range req.Events {
		if strings.TrimSpace(input.Type) == "" {
			return nil, domain.NewValidation("events[" + strconv.Itoa(index) + "].type is required")
		}
		payload, err := decodeJSONMap(input.Payload)
		if err != nil {
			return nil, domain.NewValidation("events[" + strconv.Itoa(index) + "].payload must contain valid JSON values")
		}
		events = append(events, domainengine.SimulationEvent{
			Type:    strings.TrimSpace(input.Type),
			Payload: payload,
			Source:  domainengine.SourceAdmin,
		})
	}

	result, err := s.engine.Simulate(ctx, domainengine.SimulationInput{
		TenantID:       tenantID,
		ProjectID:      "simulation",
		Definition:     &definition,
		InitialContext: initialContext,
		Events:         events,
	})
	if err != nil {
		return nil, domain.NewValidation(err.Error())
	}
	return mapSimulationResult(result)
}

func decodeJSONMap(input map[string]json.RawMessage) (map[string]any, error) {
	result := make(map[string]any, len(input))
	for key, raw := range input {
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

func encodeJSONMap(input map[string]any) (map[string]json.RawMessage, error) {
	result := make(map[string]json.RawMessage, len(input))
	for key, value := range input {
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		result[key] = raw
	}
	return result, nil
}

func mapSimulationResult(input *domainengine.SimulationResult) (*dtos.SimulationResultDTO, error) {
	finalContext, err := encodeJSONMap(input.FinalContext)
	if err != nil {
		return nil, domain.NewInternal("simulation result contains unsupported context")
	}
	result := &dtos.SimulationResultDTO{
		FinalState: dtos.SimulationStateDTO{
			ID:   input.FinalState.ID,
			Name: input.FinalState.Name,
			Kind: string(input.FinalState.Kind),
		},
		FinalContext: finalContext,
		FinalStatus:  string(input.FinalStatus),
		Steps:        make([]dtos.SimulationStepDTO, 0, len(input.Steps)),
	}
	for _, step := range input.Steps {
		context, err := encodeJSONMap(step.Context)
		if err != nil {
			return nil, domain.NewInternal("simulation step contains unsupported context")
		}
		eventPayload, err := encodeJSONMap(step.EventPayload)
		if err != nil {
			return nil, domain.NewInternal("simulation step contains unsupported payload")
		}
		dtoStep := dtos.SimulationStepDTO{
			Sequence:             step.Sequence,
			Outcome:              string(step.Outcome),
			EventType:            step.EventType,
			EventPayload:         eventPayload,
			StateBefore:          dtos.SimulationStateDTO{ID: step.StateBefore.ID, Name: step.StateBefore.Name, Kind: string(step.StateBefore.Kind)},
			Candidates:           make([]dtos.SimulationCandidateDTO, 0, len(step.Candidates)),
			SelectedTransitionID: step.SelectedTransitionID,
			Context:              context,
			CapabilityRequests:   make([]dtos.SimulationCapabilityRequestDTO, 0, len(step.CapabilityRequests)),
			ErrorCode:            step.ErrorCode,
			ErrorMessage:         step.ErrorMessage,
		}
		if step.StateAfter != nil {
			dtoStep.StateAfter = &dtos.SimulationStateDTO{ID: step.StateAfter.ID, Name: step.StateAfter.Name, Kind: string(step.StateAfter.Kind)}
		}
		for _, candidate := range step.Candidates {
			dtoStep.Candidates = append(dtoStep.Candidates, dtos.SimulationCandidateDTO{
				TransitionID: candidate.TransitionID,
				Event:        candidate.Event,
				Priority:     candidate.Priority,
				GuardPassed:  candidate.GuardPassed,
				GuardError:   candidate.GuardError,
			})
		}
		for _, request := range step.CapabilityRequests {
			dtoStep.CapabilityRequests = append(dtoStep.CapabilityRequests, dtos.SimulationCapabilityRequestDTO{
				Name:   request.Name,
				Mock:   request.Mock,
				Status: request.Status,
			})
		}
		result.Steps = append(result.Steps, dtoStep)
	}
	return result, nil
}
