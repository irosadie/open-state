package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

// RuntimeCommandPort is the existing orchestration contract consumed by the
// Admin Console. Detail and state inspection remain owned by Runtime Inspector.
type RuntimeCommandPort interface {
	SuspendWorkflow(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error)
	ResumeWorkflow(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error)
	RetryWorkflow(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error)
	ListInstances(ctx context.Context, tenantID string) ([]entities.WorkflowInstance, error)
}

type AdminRuntimeService struct {
	runtime RuntimeCommandPort
	events  repositories.IEventBrowserRepository
	audit   *AuditWriter
}

func NewAdminRuntimeService(runtime RuntimeCommandPort, events repositories.IEventBrowserRepository, audit *AuditWriter) *AdminRuntimeService {
	return &AdminRuntimeService{runtime: runtime, events: events, audit: audit}
}

func (s *AdminRuntimeService) ListInstances(ctx context.Context, tenantID string) ([]dtos.InstanceDTO, error) {
	instances, err := s.runtime.ListInstances(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]dtos.InstanceDTO, 0, len(instances))
	for i := range instances {
		result = append(result, toInstanceDTO(&instances[i]))
	}
	return result, nil
}

func (s *AdminRuntimeService) Suspend(ctx context.Context, tenantID, actor, instanceID string, correlationID *string) (*dtos.InstanceDTO, error) {
	return s.command(ctx, tenantID, actor, instanceID, correlationID, entities.AuditActionWorkflowSuspended, "suspend", s.runtime.SuspendWorkflow)
}

func (s *AdminRuntimeService) Resume(ctx context.Context, tenantID, actor, instanceID string, correlationID *string) (*dtos.InstanceDTO, error) {
	return s.command(ctx, tenantID, actor, instanceID, correlationID, entities.AuditActionWorkflowResumed, "resume", s.runtime.ResumeWorkflow)
}

func (s *AdminRuntimeService) Retry(ctx context.Context, tenantID, actor, instanceID string, correlationID *string) (*dtos.InstanceDTO, error) {
	return s.command(ctx, tenantID, actor, instanceID, correlationID, entities.AuditActionWorkflowRetried, "retry", s.runtime.RetryWorkflow)
}

func (s *AdminRuntimeService) ListEvents(ctx context.Context, tenantID string, query EventBrowserQuery) (*dtos.EventPageDTO, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	filter := repositories.EventFilter{
		WorkflowInstanceID: optionalFilter(query.WorkflowInstanceID),
		Type:               optionalFilter(query.Type),
		Source:             optionalFilter(query.Source),
		CorrelationID:      optionalFilter(query.CorrelationID),
		Offset:             (page - 1) * pageSize,
		Limit:              pageSize,
	}
	total, err := s.events.CountEventsFiltered(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}
	events, err := s.events.ListEventsFiltered(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}
	data := make([]dtos.EventDTO, 0, len(events))
	for i := range events {
		data = append(data, toEventDTO(&events[i]))
	}
	return &dtos.EventPageDTO{
		Data:     data,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasNext:  int64(page*pageSize) < total,
	}, nil
}

func (s *AdminRuntimeService) GetEvent(ctx context.Context, tenantID, eventID string) (*dtos.EventDTO, error) {
	event, err := s.events.FindEventByID(ctx, tenantID, eventID)
	if err != nil {
		return nil, err
	}
	dto := toEventDTO(event)
	return &dto, nil
}

type EventBrowserQuery struct {
	WorkflowInstanceID string
	Type               string
	Source             string
	CorrelationID      string
	Page               int
	PageSize           int
}

func (s *AdminRuntimeService) command(ctx context.Context, tenantID, actor, instanceID string, correlationID *string, action entities.AuditAction, name string, command func(context.Context, string, string) (*entities.WorkflowInstance, error)) (*dtos.InstanceDTO, error) {
	instance, err := command(ctx, tenantID, instanceID)
	if err != nil {
		s.writeRuntimeAudit(ctx, tenantID, actor, instanceID, action, map[string]any{
			"action": name, "outcome": "rejected", "error": err.Error(),
		}, correlationID)
		return nil, err
	}
	s.writeRuntimeAudit(ctx, tenantID, actor, instanceID, action, map[string]any{
		"action": name, "outcome": "accepted", "status": instance.Status, "version": instance.Version,
	}, correlationID)
	dto := toInstanceDTO(instance)
	return &dto, nil
}

func (s *AdminRuntimeService) writeRuntimeAudit(ctx context.Context, tenantID, actor, instanceID string, action entities.AuditAction, after any, correlationID *string) {
	if s.audit != nil {
		s.audit.Write(ctx, tenantID, actor, action, "workflow_instance", instanceID, nil, after, correlationID)
	}
}

func optionalFilter(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func toInstanceDTO(instance *entities.WorkflowInstance) dtos.InstanceDTO {
	return dtos.InstanceDTO{
		ID:                     instance.ID,
		TenantID:               instance.TenantID,
		WorkflowID:             instance.WorkflowID,
		WorkflowVersionID:      instance.WorkflowVersionID,
		CorrelationKey:         nullableString(instance.CorrelationKey),
		Status:                 string(instance.Status),
		Version:                instance.Version,
		CurrentStateInstanceID: instance.CurrentStateInstanceID,
		StartedAt:              formatTimePtr(instance.StartedAt),
		CompletedAt:            formatTimePtr(instance.CompletedAt),
		ExpiresAt:              formatTimePtr(instance.ExpiresAt),
		CreatedAt:              instance.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:              instance.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toEventDTO(event *entities.Event) dtos.EventDTO {
	payload := event.Payload
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	return dtos.EventDTO{
		ID:                 event.ID,
		TenantID:           event.TenantID,
		EventID:            event.EventID,
		Type:               event.Type,
		Source:             string(event.Source),
		AggregateID:        nullableString(event.AggregateID),
		WorkflowInstanceID: event.WorkflowInstanceID,
		Sequence:           event.Sequence,
		Timestamp:          event.Timestamp.UTC().Format(time.RFC3339),
		Payload:            payload,
		CorrelationID:      nullableString(event.CorrelationID),
		CausationID:        nullableString(event.CausationID),
		CreatedAt:          event.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func formatTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}
