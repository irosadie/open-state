// Package engineadapter adapts the persistence repositories (entities.*) to the
// domain engine ports (engine.EngineRepositories), so the runtime engine can be
// wired into the production MCP/application path (PRD 170, ADR-001). It is the
// only place that maps between entities.* and engine.* domain models.
package engineadapter

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// Adapter composes the persistence repositories and the sqlc query handle, and
// exposes the engine's repository ports via per-port structs. It is the
// persistence→engine seam (ADR-001).
type Adapter struct {
	projects  repositories.IProjectRepository
	workflows repositories.IWorkflowRepository
	instances repositories.IInstanceRepository
	events    repositories.IEventRepository
	q         *db.Queries
}

// New builds an Adapter over a pgx pool and the composed persistence repositories.
func New(pool *pgxpool.Pool, projects repositories.IProjectRepository, workflows repositories.IWorkflowRepository, instances repositories.IInstanceRepository, events repositories.IEventRepository) *Adapter {
	return &Adapter{
		projects:  projects,
		workflows: workflows,
		instances: instances,
		events:    events,
		q:         db.New(stdlib.OpenDBFromPool(pool)),
	}
}

// Repos returns the adapter in the engine.EngineRepositories shape.
func (a *Adapter) Repos() engine.EngineRepositories {
	return engine.EngineRepositories{
		Projects:  &projectRepo{a: a},
		Workflows: &workflowRepo{a: a},
		Instances: &instanceRepo{a: a},
		Events:    &eventRepo{a: a},
	}
}

// ---- ProjectRepository (engine port) ----

type projectRepo struct{ a *Adapter }

