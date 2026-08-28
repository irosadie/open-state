package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

// defaultAuditPageSize and maxAuditPageSize bound audit listing page sizes to
// avoid unbounded response payloads (PRD 50).
const (
	defaultAuditPageSize = 20
	maxAuditPageSize     = 100
)

// AuditQuery carries the optional filters and pagination for listing the audit
// trail.
type AuditQuery struct {
	Action        *entities.AuditAction
	ResourceType  *string
	ResourceID    *string
	Actor         *string
	CorrelationID *string
	From          *time.Time
	To            *time.Time
	Page          int
	PageSize      int
}

// AuditService provides tenant-scoped read access to the append-only audit trail
// (PRD 50). It is consumed by the audit query API (GET /api/audit).
type AuditService struct {
	repo repositories.IAuditRepository
}

// NewAuditService builds an AuditService over the audit repository.
func NewAuditService(repo repositories.IAuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// List returns a paginated, filtered view of the tenant's audit trail (PRD 50).
// Filters are optional; an empty query returns all entries newest first.
func (s *AuditService) List(ctx context.Context, tenantID string, q AuditQuery) (*dtos.AuditPageDTO, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = defaultAuditPageSize
	}
	if pageSize > maxAuditPageSize {
		pageSize = maxAuditPageSize
	}

	filter := repositories.AuditFilter{
		Action:        q.Action,
		ResourceType:  q.ResourceType,
		ResourceID:    q.ResourceID,
		Actor:         q.Actor,
		CorrelationID: q.CorrelationID,
		From:          q.From,
		To:            q.To,
		Offset:        (page - 1) * pageSize,
		Limit:         pageSize,
	}

	total, err := s.repo.CountFiltered(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}

	entries, err := s.repo.ListFiltered(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}

	data := make([]dtos.AuditEntryDTO, 0, len(entries))
	for i := range entries {
		data = append(data, *toAuditEntryDTO(&entries[i]))
	}

	return &dtos.AuditPageDTO{
		Data:     data,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasNext:  int64(page*pageSize) < total,
	}, nil
}

func toAuditEntryDTO(a *entities.AuditLog) *dtos.AuditEntryDTO {
	dto := &dtos.AuditEntryDTO{
		ID:           a.ID,
		TenantID:     a.TenantID,
		Actor:        a.Actor,
		Action:       string(a.Action),
		ResourceType: a.ResourceType,
		ResourceID:   a.ResourceID,
		OccurredAt:   a.OccurredAt.Format(time.RFC3339),
	}
	dto.Before = rawToAnyMap(a.Before)
	dto.After = rawToAnyMap(a.After)
	if a.CorrelationID != nil {
		dto.CorrelationID = a.CorrelationID
	}
	return dto
}

// rawToAnyMap converts an optional JSON raw message into a generic map for the
// DTO, returning nil when absent or malformed.
func rawToAnyMap(raw *json.RawMessage) *map[string]any {
	if raw == nil || len(*raw) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(*raw, &m); err != nil {
		return nil
	}
	return &m
}
