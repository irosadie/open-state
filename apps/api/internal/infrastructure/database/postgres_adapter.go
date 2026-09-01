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
	pool           *pgxpool.Pool
	sqlDB          *sql.DB
	queries        *db.Queries
	projects       repositories.IProjectRepository
	workflows      repositories.IWorkflowRepository
	instances      repositories.IInstanceRepository
	events         repositories.IEventRepository
	context        repositories.IContextRepository
	capabilities   repositories.ICapabilityRepository
	evidence       repositories.ICapabilityEvidenceRepository
	audit          repositories.IAuditRepository
	roles          repositories.IRoleAssignmentRepository
	identities     repositories.IUserIdentityRepository
	admin          repositories.IAdminRepository
	eventBrowser   repositories.IEventBrowserRepository
	runtimeRead    repositories.IRuntimeReadRepository
	traces         repositories.IRuntimeTraceRepository
	intents        repositories.IIntentRepository
	apiKeys        repositories.IAPIKeyRepository
	mcpConnections repositories.IMCPConnectionRepository
	mcpToolCatalog repositories.IMCPToolCatalogRepository
	mcpBindings   repositories.IProjectCapabilityMCPBindingRepository
}

// NewPostgresAdapter returns a PostgresAdapter composing all six pgx repositories.
func NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter {
	sqlDB := stdlib.OpenDBFromPool(pool)
	eventRepo := NewPgxEventRepository(pool)
	return &PostgresAdapter{
		pool:           pool,
		sqlDB:          sqlDB,
		queries:        db.New(sqlDB),
		projects:       NewPgxProjectRepository(pool),
		workflows:      NewPgxWorkflowRepository(pool),
		instances:      NewPgxInstanceRepository(pool),
		events:         eventRepo,
		context:        NewPgxContextRepository(pool),
		capabilities:   NewPgxCapabilityRepository(pool),
		evidence:       NewPgxCapabilityEvidenceRepository(pool),
		audit:          NewPgxAuditRepository(pool),
		roles:          NewPgxRoleAssignmentRepository(pool),
		identities:     NewPgxUserIdentityRepository(pool),
		admin:          NewPgxAdminRepository(pool),
		eventBrowser:   eventRepo,
		runtimeRead:    NewPgxRuntimeReadRepository(pool),
		traces:         NewPgxRuntimeTraceRepository(pool),
		intents:        NewPgxIntentRepository(pool),
		apiKeys:        NewPgxAPIKeyRepository(pool),
		mcpConnections: NewPgxMCPConnectionRepository(pool),
		mcpToolCatalog: NewPgxMCPToolCatalogRepository(pool),
		mcpBindings:    NewPgxProjectCapabilityMCPBindingRepository(pool),
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

// CapabilityEvidence returns State MCP's explicit provider execution evidence store.
func (a *PostgresAdapter) CapabilityEvidence() repositories.ICapabilityEvidenceRepository {
	return a.evidence
}

// Audit returns the append-only audit-trail repository.
func (a *PostgresAdapter) Audit() repositories.IAuditRepository { return a.audit }

// Roles returns the tenant-scoped RBAC role-assignment repository.
func (a *PostgresAdapter) Roles() repositories.IRoleAssignmentRepository { return a.roles }

// Identities returns the external OIDC identity repository (PRD §79).
func (a *PostgresAdapter) Identities() repositories.IUserIdentityRepository { return a.identities }

// Admin returns tenant profile and membership administration persistence.
func (a *PostgresAdapter) Admin() repositories.IAdminRepository { return a.admin }

// EventBrowser returns the read-only, paginated event query surface.
func (a *PostgresAdapter) EventBrowser() repositories.IEventBrowserRepository { return a.eventBrowser }

// RuntimeRead returns the tenant-scoped definition and state projections used
// by Runtime Inspector.
func (a *PostgresAdapter) RuntimeRead() repositories.IRuntimeReadRepository { return a.runtimeRead }

// RuntimeTraces returns the append-only trace repository.
func (a *PostgresAdapter) RuntimeTraces() repositories.IRuntimeTraceRepository { return a.traces }

// Intents returns the canonical intent catalog repository.
func (a *PostgresAdapter) Intents() repositories.IIntentRepository { return a.intents }

// APIKeys returns the State MCP machine credential repository.
func (a *PostgresAdapter) APIKeys() repositories.IAPIKeyRepository { return a.apiKeys }

// MCPConnections returns the project-scoped external MCP connection registry.
func (a *PostgresAdapter) MCPConnections() repositories.IMCPConnectionRepository {
	return a.mcpConnections
}

// MCPToolCatalog returns the project-scoped discovered MCP tool catalog.
func (a *PostgresAdapter) MCPToolCatalog() repositories.IMCPToolCatalogRepository {
	return a.mcpToolCatalog
}

// ProjectMCPBindings returns explicit project-scoped capability-to-tool mappings.
func (a *PostgresAdapter) ProjectMCPBindings() repositories.IProjectCapabilityMCPBindingRepository {
	return a.mcpBindings
}

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
		queries:        q,
		workflows:      newPgxWorkflowRepository(q, a.sqlDB),
		instances:      newPgxInstanceRepository(q, a.sqlDB),
		events:         newPgxEventRepository(q),
		context:        newPgxContextRepository(q),
		capabilities:   newPgxCapabilityRepository(q),
		evidence:       newPgxCapabilityEvidenceRepository(q),
		audit:          newPgxAuditRepository(q),
		roles:          newPgxRoleAssignmentRepository(q),
		identities:     newPgxUserIdentityRepository(q),
		admin:          newPgxAdminRepository(q, a.sqlDB),
		eventBrowser:   newPgxEventRepository(q),
		runtimeRead:    newPgxRuntimeReadRepository(q),
		traces:         newPgxRuntimeTraceRepository(q),
		intents:        newPgxIntentRepository(q),
		apiKeys:        newPgxAPIKeyRepository(q),
		mcpConnections: newPgxMCPConnectionRepository(q),
		mcpToolCatalog: newPgxMCPToolCatalogRepository(q, a.sqlDB),
		mcpBindings:    newPgxProjectCapabilityMCPBindingRepository(q),
	}

	if err := fn(txAdapter); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return domain.NewInternal(err.Error())
	}
	return nil
}
