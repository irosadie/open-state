package database

import (
	"context"
	"database/sql"

	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel"
)

// PostgresAdapter is the composed PostgreSQL-backed persistence port (ADR-001).
// It owns the pgx pool and the sqlc queries, composes the six repository
// interfaces (workflow, instance, event, context, capability, audit) under a
// single port, and is the ONLY composition point importing pgx/sqlc — the
// portability seam for future MySQL/SQLite/Mongo adapters. It exposes typed
// getters and a WithTx helper so multi-repository operations run atomically
// (PRD 65, 69). It contains no business logic.
type PostgresAdapter struct {
	pool         *pgxpool.Pool
	sqlDB        *sql.DB
	queries      *db.Queries
	projects     repositories.IProjectRepository
	workflows    repositories.IWorkflowRepository
	instances    repositories.IInstanceRepository
	events       repositories.IEventRepository
	context      repositories.IContextRepository
	capabilities repositories.ICapabilityRepository
	audit        repositories.IAuditRepository
	roles        repositories.IRoleAssignmentRepository
	identities   repositories.IUserIdentityRepository
}

// NewPostgresAdapter returns a PostgresAdapter composing all six pgx repositories.
func NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return &PostgresAdapter{
		pool:         pool,
		sqlDB:        sqlDB,
		queries:      db.New(sqlDB),
		projects:     NewPgxProjectRepository(pool),
		workflows:    NewPgxWorkflowRepository(pool),
		instances:    NewPgxInstanceRepository(pool),
		events:       NewPgxEventRepository(pool),
		context:      NewPgxContextRepository(pool),
		capabilities: NewPgxCapabilityRepository(pool),
		audit:        NewPgxAuditRepository(pool),
		roles:        NewPgxRoleAssignmentRepository(pool),
		identities:   NewPgxUserIdentityRepository(pool),
	}
}

// Projects returns the project repository (business areas, PRD §3.1.1).
func (a *PostgresAdapter) Projects() repositories.IProjectRepository { return a.projects }

// Workflows returns the workflow-definition repository.
func (a *PostgresAdapter) Workflows() repositories.IWorkflowRepository { return a.workflows }

// Instances returns the runtime workflow/state-instance repository.
func (a *PostgresAdapter) Instances() repositories.IInstanceRepository { return a.instances }

// Events returns the event system repository (history/inbox/outbox/idempotency).
func (a *PostgresAdapter) Events() repositories.IEventRepository { return a.events }

// Context returns the scoped context + memory repository.
func (a *PostgresAdapter) Context() repositories.IContextRepository { return a.context }

// Capabilities returns the capability registry + policy repository.
func (a *PostgresAdapter) Capabilities() repositories.ICapabilityRepository { return a.capabilities }

// Audit returns the append-only audit-trail repository.
func (a *PostgresAdapter) Audit() repositories.IAuditRepository { return a.audit }

// Roles returns the tenant-scoped RBAC role-assignment repository.
func (a *PostgresAdapter) Roles() repositories.IRoleAssignmentRepository { return a.roles }

// Identities returns the external OIDC identity repository (PRD §79).
func (a *PostgresAdapter) Identities() repositories.IUserIdentityRepository { return a.identities }

// WithTx runs fn within a single DB transaction, binding all six repositories to
// that transaction so multi-repository operations (e.g. append an audit entry and
// emit an outbox event, or a state transition) commit or roll back together
// (PRD 65, 69). If fn returns an error the transaction is rolled back; otherwise
// it is committed.
func (a *PostgresAdapter) WithTx(ctx context.Context, fn func(adapter *PostgresAdapter) error) error {
	// DB span under the active trace context (PRD §84). No-op without a global
	// TracerProvider.
	_, span := otel.Tracer("openstate.db").Start(ctx, "DB WithTx")
	defer span.End()

	tx, err := a.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return domain.NewInternal(err.Error())
	}
	defer tx.Rollback() //nolint:errcheck // rollback is a no-op after commit

	q := a.queries.WithTx(tx)
	txAdapter := &PostgresAdapter{
		queries:      q,
		workflows:    newPgxWorkflowRepository(q, a.sqlDB),
		instances:    newPgxInstanceRepository(q, a.sqlDB),
		events:       newPgxEventRepository(q),
		context:      newPgxContextRepository(q),
		capabilities: newPgxCapabilityRepository(q),
		audit:        newPgxAuditRepository(q),
		roles:        newPgxRoleAssignmentRepository(q),
		identities:   newPgxUserIdentityRepository(q),
	}

	if err := fn(txAdapter); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return domain.NewInternal(err.Error())
	}
	return nil
}
