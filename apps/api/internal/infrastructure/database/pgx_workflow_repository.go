package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

const pgUniqueViolation = "23505"

// PgxWorkflowRepository implements IWorkflowRepository on PostgreSQL via sqlc.
type PgxWorkflowRepository struct {
	queries *db.Queries
	db      *sql.DB
}

// NewPgxWorkflowRepository returns a PostgreSQL-backed IWorkflowRepository.
func NewPgxWorkflowRepository(pool *pgxpool.Pool) repositories.IWorkflowRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxWorkflowRepository(db.New(sqlDB), sqlDB)
}

// newPgxWorkflowRepository builds an IWorkflowRepository from a sqlc queries
// handle and the underlying *sql.DB (used for the repo's own standalone
// transactions). It enables the composed PostgresAdapter to bind this repository
// to a shared transaction via WithTx.
func newPgxWorkflowRepository(q *db.Queries, sqlDB *sql.DB) repositories.IWorkflowRepository {
	return &PgxWorkflowRepository{queries: q, db: sqlDB}
}

func (r *PgxWorkflowRepository) Create(ctx context.Context, tenantID, projectID, slug, name string, description *string, draftDefinition []byte) (*entities.Workflow, error) {
	row, err := r.queries.CreateWorkflow(ctx, db.CreateWorkflowParams{
		TenantID:        mustUUID(tenantID),
		ProjectID:       mustUUID(projectID),
		Slug:            slug,
		Name:            name,
		Description:     nullString(description),
		Status:          string(entities.WorkflowDraft),
		DraftDefinition: json.RawMessage(draftDefinition),
	})
	if err != nil {
		return nil, mapPgError(err, "create workflow")
	}
	return mapWorkflow(row), nil
}

func (r *PgxWorkflowRepository) UpdateDraft(ctx context.Context, tenantID, projectID, id, name string, description *string, draftDefinition []byte, expectedVersion int) (*entities.Workflow, error) {
	row, err := r.queries.UpdateWorkflowDraft(ctx, db.UpdateWorkflowDraftParams{
		ID:              mustUUID(id),
		TenantID:        mustUUID(tenantID),
		ProjectID:       mustUUID(projectID),
		Name:            name,
		Description:     nullString(description),
		DraftDefinition: json.RawMessage(draftDefinition),
		Version:         int32(expectedVersion),
	})
	if err != nil {
		return nil, mapOptimisticError(err)
	}
	return mapWorkflow(row), nil
}

func (r *PgxWorkflowRepository) FindByID(ctx context.Context, tenantID, projectID, id string) (*entities.Workflow, error) {
	row, err := r.queries.FindWorkflowByID(ctx, db.FindWorkflowByIDParams{
		ID:        mustUUID(id),
		TenantID:  mustUUID(tenantID),
		ProjectID: mustUUID(projectID),
	})
	if err != nil {
		return nil, mapNotFound(err, "workflow")
	}
	return mapWorkflow(row), nil
}

func (r *PgxWorkflowRepository) FindBySlug(ctx context.Context, tenantID, projectID, slug string) (*entities.Workflow, error) {
	row, err := r.queries.FindWorkflowBySlug(ctx, db.FindWorkflowBySlugParams{
		Slug:      slug,
		TenantID:  mustUUID(tenantID),
		ProjectID: mustUUID(projectID),
	})
	if err != nil {
		return nil, mapNotFound(err, "workflow")
	}
	return mapWorkflow(row), nil
}

