package services

import (
	"context"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/domain/trace"
)

// TraceRecordInput is the application-owned trace boundary. It contains no
// provider client or credential and accepts only observed metadata.
type TraceRecordInput struct {
	TurnID            *string
	Stage             entities.RuntimeTraceStage
	Source            entities.RuntimeTraceSource
	Status            entities.RuntimeTraceStatus
	OccurredAt        time.Time
	CorrelationID     *string
	DurationMS        *int64
	ReasonCode        *string
	ErrorCode         *string
	ProviderAlias     *string
	ProviderReference *string
	Summary           *string
	Attributes        map[string]any
}

// ExternalProviderMetadata is the only provider-facing integration envelope
// accepted by the recorder. Raw request/response/document payloads have no
// field in this type.
type ExternalProviderMetadata struct {
	ProviderAlias      string
	OperationReference string
	Status             entities.RuntimeTraceStatus
	DurationMS         *int64
	CorrelationID      *string
	Summary            string
	Attributes         map[string]any
}

// RuntimeTraceRecorder sanitizes data before handing it to persistence.
type RuntimeTraceRecorder struct {
	repo repositories.IRuntimeTraceRepository
	now  func() time.Time
}

func NewRuntimeTraceRecorder(repo repositories.IRuntimeTraceRepository) *RuntimeTraceRecorder {
	return &RuntimeTraceRecorder{repo: repo, now: time.Now}
}

// Record appends one application-observed stage to the tenant's trace.
func (r *RuntimeTraceRecorder) Record(ctx context.Context, tenantID, instanceID string, input TraceRecordInput) (*entities.RuntimeTraceEntry, error) {
	if r == nil || r.repo == nil {
		return nil, nil
	}
	when := input.OccurredAt
	if when.IsZero() {
		when = r.now().UTC()
	}
	if input.Source == "" {
		input.Source = entities.RuntimeTraceSourceOpenState
	}
	if input.Status == "" {
		input.Status = entities.RuntimeTraceStatusSucceeded
	}
	return r.repo.Append(ctx, tenantID, repositories.AppendRuntimeTraceInput{
		WorkflowInstanceID: instanceID,
		TurnID:             input.TurnID,
		Stage:              input.Stage,
		Source:             input.Source,
		Status:             input.Status,
		OccurredAt:         when,
		CorrelationID:      input.CorrelationID,
		DurationMS:         input.DurationMS,
		ReasonCode:         input.ReasonCode,
		ErrorCode:          input.ErrorCode,
		ProviderAlias:      input.ProviderAlias,
		ProviderReference:  input.ProviderReference,
		Summary:            sanitizeSummary(input.Summary),
		Attributes:         trace.SanitizeAttributes(input.Attributes),
	})
}

// RecordExternal records only sanitized provider metadata supplied by an
// integration adapter. It never performs a provider request.
func (r *RuntimeTraceRecorder) RecordExternal(ctx context.Context, tenantID, instanceID string, stage entities.RuntimeTraceStage, metadata ExternalProviderMetadata) (*entities.RuntimeTraceEntry, error) {
	var reference *string
	if metadata.OperationReference != "" {
		reference = &metadata.OperationReference
	}
	var summary *string
	if metadata.Summary != "" {
		summary = &metadata.Summary
	}
	return r.Record(ctx, tenantID, instanceID, TraceRecordInput{
		Stage:             stage,
		Source:            entities.RuntimeTraceSourceExternalProvider,
		Status:            metadata.Status,
		CorrelationID:     metadata.CorrelationID,
		DurationMS:        metadata.DurationMS,
		ProviderAlias:     stringPtrOrNil(metadata.ProviderAlias),
		ProviderReference: reference,
		Summary:           summary,
		Attributes:        metadata.Attributes,
	})
}

func sanitizeSummary(summary *string) *string {
	if summary == nil {
		return nil
	}
	sanitized := trace.SanitizeValue(*summary).(string)
	return &sanitized
}

func stringPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
