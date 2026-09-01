package database

import (
	"context"
	"encoding/json"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// PgxIntentRepository implements IIntentRepository on PostgreSQL via sqlc.
type PgxIntentRepository struct {
	queries *db.Queries
}

// NewPgxIntentRepository returns a PostgreSQL-backed IIntentRepository.
func NewPgxIntentRepository(pool *pgxpool.Pool) repositories.IIntentRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxIntentRepository(db.New(sqlDB))
}

func newPgxIntentRepository(q *db.Queries) repositories.IIntentRepository {
	return &PgxIntentRepository{queries: q}
}

func (r *PgxIntentRepository) ListRoutable(ctx context.Context, tenantID, projectID string) ([]entities.Intent, error) {
	rows, err := r.queries.ListRoutableIntents(ctx, db.ListRoutableIntentsParams{
		TenantID:  mustUUID(tenantID),
		ProjectID: mustUUID(projectID),
	})
	if err != nil {
		return nil, mapPgError(err, "list intents")
	}

	intents := make([]entities.Intent, 0, len(rows))
	for _, row := range rows {
		intent, err := mapIntentListRow(row)
		if err != nil {
			return nil, err
		}
		intents = append(intents, *intent)
	}
	return intents, nil
}

func (r *PgxIntentRepository) FindRoutable(ctx context.Context, tenantID, projectID, key string) (*entities.Intent, error) {
	row, err := r.queries.FindRoutableIntent(ctx, db.FindRoutableIntentParams{
		TenantID:  mustUUID(tenantID),
		ProjectID: mustUUID(projectID),
		IntentKey: key,
	})
	if err != nil {
		return nil, mapNotFound(err, "intent")
	}
	return mapIntentRow(row)
}

func (r *PgxIntentRepository) Upsert(ctx context.Context, tenantID, projectID, workflowID, key, name, description string, examples []string) (*entities.Intent, error) {
	rawExamples, err := json.Marshal(examples)
	if err != nil {
		return nil, mapPgError(err, "encode intent examples")
	}
	row, err := r.queries.UpsertIntent(ctx, db.UpsertIntentParams{
		TenantID:    mustUUID(tenantID),
		ProjectID:   mustUUID(projectID),
		WorkflowID:  mustUUID(workflowID),
		IntentKey:   key,
		Name:        name,
		Description: description,
		Examples:    rawExamples,
	})
	if err != nil {
		return nil, mapPgError(err, "upsert intent")
	}
	return mapIntent(row)
}

func mapIntent(row db.Intent) (*entities.Intent, error) {
	var examples []string
	if err := json.Unmarshal(row.Examples, &examples); err != nil {
		return nil, mapPgError(err, "decode intent examples")
	}
	if examples == nil {
		examples = []string{}
	}
	return &entities.Intent{
		ID:          row.ID.String(),
		TenantID:    row.TenantID.String(),
		ProjectID:   row.ProjectID.String(),
		WorkflowID:  row.WorkflowID.String(),
		Key:         row.IntentKey,
		Name:        row.Name,
		Description: row.Description,
		Examples:    examples,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func mapIntentRow(row db.FindRoutableIntentRow) (*entities.Intent, error) {
	intent, err := mapIntent(db.Intent{
		ID:          row.ID,
		TenantID:    row.TenantID,
		ProjectID:   row.ProjectID,
		WorkflowID:  row.WorkflowID,
		IntentKey:   row.IntentKey,
		Name:        row.Name,
		Description: row.Description,
		Examples:    row.Examples,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	intent.WorkflowSlug = row.WorkflowSlug
	return intent, nil
}

func mapIntentListRow(row db.ListRoutableIntentsRow) (*entities.Intent, error) {
	intent, err := mapIntent(db.Intent{
		ID:          row.ID,
		TenantID:    row.TenantID,
		ProjectID:   row.ProjectID,
		WorkflowID:  row.WorkflowID,
		IntentKey:   row.IntentKey,
		Name:        row.Name,
		Description: row.Description,
		Examples:    row.Examples,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	intent.WorkflowSlug = row.WorkflowSlug
	return intent, nil
}
