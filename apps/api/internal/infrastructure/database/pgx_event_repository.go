package database

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	sharedomain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/sqlc-dev/pqtype"
)

// PgxEventRepository implements IEventRepository on PostgreSQL via sqlc.
type PgxEventRepository struct {
	queries *db.Queries
}

// NewPgxEventRepository returns a PostgreSQL-backed event repository.
func NewPgxEventRepository(pool *pgxpool.Pool) *PgxEventRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxEventRepository(db.New(sqlDB))
}

// newPgxEventRepository builds an IEventRepository from a sqlc queries handle.
// It enables the composed PostgresAdapter to bind this repository to a shared
// transaction via WithTx.
func newPgxEventRepository(q *db.Queries) *PgxEventRepository {
	return &PgxEventRepository{queries: q}
}

func (r *PgxEventRepository) Append(ctx context.Context, tenantID string, input repositories.AppendEventInput) (*entities.Event, error) {
	row, err := r.queries.AppendEvent(ctx, db.AppendEventParams{
		TenantID:           mustUUID(tenantID),
		EventID:            input.EventID,
		Type:               input.Type,
		Source:             string(input.Source),
		AggregateID:        nullString(input.AggregateID),
		WorkflowInstanceID: optUUID(input.WorkflowInstanceID),
		Timestamp:          input.Timestamp,
		Payload:            input.Payload,
		CorrelationID:      nullString(input.CorrelationID),
		CausationID:        nullString(input.CausationID),
		IdempotencyKey:     nullString(input.IdempotencyKey),
	})
	if err != nil {
		return nil, mapPgError(err, "append event")
	}
	return mapEvent(row), nil
}

func (r *PgxEventRepository) FindEventByID(ctx context.Context, tenantID, id string) (*entities.Event, error) {
	eventUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, sharedomain.NewValidation("invalid event id")
	}
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, sharedomain.NewValidation("invalid tenant id")
	}
	row, err := r.queries.FindEventByID(ctx, db.FindEventByIDParams{
		ID:       eventUUID,
		TenantID: tenantUUID,
	})
	if err != nil {
		return nil, mapNotFound(err, "event")
	}
	return mapEvent(row), nil
}

func (r *PgxEventRepository) ListEventsByInstance(ctx context.Context, tenantID, workflowInstanceID string) ([]entities.Event, error) {
	rows, err := r.queries.ListEventsByInstance(ctx, db.ListEventsByInstanceParams{
		WorkflowInstanceID: optUUID(&workflowInstanceID),
		TenantID:           mustUUID(tenantID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.Event, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapEvent(row))
	}
	return out, nil
}

func (r *PgxEventRepository) ListEventsByTenant(ctx context.Context, tenantID string) ([]entities.Event, error) {
	rows, err := r.queries.ListEventsByTenant(ctx, mustUUID(tenantID))
	if err != nil {
		return nil, err
	}
	out := make([]entities.Event, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapEvent(row))
	}
	return out, nil
}

func (r *PgxEventRepository) ListEventsFiltered(ctx context.Context, tenantID string, filter repositories.EventFilter) ([]entities.Event, error) {
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, sharedomain.NewValidation("invalid tenant id")
	}
	workflowInstanceID, err := optionalUUID(filter.WorkflowInstanceID)
	if err != nil {
		return nil, sharedomain.NewValidation("invalid workflow instance id")
	}
	rows, err := r.queries.ListEventsFiltered(ctx, db.ListEventsFilteredParams{
		TenantID:           tenantUUID,
		WorkflowInstanceID: workflowInstanceID,
		Type:               nullString(filter.Type),
		Source:             nullString(filter.Source),
		CorrelationID:      nullString(filter.CorrelationID),
		PageSize:           int32(filter.Limit),
		PageOffset:         int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.Event, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapEvent(row))
	}
	return out, nil
}

func (r *PgxEventRepository) CountEventsFiltered(ctx context.Context, tenantID string, filter repositories.EventFilter) (int64, error) {
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return 0, sharedomain.NewValidation("invalid tenant id")
	}
	workflowInstanceID, err := optionalUUID(filter.WorkflowInstanceID)
	if err != nil {
		return 0, sharedomain.NewValidation("invalid workflow instance id")
	}
	return r.queries.CountEventsFiltered(ctx, db.CountEventsFilteredParams{
		TenantID:           tenantUUID,
		WorkflowInstanceID: workflowInstanceID,
		Type:               nullString(filter.Type),
		Source:             nullString(filter.Source),
		CorrelationID:      nullString(filter.CorrelationID),
	})
}

func (r *PgxEventRepository) InsertInbox(ctx context.Context, tenantID string, input repositories.InsertInboxEventInput) (*entities.InboxEvent, error) {
	row, err := r.queries.InsertInboxEvent(ctx, db.InsertInboxEventParams{
		TenantID:       mustUUID(tenantID),
		IdempotencyKey: input.IdempotencyKey,
		EventType:      input.EventType,
		Source:         string(input.Source),
		Payload:        input.Payload,
	})
	if err != nil {
		return nil, mapPgError(err, "insert inbox event")
	}
	return mapInboxEvent(row), nil
}

func (r *PgxEventRepository) ClaimInbox(ctx context.Context, tenantID string, limit int) ([]entities.InboxEvent, error) {
	rows, err := r.queries.ClaimInboxEvents(ctx, db.ClaimInboxEventsParams{
		TenantID: mustUUID(tenantID),
		Limit:    int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.InboxEvent, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapInboxEvent(row))
	}
	return out, nil
}