func (r *PgxWorkflowRepository) ListByTenant(ctx context.Context, tenantID, projectID string) ([]entities.Workflow, error) {
	rows, err := r.queries.ListWorkflowsByTenant(ctx, db.ListWorkflowsByTenantParams{
		TenantID:  mustUUID(tenantID),
		ProjectID: mustUUID(projectID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.Workflow, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapWorkflow(row))
	}
	return out, nil
}

func (r *PgxWorkflowRepository) UpdateStatus(ctx context.Context, tenantID, projectID, id string, status entities.WorkflowStatus, expectedVersion int) (*entities.Workflow, error) {
	row, err := r.queries.UpdateWorkflowStatus(ctx, db.UpdateWorkflowStatusParams{
		ID:        mustUUID(id),
		TenantID:  mustUUID(tenantID),
		ProjectID: mustUUID(projectID),
		Status:    string(status),
		Version:   int32(expectedVersion),
	})
	if err != nil {
		return nil, mapOptimisticError(err)
	}
	return mapWorkflow(row), nil
}

func (r *PgxWorkflowRepository) CreateVersion(ctx context.Context, tenantID, projectID, workflowID string, versionNo int, definition []byte, status entities.VersionStatus, isCurrent bool) (*entities.WorkflowVersion, error) {
	row, err := r.queries.CreateWorkflowVersion(ctx, db.CreateWorkflowVersionParams{
		WorkflowID: mustUUID(workflowID),
		TenantID:   mustUUID(tenantID),
		ProjectID:  mustUUID(projectID),
		VersionNo:  int32(versionNo),
		Definition: definition,
		Status:     string(status),
		IsCurrent:  isCurrent,
	})
	if err != nil {
		return nil, mapPgError(err, "create workflow version")
	}
	return mapWorkflowVersion(row), nil
}

// Publish atomically inserts a new version, marks it current, and bumps the
// workflow's current_version + optimistic version counter (PRD §9, §65, §69).
func (r *PgxWorkflowRepository) Publish(ctx context.Context, tenantID, projectID, workflowID string, versionNo int, definition []byte, status entities.VersionStatus, expectedVersion int) (*entities.WorkflowVersion, error) {
	uid, err := uuid.Parse(workflowID)
	if err != nil {
		return nil, domain.NewValidation("invalid workflow id")
	}
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, domain.NewValidation("invalid tenant id")
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.NewValidation("invalid project id")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, domain.NewInternal(err.Error())
	}
	defer tx.Rollback() //nolint:errcheck // rollback is a no-op after commit

	q := r.queries.WithTx(tx)

	// Insert the new immutable version.
	if _, err := q.CreateWorkflowVersion(ctx, db.CreateWorkflowVersionParams{
		WorkflowID: uid,
		TenantID:   tid,
		ProjectID:  pid,
		VersionNo:  int32(versionNo),
		Definition: definition,
		Status:     string(status),
		IsCurrent:  true,
	}); err != nil {
		return nil, mapPgError(err, "create workflow version")
	}
	// Mark the new version current (and unset others).
	if err := q.SetCurrentWorkflowVersion(ctx, db.SetCurrentWorkflowVersionParams{
		WorkflowID: uid,
		VersionNo:  int32(versionNo),
		TenantID:   tid,
		ProjectID:  pid,
	}); err != nil {
		return nil, domain.NewInternal(err.Error())
	}
	// Bump the workflow root current_version + optimistic version counter.
	if _, err := q.UpdateWorkflowVersion(ctx, db.UpdateWorkflowVersionParams{
		ID:             uid,
		TenantID:       tid,
		ProjectID:      pid,
		CurrentVersion: int32(versionNo),
		Version:        int32(expectedVersion),
	}); err != nil {
		return nil, mapOptimisticError(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, domain.NewInternal(err.Error())
	}

	// Return the freshly inserted version (the "current" version).
	return r.FindVersionByNumber(ctx, tenantID, projectID, workflowID, versionNo)
}

func (r *PgxWorkflowRepository) FindCurrentVersion(ctx context.Context, tenantID, projectID, workflowID string) (*entities.WorkflowVersion, error) {
	row, err := r.queries.FindCurrentWorkflowVersion(ctx, db.FindCurrentWorkflowVersionParams{
		WorkflowID: mustUUID(workflowID),
		TenantID:   mustUUID(tenantID),
		ProjectID:  mustUUID(projectID),
	})
	if err != nil {
		return nil, mapNotFound(err, "workflow version")
	}
	return mapWorkflowVersion(row), nil
}

func (r *PgxWorkflowRepository) FindCurrentVersionByWorkflow(ctx context.Context, tenantID, workflowID string) (*entities.WorkflowVersion, error) {
	row, err := r.queries.FindCurrentWorkflowVersionByWorkflow(ctx, db.FindCurrentWorkflowVersionByWorkflowParams{
		WorkflowID: mustUUID(workflowID),
		TenantID:   mustUUID(tenantID),
	})
	if err != nil {
		return nil, mapNotFound(err, "workflow version")
	}
	return mapWorkflowVersion(row), nil
}

func (r *PgxWorkflowRepository) ListVersions(ctx context.Context, tenantID, projectID, workflowID string) ([]entities.WorkflowVersion, error) {
	rows, err := r.queries.ListWorkflowVersions(ctx, db.ListWorkflowVersionsParams{
		WorkflowID: mustUUID(workflowID),
		TenantID:   mustUUID(tenantID),
		ProjectID:  mustUUID(projectID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.WorkflowVersion, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapWorkflowVersion(row))
	}
	return out, nil
}

func (r *PgxWorkflowRepository) FindVersionByNumber(ctx context.Context, tenantID, projectID, workflowID string, versionNo int) (*entities.WorkflowVersion, error) {
	row, err := r.queries.FindWorkflowVersionByNumber(ctx, db.FindWorkflowVersionByNumberParams{
		WorkflowID: mustUUID(workflowID),
		VersionNo:  int32(versionNo),
		TenantID:   mustUUID(tenantID),
		ProjectID:  mustUUID(projectID),
	})
	if err != nil {
		return nil, mapNotFound(err, "workflow version")
	}
	return mapWorkflowVersion(row), nil
}

func (r *PgxWorkflowRepository) CreateState(ctx context.Context, tenantID, projectID, workflowVersionID, key string, kind entities.StateKind, name string, description, instructions *string, requiredContext, capabilities, policy, position []byte, isTerminal bool) (*entities.State, error) {
	row, err := r.queries.CreateState(ctx, db.CreateStateParams{
		WorkflowVersionID: mustUUID(workflowVersionID),
		Key:               key,
		Kind:              string(kind),
		Name:              name,
		Description:       nullString(description),
		Instructions:      nullString(instructions),
		RequiredContext:   requiredContext,
		Capabilities:      capabilities,
		Policy:            policy,
		IsTerminal:        isTerminal,
		Position:          position,
	})
	if err != nil {
		return nil, mapPgError(err, "create state")
	}
	return mapState(row), nil
}

func (r *PgxWorkflowRepository) ListStatesByVersion(ctx context.Context, tenantID, projectID, workflowVersionID string) ([]entities.State, error) {
	rows, err := r.queries.ListStatesByVersion(ctx, db.ListStatesByVersionParams{
		WorkflowVersionID: mustUUID(workflowVersionID),
		TenantID:          mustUUID(tenantID),
		ProjectID:         mustUUID(projectID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.State, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapState(row))
	}
	return out, nil
}

func (r *PgxWorkflowRepository) CreateTransition(ctx context.Context, tenantID, projectID, workflowVersionID, key, sourceStateID, targetStateID, event string, priority int, isActive bool) (*entities.Transition, error) {
	row, err := r.queries.CreateTransition(ctx, db.CreateTransitionParams{
		WorkflowVersionID: mustUUID(workflowVersionID),
		Key:               key,
		SourceStateID:     mustUUID(sourceStateID),
		TargetStateID:     mustUUID(targetStateID),
		Event:             event,
		Priority:          int32(priority),
		IsActive:          isActive,
	})
	if err != nil {
		return nil, mapPgError(err, "create transition")
	}
	return mapTransition(row), nil
}

func (r *PgxWorkflowRepository) ListTransitionsByVersion(ctx context.Context, tenantID, projectID, workflowVersionID string) ([]entities.Transition, error) {
	rows, err := r.queries.ListTransitionsByVersion(ctx, db.ListTransitionsByVersionParams{
		WorkflowVersionID: mustUUID(workflowVersionID),
		TenantID:          mustUUID(tenantID),
		ProjectID:         mustUUID(projectID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.Transition, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapTransition(row))
	}
	return out, nil
}

func (r *PgxWorkflowRepository) CreateTransitionGuard(ctx context.Context, tenantID, projectID, transitionID, workflowVersionID, logic string, conditions []byte) (*entities.TransitionGuard, error) {
	row, err := r.queries.CreateTransitionGuard(ctx, db.CreateTransitionGuardParams{
		TransitionID:      mustUUID(transitionID),
		WorkflowVersionID: mustUUID(workflowVersionID),
		Logic:             logic,
		Conditions:        conditions,
	})
	if err != nil {
		return nil, mapPgError(err, "create transition guard")
	}
	return mapTransitionGuard(row), nil
}

func (r *PgxWorkflowRepository) ListGuardsByTransition(ctx context.Context, tenantID, projectID, transitionID string) ([]entities.TransitionGuard, error) {
	rows, err := r.queries.ListGuardsByTransition(ctx, db.ListGuardsByTransitionParams{
		TransitionID: mustUUID(transitionID),
		TenantID:     mustUUID(tenantID),
		ProjectID:    mustUUID(projectID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.TransitionGuard, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapTransitionGuard(row))
	}
	return out, nil
}

// ---- mappers ----

func mapWorkflow(row db.Workflow) *entities.Workflow {
	return &entities.Workflow{
		ID:              row.ID.String(),
		TenantID:        row.TenantID.String(),
		ProjectID:       row.ProjectID.String(),
		Slug:            row.Slug,
		Name:            row.Name,
		Description:     row.Description,
		Status:          entities.WorkflowStatus(row.Status),
		CurrentVersion:  int(row.CurrentVersion),
		Version:         int(row.Version),
		DraftDefinition: row.DraftDefinition,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func mapWorkflowVersion(row db.WorkflowVersion) *entities.WorkflowVersion {
	return &entities.WorkflowVersion{
		ID:         row.ID.String(),
		WorkflowID: row.WorkflowID.String(),
		TenantID:   row.TenantID.String(),
		ProjectID:  row.ProjectID.String(),
		VersionNo:  int(row.VersionNo),
		Definition: row.Definition,
		Status:     entities.VersionStatus(row.Status),
		IsCurrent:  row.IsCurrent,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

func mapState(row db.State) *entities.State {
	return &entities.State{
		ID:                row.ID.String(),
		WorkflowVersionID: row.WorkflowVersionID.String(),
		Key:               row.Key,
		Kind:              entities.StateKind(row.Kind),
		Name:              row.Name,
		Description:       row.Description,
		Instructions:      row.Instructions,
		RequiredContext:   row.RequiredContext,
		Capabilities:      row.Capabilities,
		Policy:            row.Policy,
		IsTerminal:        row.IsTerminal,
		Position:          row.Position,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}

func mapTransition(row db.Transition) *entities.Transition {
	return &entities.Transition{
		ID:                row.ID.String(),
		WorkflowVersionID: row.WorkflowVersionID.String(),
		Key:               row.Key,
		SourceStateID:     row.SourceStateID.String(),
		TargetStateID:     row.TargetStateID.String(),
		Event:             row.Event,
		Priority:          int(row.Priority),
		IsActive:          row.IsActive,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}

func mapTransitionGuard(row db.TransitionGuard) *entities.TransitionGuard {
	return &entities.TransitionGuard{
		ID:                row.ID.String(),
		TransitionID:      row.TransitionID.String(),
		WorkflowVersionID: row.WorkflowVersionID.String(),
		Logic:             row.Logic,
		Conditions:        row.Conditions,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}

// ---- helpers ----

func mustUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func mapPgError(err error, op string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return domain.NewConflict(op + ": already exists")
	}
	return domain.NewInternal(err.Error())
}

func mapNotFound(err error, resource string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NewNotFound(resource + " not found")
	}
	return domain.NewInternal(err.Error())
}

func mapOptimisticError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NewConflict("optimistic lock conflict: resource changed")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return domain.NewConflict("duplicate resource")
	}
	msg := err.Error()
	if strings.Contains(msg, "no rows") || strings.Contains(msg, sql.ErrNoRows.Error()) {
		return domain.NewConflict("optimistic lock conflict: resource changed")
	}
	return domain.NewInternal(err.Error())
}
