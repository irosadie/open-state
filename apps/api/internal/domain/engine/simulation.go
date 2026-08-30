package engine

import (
	"context"
	"fmt"
)

// MaxSimulationEvents bounds work and response size for an unpersisted sandbox
// request. It is deliberately a domain limit so all callers get the same guard.
const MaxSimulationEvents = 100

type SimulationEvent struct {
	Type    string
	Payload map[string]any
	Source  EventSource
}

type SimulationInput struct {
	TenantID       string
	ProjectID      string
	Definition     *WorkflowDefinition
	InitialContext map[string]any
	Events         []SimulationEvent
}

type SimulationOutcome string

const (
	SimulationEntered      SimulationOutcome = "ENTERED"
	SimulationTransitioned SimulationOutcome = "TRANSITIONED"
	SimulationRejected     SimulationOutcome = "REJECTED"
)

type SimulationState struct {
	ID   string
	Name string
	Kind WorkflowNodeKind
}

type SimulationCandidate struct {
	TransitionID string
	Event        string
	Priority     int
	GuardPassed  bool
	GuardError   string
}

type SimulationCapabilityRequest struct {
	Name   string
	Mock   bool
	Status string
}

type SimulationStep struct {
	Sequence             int
	Outcome              SimulationOutcome
	EventType            string
	EventPayload         map[string]any
	StateBefore          SimulationState
	StateAfter           *SimulationState
	Candidates           []SimulationCandidate
	SelectedTransitionID string
	Context              map[string]any
	CapabilityRequests   []SimulationCapabilityRequest
	ErrorCode            string
	ErrorMessage         string
}

type SimulationResult struct {
	FinalState   SimulationState
	FinalContext map[string]any
	FinalStatus  WorkflowInstanceStatus
	Steps        []SimulationStep
}

// Simulate executes a draft against fresh in-memory repositories. No caller's
// repositories are read or written, and the result contains only deterministic
// workflow facts suitable for an operator-facing trace.
func (e *Engine) Simulate(ctx context.Context, input SimulationInput) (*SimulationResult, error) {
	if input.Definition == nil {
		return nil, fmt.Errorf("workflow definition is required")
	}
	if input.Definition.EntryNodeID == "" {
		return nil, fmt.Errorf("workflow entry node is required")
	}
	if len(input.Definition.Nodes) == 0 {
		return nil, fmt.Errorf("workflow has no nodes")
	}
	if len(input.Events) > MaxSimulationEvents {
		return nil, fmt.Errorf("simulation supports at most %d events", MaxSimulationEvents)
	}
	projectID := input.ProjectID
	if projectID == "" {
		projectID = "simulation"
	}

	def := cloneWorkflowDefinition(input.Definition)
	// Always pin the draft to the ephemeral project supplied by the caller;
	// client-provided project identifiers must never select a persistent scope.
	def.ProjectID = projectID
	repos := replayRepos()
	sandbox := NewEngine(repos)
	if err := repos.Workflows.Save(ctx, def); err != nil {
		return nil, err
	}
	inst, err := sandbox.StartWorkflow(ctx, input.TenantID, projectID, "simulation", def, entryEvent(def))
	if err != nil {
		return nil, err
	}
	inst.Context = cloneContext(input.InitialContext)
	if err := repos.Instances.UpdateWithVersion(ctx, inst, inst.Version); err != nil {
		return nil, err
	}

	result := &SimulationResult{Steps: make([]SimulationStep, 0, len(input.Events)+1)}
	entry, ok := nodeByID(def, inst.CurrentStateID)
	if !ok {
		return nil, fmt.Errorf("workflow entry node not found: %s", inst.CurrentStateID)
	}
	result.Steps = append(result.Steps, SimulationStep{
		Sequence:           0,
		Outcome:            SimulationEntered,
		StateBefore:        stateSnapshot(entry),
		Context:            cloneContext(inst.Context),
		CapabilityRequests: capabilityRequests(entry),
	})

	for index, simEvent := range input.Events {
		current, err := repos.Instances.Get(ctx, input.TenantID, inst.ID)
		if err != nil {
			return nil, err
		}
		currentNode, ok := nodeByID(def, current.CurrentStateID)
		if !ok {
			return nil, fmt.Errorf("current state not found in workflow: %s", current.CurrentStateID)
		}
		payload := cloneContext(simEvent.Payload)
		candidateContext := cloneContext(current.Context)
		mergePayload(candidateContext, payload)
		candidates := transitionsFrom(def, current.CurrentStateID, simEvent.Type)
		selected, evaluations, evaluationErr := evaluateTransitions(candidates, candidateContext)
		step := SimulationStep{
			Sequence:     index + 1,
			Outcome:      SimulationRejected,
			EventType:    simEvent.Type,
			EventPayload: payload,
			StateBefore:  stateSnapshot(currentNode),
			Context:      cloneContext(current.Context),
			Candidates:   simulationCandidates(evaluations),
		}
		if len(candidates) == 0 {
			step.ErrorCode = "EVENT_NOT_ALLOWED"
			step.ErrorMessage = "event is not allowed from the current state"
			result.Steps = append(result.Steps, step)
			break
		}
		if evaluationErr != nil {
			step.ErrorCode = "GUARD_EVALUATION_ERROR"
			step.ErrorMessage = "one or more guards could not be evaluated"
			result.Steps = append(result.Steps, step)
			break
		}
		if selected == nil {
			step.ErrorCode = "GUARD_FAILED"
			step.ErrorMessage = "no candidate transition passed its guards"
			result.Steps = append(result.Steps, step)
			break
		}

		next, _, processErr := sandbox.ProcessEvent(ctx, input.TenantID, inst.ID, &Event{
			ID:      fmt.Sprintf("simulation-event-%d", index+1),
			Type:    simEvent.Type,
			Source:  simulationSource(simEvent.Source),
			Payload: payload,
		})
		if processErr != nil {
			return nil, processErr
		}
		targetNode, ok := nodeByID(def, selected.TargetStateID)
		if !ok {
			return nil, fmt.Errorf("transition target not found: %s", selected.TargetStateID)
		}
		step.Outcome = SimulationTransitioned
		afterState := stateSnapshot(targetNode)
		step.StateAfter = &afterState
		step.SelectedTransitionID = selected.ID
		step.Context = cloneContext(next.Context)
		step.CapabilityRequests = capabilityRequests(targetNode)
		result.Steps = append(result.Steps, step)
		inst = next
	}

	final, err := repos.Instances.Get(ctx, input.TenantID, inst.ID)
	if err != nil {
		return nil, err
	}
	finalNode, ok := nodeByID(def, final.CurrentStateID)
	if !ok {
		return nil, fmt.Errorf("final state not found in workflow: %s", final.CurrentStateID)
	}
	result.FinalState = stateSnapshot(finalNode)
	result.FinalContext = cloneContext(final.Context)
	result.FinalStatus = InstanceRunning
	if finalNode.IsTerminal || finalNode.Kind == NodeKindEnd {
		result.FinalStatus = InstanceCompleted
	}
	return result, nil
}