func (r *projectRepo) Get(ctx context.Context, tenantID, projectID string) (*engine.Project, error) {
	p, err := r.a.projects.FindByID(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	return &engine.Project{ID: p.ID, TenantID: p.TenantID, Name: p.Name, Slug: p.Slug, Status: string(p.Status)}, nil
}

func (r *projectRepo) Save(_ context.Context, _ *engine.Project) error {
	// Projects are seeded; the engine does not create projects at runtime.
	return nil
}

// ---- WorkflowRepository (engine port) ----

type workflowRepo struct{ a *Adapter }

func (r *workflowRepo) GetBySlug(ctx context.Context, tenantID, projectID, slug string) (*engine.WorkflowDefinition, error) {
	wf, err := r.a.workflows.FindBySlug(ctx, tenantID, projectID, slug)
	if err != nil {
		return nil, err
	}
	version, err := r.a.workflows.FindCurrentVersion(ctx, tenantID, projectID, wf.ID)
	if err != nil {
		return nil, err
	}
	return definitionFromJSON(version.Definition)
}

func (r *workflowRepo) Save(_ context.Context, _ *engine.WorkflowDefinition) error {
	// Definitions are authored/published via the Builder API and seed; the engine
	// does not persist definitions at runtime.
	return nil
}

// definitionFromJSON unmarshals a persisted WorkflowVersion.Definition (engine
// format, PRD §68) into an engine.WorkflowDefinition.
func definitionFromJSON(raw json.RawMessage) (*engine.WorkflowDefinition, error) {
	if len(raw) == 0 {
		return nil, domain.NewNotFound("workflow definition is empty")
	}
	var def engine.WorkflowDefinition
	if err := json.Unmarshal(raw, &def); err != nil {
		return nil, domain.NewInternal("invalid workflow definition: " + err.Error())
	}
	return &def, nil
}

// ---- InstanceRepository (engine port) ----

type instanceRepo struct{ a *Adapter }

func (r *instanceRepo) Create(ctx context.Context, inst *engine.WorkflowInstance) error {
	_, err := r.a.instances.Create(ctx, inst.TenantID, repositories.CreateWorkflowInstanceInput{
		WorkflowID:        inst.WorkflowID,
		WorkflowVersionID: inst.WorkflowVersionID,
		CorrelationKey:    strPtr(inst.ConversationID),
		StartedAt:         &inst.CreatedAt,
	})
	return err
}

func (r *instanceRepo) Get(ctx context.Context, tenantID, instanceID string) (*engine.WorkflowInstance, error) {
	inst, err := r.a.instances.FindByID(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	return r.a.resolveEngineInstance(ctx, tenantID, inst)
}

// resolveEngineInstance maps a persisted instance to an engine instance, resolving
// the project id + workflow slug (from the pinned version) and the current state
// key (from the current state instance), which the engine needs to load the
// definition and evaluate transitions.
func (a *Adapter) resolveEngineInstance(ctx context.Context, tenantID string, e *entities.WorkflowInstance) (*engine.WorkflowInstance, error) {
	out := toEngineInstance(e)

	// Resolve project id + workflow slug from the pinned workflow version.
	row, err := a.q.FindWorkflowVersionByID(ctx, db.FindWorkflowVersionByIDParams{
		ID:       mustUUID(e.WorkflowVersionID),
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return nil, domain.NewInternal("resolve workflow version: " + err.Error())
	}
	out.ProjectID = row.ProjectID.String()
	out.WorkflowID = row.WorkflowSlug

	// Resolve the current state key from the current state instance.
	if e.CurrentStateInstanceID != nil && *e.CurrentStateInstanceID != "" {
		stateRow, err := a.q.FindStateInstanceByID(ctx, db.FindStateInstanceByIDParams{
			ID:       mustUUID(*e.CurrentStateInstanceID),
			TenantID: mustUUID(tenantID),
		})
		if err == nil {
			out.CurrentStateID = stateRow.StateKey
		}
	}
	return out, nil
}

// UpdateWithVersion persists the engine-computed current state of an instance and
// bumps its optimistic version. It creates a state instance for the new current
// state, repoints the workflow instance's current_state_instance_id, and bumps the
// version with optimistic locking (PRD §69).
func (r *instanceRepo) UpdateWithVersion(ctx context.Context, inst *engine.WorkflowInstance, expectedVersion int) error {
	state, err := r.a.instances.CreateStateInstance(ctx, inst.TenantID, repositories.CreateStateInstanceInput{
		WorkflowInstanceID: inst.ID,
		WorkflowVersionID:  inst.WorkflowVersionID,
		StateKey:           inst.CurrentStateID,
	})
	if err != nil {
		return err
	}

	if err := r.a.q.SetCurrentStateInstance(ctx, db.SetCurrentStateInstanceParams{
		ID:                     mustUUID(inst.ID),
		TenantID:               mustUUID(inst.TenantID),
		CurrentStateInstanceID: uuid.NullUUID{UUID: mustUUID(state.ID), Valid: true},
	}); err != nil {
		return domain.NewInternal(err.Error())
	}

	if _, err := r.a.q.IncrementWorkflowInstanceVersion(ctx, db.IncrementWorkflowInstanceVersionParams{
		ID:       mustUUID(inst.ID),
		TenantID: mustUUID(inst.TenantID),
		Version:  int32(expectedVersion),
	}); err != nil {
		return mapVersionErr(err)
	}
	return nil
}

// ---- EventRepository (engine port) ----

type eventRepo struct{ a *Adapter }

func (r *eventRepo) Append(ctx context.Context, evt *engine.Event) error {
	payload := evt.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return domain.NewValidation("invalid event payload")
	}
	var idem *string
	if evt.IdempotencyKey != "" {
		idem = &evt.IdempotencyKey
	}
	instID := evt.WorkflowInstanceID
	_, err = r.a.events.Append(ctx, evt.TenantID, repositories.AppendEventInput{
		EventID:            evt.ID,
		Type:               evt.Type,
		Source:             toEntitiesSource(evt.Source),
		WorkflowInstanceID: &instID,
		Timestamp:          evt.Timestamp,
		Payload:            raw,
		IdempotencyKey:     idem,
	})
	return err
}

func (r *eventRepo) IsProcessed(ctx context.Context, tenantID, idempotencyKey string) (bool, error) {
	_, err := r.a.events.FindIdempotency(ctx, tenantID, idempotencyKey)
	if err != nil {
		var de *domain.DomainError
		if errors.As(err, &de) && de.Code == domain.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *eventRepo) MarkProcessed(ctx context.Context, tenantID, idempotencyKey, eventID string) error {
	_, err := r.a.events.UpsertIdempotency(ctx, tenantID, repositories.UpsertIdempotencyInput{
		IdempotencyKey: idempotencyKey,
		Scope:          eventID,
		ResultStatus:   entities.IdempotencyResultProcessed,
	})
	return err
}

// ---- mapping helpers ----

func toEngineInstance(e *entities.WorkflowInstance) *engine.WorkflowInstance {
	out := &engine.WorkflowInstance{
		ID:                e.ID,
		TenantID:          e.TenantID,
		WorkflowID:        e.WorkflowID,
		WorkflowVersionID: e.WorkflowVersionID,
		Status:            engine.WorkflowInstanceStatus(e.Status),
		Version:           e.Version,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
	if e.CurrentStateInstanceID != nil && *e.CurrentStateInstanceID != "" {
		out.CurrentStateID = *e.CurrentStateInstanceID
	}
	if e.CorrelationKey.Valid {
		out.ConversationID = e.CorrelationKey.String
	}
	return out
}

func toEntitiesSource(s engine.EventSource) entities.EventSource { return entities.EventSource(s) }

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func mustUUID(s string) uuid.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return u
}

// mapVersionErr maps a NOT_FOUND (from IncrementWorkflowInstanceVersion when the
// expected version does not match) to a CONFLICT, mirroring the persistence repos.
func mapVersionErr(err error) error {
	var de *domain.DomainError
	if errors.As(err, &de) && de.Code == domain.ErrNotFound {
		return domain.NewConflict("optimistic lock conflict: resource changed")
	}
	return domain.NewInternal(err.Error())
}