func (r *PgxEventRepository) MarkInboxProcessed(ctx context.Context, tenantID, id string) (*entities.InboxEvent, error) {
	row, err := r.queries.MarkInboxProcessed(ctx, db.MarkInboxProcessedParams{
		ID:       mustUUID(id),
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return nil, mapNotFound(err, "inbox event")
	}
	return mapInboxEvent(row), nil
}

func (r *PgxEventRepository) InsertOutbox(ctx context.Context, tenantID string, input repositories.InsertOutboxEventInput) (*entities.OutboxEvent, error) {
	row, err := r.queries.InsertOutboxEvent(ctx, db.InsertOutboxEventParams{
		TenantID: mustUUID(tenantID),
		EventID:  optUUID(input.EventID),
		Payload:  input.Payload,
		Topic:    input.Topic,
	})
	if err != nil {
		return nil, mapPgError(err, "insert outbox event")
	}
	return mapOutboxEvent(row), nil
}

func (r *PgxEventRepository) ClaimOutbox(ctx context.Context, tenantID string, limit int) ([]entities.OutboxEvent, error) {
	rows, err := r.queries.ClaimOutboxEvents(ctx, db.ClaimOutboxEventsParams{
		TenantID: mustUUID(tenantID),
		Limit:    int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.OutboxEvent, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapOutboxEvent(row))
	}
	return out, nil
}

func (r *PgxEventRepository) MarkOutboxPublished(ctx context.Context, tenantID, id string) (*entities.OutboxEvent, error) {
	row, err := r.queries.MarkOutboxPublished(ctx, db.MarkOutboxPublishedParams{
		ID:       mustUUID(id),
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return nil, mapNotFound(err, "outbox event")
	}
	return mapOutboxEvent(row), nil
}

func (r *PgxEventRepository) UpsertIdempotency(ctx context.Context, tenantID string, input repositories.UpsertIdempotencyInput) (*entities.IdempotencyRecord, error) {
	row, err := r.queries.UpsertIdempotencyRecord(ctx, db.UpsertIdempotencyRecordParams{
		TenantID:       mustUUID(tenantID),
		IdempotencyKey: input.IdempotencyKey,
		Scope:          input.Scope,
		ResultStatus:   string(input.ResultStatus),
		Payload:        nullRawMessage(input.Payload),
	})
	if err != nil {
		return nil, mapPgError(err, "upsert idempotency record")
	}
	return mapIdempotencyRecord(row), nil
}

func (r *PgxEventRepository) FindIdempotency(ctx context.Context, tenantID, idempotencyKey string) (*entities.IdempotencyRecord, error) {
	row, err := r.queries.FindIdempotencyRecord(ctx, db.FindIdempotencyRecordParams{
		TenantID:       mustUUID(tenantID),
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, mapNotFound(err, "idempotency record")
	}
	return mapIdempotencyRecord(row), nil
}

// ---- mappers ----

func mapEvent(row db.Event) *entities.Event {
	return &entities.Event{
		ID:                 row.ID.String(),
		TenantID:           row.TenantID.String(),
		EventID:            row.EventID,
		Type:               row.Type,
		Source:             entities.EventSource(row.Source),
		AggregateID:        row.AggregateID,
		WorkflowInstanceID: nullUUIDPtr(row.WorkflowInstanceID),
		Sequence:           row.Sequence,
		Timestamp:          row.Timestamp,
		Payload:            row.Payload,
		CorrelationID:      row.CorrelationID,
		CausationID:        row.CausationID,
		IdempotencyKey:     row.IdempotencyKey,
		CreatedAt:          row.CreatedAt,
	}
}

func mapInboxEvent(row db.EventInbox) *entities.InboxEvent {
	return &entities.InboxEvent{
		ID:             row.ID.String(),
		TenantID:       row.TenantID.String(),
		IdempotencyKey: row.IdempotencyKey,
		EventType:      row.EventType,
		Source:         entities.EventSource(row.Source),
		Payload:        row.Payload,
		Status:         entities.InboxEventStatus(row.Status),
		AttemptCount:   int(row.AttemptCount),
		ReceivedAt:     row.ReceivedAt,
		ProcessedAt:    nullTimePtr(row.ProcessedAt),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func mapOutboxEvent(row db.EventOutbox) *entities.OutboxEvent {
	return &entities.OutboxEvent{
		ID:           row.ID.String(),
		TenantID:     row.TenantID.String(),
		EventID:      nullUUIDPtr(row.EventID),
		Payload:      row.Payload,
		Topic:        row.Topic,
		Status:       entities.OutboxEventStatus(row.Status),
		AttemptCount: int(row.AttemptCount),
		PublishedAt:  nullTimePtr(row.PublishedAt),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func mapIdempotencyRecord(row db.IdempotencyRecord) *entities.IdempotencyRecord {
	return &entities.IdempotencyRecord{
		ID:             row.ID.String(),
		TenantID:       row.TenantID.String(),
		IdempotencyKey: row.IdempotencyKey,
		Scope:          row.Scope,
		ResultStatus:   entities.IdempotencyResultStatus(row.ResultStatus),
		Payload:        row.Payload.RawMessage,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func optionalUUID(value *string) (uuid.NullUUID, error) {
	if value == nil || *value == "" {
		return uuid.NullUUID{}, nil
	}
	id, err := uuid.Parse(*value)
	if err != nil {
		return uuid.NullUUID{}, err
	}
	return uuid.NullUUID{UUID: id, Valid: true}, nil
}

// ---- helpers ----

func nullRawMessage(raw json.RawMessage) pqtype.NullRawMessage {
	if raw == nil {
		return pqtype.NullRawMessage{Valid: false}
	}
	return pqtype.NullRawMessage{RawMessage: raw, Valid: true}
}