func simulationSource(source EventSource) EventSource {
	if source == "" {
		return SourceAdmin
	}
	return source
}

func stateSnapshot(node WorkflowNode) SimulationState {
	return SimulationState{ID: node.ID, Name: node.Name, Kind: node.Kind}
}

func capabilityRequests(node WorkflowNode) []SimulationCapabilityRequest {
	requests := make([]SimulationCapabilityRequest, 0, len(node.Capabilities))
	for _, name := range node.Capabilities {
		requests = append(requests, SimulationCapabilityRequest{Name: name, Mock: true, Status: "PLANNED"})
	}
	return requests
}

func simulationCandidates(evaluations []transitionEvaluation) []SimulationCandidate {
	candidates := make([]SimulationCandidate, 0, len(evaluations))
	for _, evaluation := range evaluations {
		candidate := SimulationCandidate{
			TransitionID: evaluation.transition.ID,
			Event:        evaluation.transition.Event,
			Priority:     evaluation.transition.Priority,
			GuardPassed:  evaluation.passed,
		}
		if evaluation.err != nil {
			candidate.GuardError = "guard evaluation failed"
		}
		candidates = append(candidates, candidate)
	}
	return candidates
}

func cloneWorkflowDefinition(def *WorkflowDefinition) *WorkflowDefinition {
	copyDef := *def
	copyDef.Nodes = append([]WorkflowNode(nil), def.Nodes...)
	copyDef.Transitions = append([]TransitionDefinition(nil), def.Transitions...)
	copyDef.Triggers = append([]WorkflowTrigger(nil), def.Triggers...)
	return &copyDef
}

func cloneContext(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = cloneValue(value)
	}
	return output
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneContext(typed)
	case []any:
		items := make([]any, len(typed))
		for index, item := range typed {
			items[index] = cloneValue(item)
		}
		return items
	default:
		return value
	}
}
