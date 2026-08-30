package services

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/domain/trace"
)

const (
	defaultRuntimePageSize = 20
	maxRuntimePageSize     = 100
)

// RuntimeInstanceQuery contains tenant-scoped discovery filters.
type RuntimeInstanceQuery struct {
	Status         *entities.WorkflowInstanceStatus
	WorkflowID     *string
	CorrelationKey *string
	Page           int
	PageSize       int
}

// RuntimeInspectorService composes existing runtime projections into an
// operator read model. It has no provider clients and performs no external I/O.
type RuntimeInspectorService struct {
	instances repositories.IInstanceRepository
	read      repositories.IRuntimeReadRepository
	events    repositories.IEventRepository
	contexts  repositories.IContextRepository
	audit     repositories.IAuditRepository
	traces    repositories.IRuntimeTraceRepository
}

func NewRuntimeInspectorService(
	instances repositories.IInstanceRepository,
	read repositories.IRuntimeReadRepository,
	events repositories.IEventRepository,
	contexts repositories.IContextRepository,
	audit repositories.IAuditRepository,
	traces repositories.IRuntimeTraceRepository,
) *RuntimeInspectorService {
	return &RuntimeInspectorService{instances: instances, read: read, events: events, contexts: contexts, audit: audit, traces: traces}
}

// List returns only instances from tenantID and applies filters before paging.
func (s *RuntimeInspectorService) List(ctx context.Context, tenantID string, q RuntimeInstanceQuery) (*dtos.RuntimeInstanceListDTO, error) {
	instances, err := s.instances.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	filtered := make([]entities.WorkflowInstance, 0, len(instances))
	for i := range instances {
		instance := instances[i]
		if q.Status != nil && instance.Status != *q.Status {
			continue
		}
		if q.WorkflowID != nil && instance.WorkflowID != *q.WorkflowID {
			continue
		}
		if q.CorrelationKey != nil && (!instance.CorrelationKey.Valid || instance.CorrelationKey.String != *q.CorrelationKey) {
			continue
		}
		filtered = append(filtered, instance)
	}

	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = defaultRuntimePageSize
	}
	if pageSize > maxRuntimePageSize {
		pageSize = maxRuntimePageSize
	}
	start := (page - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	data := make([]dtos.RuntimeInstanceSummaryDTO, 0, end-start)
	for i := start; i < end; i++ {
		data = append(data, s.summary(ctx, tenantID, &filtered[i]))
	}
	return &dtos.RuntimeInstanceListDTO{
		Data: data, Page: page, PageSize: pageSize, Total: len(filtered), HasNext: end < len(filtered),
	}, nil
}

// Get composes workflow/version, current state, sanitized context, and ordered
// event/state/decision activity for one tenant-scoped instance.
func (s *RuntimeInspectorService) Get(ctx context.Context, tenantID, instanceID string) (*dtos.RuntimeInstanceDetailDTO, error) {
	instance, err := s.instances.FindByID(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	stateInstances, err := s.stateInstances(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	states, err := s.states(ctx, tenantID, instance.WorkflowVersionID)
	if err != nil {
		return nil, err
	}
	stateByKey := make(map[string]entities.State, len(states))
	for _, state := range states {
		stateByKey[state.Key] = state
	}
	current := currentState(instance, stateInstances, stateByKey)
	contextDTO, err := s.context(ctx, tenantID, instance, current, stateByKey)
	if err != nil {
		return nil, err
	}
	events, err := s.events.ListEventsByInstance(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	audits := s.auditEntries(ctx, tenantID, instanceID)
	timeline := buildTimeline(stateInstances, stateByKey, events, audits)
	return &dtos.RuntimeInstanceDetailDTO{
		Summary:             s.summaryWithCurrent(ctx, tenantID, instance, current),
		CurrentState:        current,
		Context:             contextDTO,
		Timeline:            timeline,
		AuditCorrelationIDs: auditCorrelationIDs(audits),
	}, nil
}

// DebugTrace reads only OpenState's local trace projection. A nil/empty result
// is explicitly marked unavailable so the UI cannot infer provider success.
func (s *RuntimeInspectorService) DebugTrace(ctx context.Context, tenantID, instanceID, turnID string) (*dtos.RuntimeTraceDTO, error) {
	if _, err := s.instances.FindByID(ctx, tenantID, instanceID); err != nil {
		return nil, err
	}
	if s.traces == nil {
		return &dtos.RuntimeTraceDTO{Available: false, Data: []dtos.RuntimeTraceEntryDTO{}}, nil
	}
	var entries []entities.RuntimeTraceEntry
	var err error
	if turnID == "" {
		entries, err = s.traces.ListByInstance(ctx, tenantID, instanceID)
	} else {
		entries, err = s.traces.ListByTurn(ctx, tenantID, instanceID, turnID)
	}
	if err != nil {
		return nil, err
	}
	data := make([]dtos.RuntimeTraceEntryDTO, 0, len(entries))
	for i := range entries {
		data = append(data, toRuntimeTraceDTO(&entries[i]))
	}
	return &dtos.RuntimeTraceDTO{Available: len(data) > 0, Data: data}, nil
}

func (s *RuntimeInspectorService) summary(ctx context.Context, tenantID string, instance *entities.WorkflowInstance) dtos.RuntimeInstanceSummaryDTO {
	return s.summaryWithCurrent(ctx, tenantID, instance, s.currentStateForList(ctx, tenantID, instance))
}

func (s *RuntimeInspectorService) summaryWithCurrent(ctx context.Context, tenantID string, instance *entities.WorkflowInstance, current *dtos.RuntimeStateDTO) dtos.RuntimeInstanceSummaryDTO {
	workflow := dtos.RuntimeWorkflowDTO{ID: instance.WorkflowID, VersionID: instance.WorkflowVersionID}
	if s.read != nil {
		if definition, err := s.read.FindWorkflow(ctx, tenantID, instance.WorkflowID); err == nil {
			workflow.Name = definition.Name
			workflow.Slug = definition.Slug
		}
		if version, err := s.read.FindWorkflowVersion(ctx, tenantID, instance.WorkflowVersionID); err == nil {
			workflow.Version = version.VersionNo
		}
	}
	lastActivity := instance.UpdatedAt
	if lastActivity.IsZero() {
		lastActivity = instance.CreatedAt
	}
	var correlationID *string
	if instance.CorrelationKey.Valid {
		correlationID = &instance.CorrelationKey.String
	}
	return dtos.RuntimeInstanceSummaryDTO{
		ID: instance.ID, Workflow: workflow, Status: string(instance.Status), CurrentState: current,
		CorrelationID: correlationID, LastActivityAt: lastActivity.UTC().Format(time.RFC3339),
	}
}

func (s *RuntimeInspectorService) currentStateForList(ctx context.Context, tenantID string, instance *entities.WorkflowInstance) *dtos.RuntimeStateDTO {
	if s.read == nil {
		return nil
	}
	stateInstances, err := s.read.ListStateInstancesByWorkflowInstance(ctx, tenantID, instance.ID)
	if err != nil {
		return nil
	}
	states, err := s.read.ListStatesByVersion(ctx, tenantID, instance.WorkflowVersionID)
	if err != nil {
		return nil
	}
	byKey := make(map[string]entities.State, len(states))
	for _, state := range states {
		byKey[state.Key] = state
	}
	return currentState(instance, stateInstances, byKey)
}

func (s *RuntimeInspectorService) stateInstances(ctx context.Context, tenantID, instanceID string) ([]entities.StateInstance, error) {
	if s.read == nil {
		return []entities.StateInstance{}, nil
	}
	return s.read.ListStateInstancesByWorkflowInstance(ctx, tenantID, instanceID)
}

func (s *RuntimeInspectorService) states(ctx context.Context, tenantID, versionID string) ([]entities.State, error) {
	if s.read == nil {
		return []entities.State{}, nil
	}
	return s.read.ListStatesByVersion(ctx, tenantID, versionID)
}

func (s *RuntimeInspectorService) context(ctx context.Context, tenantID string, instance *entities.WorkflowInstance, current *dtos.RuntimeStateDTO, states map[string]entities.State) (dtos.RuntimeContextDTO, error) {
	available := map[string]any{}
	redacted := false
	load := func(scopeType entities.ContextScopeType, scopeID string) error {
		if s.contexts == nil || scopeID == "" {
			return nil
		}
		records, err := s.contexts.ListContextByScope(ctx, tenantID, scopeType, scopeID)
		if err != nil {
			return err
		}
		for _, record := range records {
			value := rawJSONValue(record.Value)
			sanitized := trace.SanitizeAttributes(map[string]any{record.Key: value})[record.Key]
			if !jsonEqual(value, sanitized) {
				redacted = true
			}
			available[record.Key] = sanitized
		}
		return nil
	}
	if instance.CorrelationKey.Valid {
		if err := load(entities.ContextScopeConversation, instance.CorrelationKey.String); err != nil {
			return dtos.RuntimeContextDTO{}, err
		}
	}
	if err := load(entities.ContextScopeWorkflowInstance, instance.ID); err != nil {
		return dtos.RuntimeContextDTO{}, err
	}
	if current != nil {
		if err := load(entities.ContextScopeStateInstance, current.ID); err != nil {
			return dtos.RuntimeContextDTO{}, err
		}
	}
	missing := []string{}
	if current != nil {
		if state, ok := states[current.Key]; ok {
			var required []string
			if err := json.Unmarshal(state.RequiredContext, &required); err == nil {
				for _, key := range required {
					if _, ok := available[key]; !ok {
						missing = append(missing, key)
					}
				}
			}
		}
	}
	sort.Strings(missing)
	return dtos.RuntimeContextDTO{Available: available, Missing: missing, Redacted: redacted}, nil
}

func (s *RuntimeInspectorService) auditEntries(ctx context.Context, tenantID, instanceID string) []entities.AuditLog {
	if s.audit == nil {
		return nil
	}
	entries, err := s.audit.ListByResource(ctx, tenantID, "workflow_instance", instanceID)
	if err == nil && len(entries) > 0 {
		return entries
	}
	entries, _ = s.audit.ListByResource(ctx, tenantID, "instance", instanceID)
	return entries
}

func currentState(instance *entities.WorkflowInstance, occurrences []entities.StateInstance, definitions map[string]entities.State) *dtos.RuntimeStateDTO {
	var selected *entities.StateInstance
	for i := range occurrences {
		occurrence := &occurrences[i]
		if instance.CurrentStateInstanceID != nil && occurrence.ID == *instance.CurrentStateInstanceID {
			selected = occurrence
			break
		}
		if selected == nil || occurrence.EnteredAt.After(selected.EnteredAt) {
			selected = occurrence
		}
	}
	if selected == nil {
		return nil
	}
	name := selected.StateKey
	if definition, ok := definitions[selected.StateKey]; ok && definition.Name != "" {
		name = definition.Name
	}
	var exitedAt *string
	if selected.ExitedAt != nil {
		value := selected.ExitedAt.UTC().Format(time.RFC3339)
		exitedAt = &value
	}
	return &dtos.RuntimeStateDTO{ID: selected.ID, Key: selected.StateKey, Name: name, Status: string(selected.Status), EnteredAt: selected.EnteredAt.UTC().Format(time.RFC3339), ExitedAt: exitedAt}
}

func buildTimeline(states []entities.StateInstance, definitions map[string]entities.State, events []entities.Event, audits []entities.AuditLog) []dtos.RuntimeTimelineEntryDTO {
	timeline := make([]dtos.RuntimeTimelineEntryDTO, 0, len(states)+len(events)+len(audits))
	for i := range states {
		state := &states[i]
		label := state.StateKey
		if definition, ok := definitions[state.StateKey]; ok && definition.Name != "" {
			label = definition.Name
		}
		timeline = append(timeline, dtos.RuntimeTimelineEntryDTO{ID: state.ID, Kind: "STATE", Type: "state", Label: label, Status: string(state.Status), Sequence: int64(i + 1), OccurredAt: state.EnteredAt.UTC().Format(time.RFC3339)})
	}
	for i := range events {
		event := &events[i]
		var correlation *string
		if event.CorrelationID.Valid {
			correlation = &event.CorrelationID.String
		}
		timeline = append(timeline, dtos.RuntimeTimelineEntryDTO{ID: event.ID, Kind: "EVENT", Type: event.Type, Label: event.Type, Status: "RECORDED", Sequence: event.Sequence, OccurredAt: event.Timestamp.UTC().Format(time.RFC3339), CorrelationID: correlation})
	}
	for i := range audits {
		audit := &audits[i]
		reason := reasonCodeForAudit(audit)
		var correlation *string
		if audit.CorrelationID != nil {
			correlation = audit.CorrelationID
		}
		timeline = append(timeline, dtos.RuntimeTimelineEntryDTO{ID: audit.ID, Kind: "DECISION", Type: string(audit.Action), Label: audit.ResourceType, Status: "RECORDED", Sequence: int64(i + 1), OccurredAt: audit.OccurredAt.UTC().Format(time.RFC3339), CorrelationID: correlation, ReasonCode: reason})
	}
	sort.SliceStable(timeline, func(i, j int) bool {
		left, _ := time.Parse(time.RFC3339, timeline[i].OccurredAt)
		right, _ := time.Parse(time.RFC3339, timeline[j].OccurredAt)
		if left.Equal(right) {
			return timeline[i].Sequence < timeline[j].Sequence
		}
		return left.Before(right)
	})
	return timeline
}

func reasonCodeForAudit(audit *entities.AuditLog) *string {
	reason := ""
	switch audit.Action {
	case entities.AuditActionGuardFailed:
		reason = "GUARD_FAILED"
	case entities.AuditActionTransitionExecuted:
		reason = "TRANSITION_EXECUTED"
	case entities.AuditActionStateEntered:
		reason = "STATE_ENTERED"
	case entities.AuditActionCapabilityDenied:
		reason = "CAPABILITY_DENIED"
	}
	if reason == "" {
		return nil
	}
	return &reason
}

func auditCorrelationIDs(entries []entities.AuditLog) []string {
	seen := map[string]struct{}{}
	ids := []string{}
	for _, entry := range entries {
		if entry.CorrelationID == nil || *entry.CorrelationID == "" {
			continue
		}
		if _, ok := seen[*entry.CorrelationID]; ok {
			continue
		}
		seen[*entry.CorrelationID] = struct{}{}
		ids = append(ids, *entry.CorrelationID)
	}
	sort.Strings(ids)
	return ids
}

func toRuntimeTraceDTO(entry *entities.RuntimeTraceEntry) dtos.RuntimeTraceEntryDTO {
	attributes := map[string]any{}
	for key, value := range entry.Attributes {
		attributes[key] = trace.SanitizeValue(value)
	}
	return dtos.RuntimeTraceEntryDTO{ID: entry.ID, TurnID: entry.TurnID, Sequence: entry.Sequence, Stage: string(entry.Stage), Source: string(entry.Source), Status: string(entry.Status), OccurredAt: entry.OccurredAt.UTC().Format(time.RFC3339), CorrelationID: entry.CorrelationID, DurationMS: entry.DurationMS, ReasonCode: entry.ReasonCode, ErrorCode: entry.ErrorCode, ProviderAlias: entry.ProviderAlias, ProviderReference: entry.ProviderReference, Summary: entry.Summary, Attributes: attributes}
}

func rawJSONValue(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return trace.RedactedMarker
	}
	return value
}

func jsonEqual(left, right any) bool {
	leftBytes, _ := json.Marshal(left)
	rightBytes, _ := json.Marshal(right)
	return string(leftBytes) == string(rightBytes)
}
