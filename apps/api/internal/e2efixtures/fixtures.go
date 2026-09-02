package e2efixtures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Fixed synthetic identifiers make browser assertions stable while keeping
// every fixture isolated from the normal demo tenant.
const (
	TenantA = "00000000-0000-0000-0000-0000000000a1"
	TenantB = "00000000-0000-0000-0000-0000000000b1"

	EditorUser   = "10000000-0000-0000-0000-000000000001"
	OperatorUser = "10000000-0000-0000-0000-000000000002"
	ViewerUser   = "10000000-0000-0000-0000-000000000003"
	SentinelUser = "10000000-0000-0000-0000-000000000004"

	ProjectID          = "20000000-0000-0000-0000-000000000001"
	BuilderWorkflowID  = "30000000-0000-0000-0000-000000000001"
	BuilderIntentID    = "60000000-0000-0000-0000-000000000001"
	InvalidWorkflowID  = "30000000-0000-0000-0000-000000000002"
	StaleWorkflowID    = "30000000-0000-0000-0000-000000000003"
	RuntimeWorkflowID  = "30000000-0000-0000-0000-000000000004"
	SentinelWorkflowID = "30000000-0000-0000-0000-000000000005"

	BuilderVersionOneID = "40000000-0000-0000-0000-000000000001"
	BuilderVersionTwoID = "40000000-0000-0000-0000-000000000002"
	RuntimeVersionID    = "40000000-0000-0000-0000-000000000004"
	SentinelVersionID   = "40000000-0000-0000-0000-000000000005"

	RunningInstanceID        = "50000000-0000-0000-0000-000000000001"
	SuspendedInstanceID      = "50000000-0000-0000-0000-000000000002"
	FailedInstanceID         = "50000000-0000-0000-0000-000000000003"
	SentinelInstanceID       = "50000000-0000-0000-0000-000000000004"
	RunningStateInstanceID   = "51000000-0000-0000-0000-000000000001"
	SuspendedStateInstanceID = "51000000-0000-0000-0000-000000000002"
	FailedStateInstanceID    = "51000000-0000-0000-0000-000000000003"
	SentinelStateInstanceID  = "51000000-0000-0000-0000-000000000004"
)

const (
	startStateV1        = "41000000-0000-0000-0000-000000000001"
	intakeStateV1       = "41000000-0000-0000-0000-000000000002"
	endStateV1          = "41000000-0000-0000-0000-000000000003"
	startStateV2        = "42000000-0000-0000-0000-000000000001"
	intakeStateV2       = "42000000-0000-0000-0000-000000000002"
	reviewStateV2       = "42000000-0000-0000-0000-000000000003"
	endStateV2          = "42000000-0000-0000-0000-000000000004"
	runtimeStartState   = "45000000-0000-0000-0000-000000000001"
	runtimeIntakeState  = "45000000-0000-0000-0000-000000000002"
	runtimeEndState     = "45000000-0000-0000-0000-000000000003"
	sentinelStartState  = "46000000-0000-0000-0000-000000000001"
	sentinelIntakeState = "46000000-0000-0000-0000-000000000002"
	sentinelEndState    = "46000000-0000-0000-0000-000000000003"

	startIntakeTransitionV1  = "43000000-0000-0000-0000-000000000001"
	intakeEndTransitionV1    = "43000000-0000-0000-0000-000000000002"
	startIntakeTransitionV2  = "44000000-0000-0000-0000-000000000001"
	intakeReviewTransitionV2 = "44000000-0000-0000-0000-000000000002"
	reviewEndTransitionV2    = "44000000-0000-0000-0000-000000000003"
)

// IDs exposes the values needed by browser manifests and verification helpers.
type IDs struct {
	TenantA           string
	TenantB           string
	EditorUser        string
	OperatorUser      string
	ViewerUser        string
	SentinelUser      string
	BuilderWorkflow   string
	InvalidWorkflow   string
	StaleWorkflow     string
	RunningInstance   string
	SuspendedInstance string
	FailedInstance    string
	SentinelInstance  string
}

func FixtureIDs() IDs {
	return IDs{
		TenantA:           TenantA,
		TenantB:           TenantB,
		EditorUser:        EditorUser,
		OperatorUser:      OperatorUser,
		ViewerUser:        ViewerUser,
		SentinelUser:      SentinelUser,
		BuilderWorkflow:   BuilderWorkflowID,
		InvalidWorkflow:   InvalidWorkflowID,
		StaleWorkflow:     StaleWorkflowID,
		RunningInstance:   RunningInstanceID,
		SuspendedInstance: SuspendedInstanceID,
		FailedInstance:    FailedInstanceID,
		SentinelInstance:  SentinelInstanceID,
	}
}

type Verification struct {
	Mode   string   `json:"mode"`
	Checks []string `json:"checks"`
}

// Seed resets only the disposable E2E database contents and recreates all
// synthetic identities, graphs, runtime records, and audit expectations.
func Seed(ctx context.Context, pool *pgxpool.Pool, password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("E2E_FIXTURE_PASSWORD is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash fixture password: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin fixture reset: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback is a no-op after commit

	if _, err := tx.Exec(ctx, `
		TRUNCATE TABLE
			auth_sessions, audit_logs, context_records, memory_references,
			events, event_inbox, event_outbox, idempotency_records,
			runtime_trace_entries, state_instances, transition_guards,
			transitions, states, workflow_versions, workflow_instances,
			projects, workflows, role_assignments, users
		RESTART IDENTITY CASCADE
	`); err != nil {
		return fmt.Errorf("reset fixture tables: %w", err)
	}

	if err := insertUsers(ctx, tx, string(hash)); err != nil {
		return err
	}
	if err := insertProjectsAndWorkflows(ctx, tx); err != nil {
		return err
	}
	if err := insertBuilderVersions(ctx, tx); err != nil {
		return err
	}
	if err := insertIntentFixtures(ctx, tx); err != nil {
		return err
	}
	if err := insertRuntimeFixtures(ctx, tx); err != nil {
		return err
	}
	if err := insertAuditExpectations(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit fixture seed: %w", err)
	}
	return nil
}

func insertIntentFixtures(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO intents
			(id, tenant_id, project_id, workflow_id, intent_key, name, description, examples)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, 'BOOKING_PADEL',
			'Booking Padel', 'Routes a padel court booking request to the Builder workflow.',
			$5::jsonb)
	`, BuilderIntentID, TenantA, ProjectID, BuilderWorkflowID, `[
		"saya mau order lapangan",
		"saya mau booking lapangan padel"
	]`); err != nil {
		return fmt.Errorf("insert fixture intent: %w", err)
	}
	return nil
}

func insertUsers(ctx context.Context, tx pgx.Tx, passwordHash string) error {
	users := []struct {
		id, email, name, role, tenant string
	}{
		{EditorUser, "editor.golden@tenant-a.invalid", "Golden Editor", "EDITOR", TenantA},
		{OperatorUser, "operator.golden@tenant-a.invalid", "Golden Operator", "OPERATOR", TenantA},
		{ViewerUser, "viewer.golden@tenant-a.invalid", "Golden Viewer", "VIEWER", TenantA},
		{SentinelUser, "sentinel.golden@tenant-b.invalid", "Sentinel Viewer", "VIEWER", TenantB},
	}

	for _, user := range users {
		if _, err := tx.Exec(ctx, `
			INSERT INTO users (id, email, password_hash, name, role, status)
			VALUES ($1::uuid, $2, $3, $4, 'USER', 'ACTIVE')
		`, user.id, user.email, passwordHash, user.name); err != nil {
			return fmt.Errorf("insert fixture user %s: %w", user.email, err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO role_assignments (user_id, tenant_id, role)
			VALUES ($1::uuid, $2::uuid, $3)
		`, user.id, user.tenant, user.role); err != nil {
			return fmt.Errorf("assign fixture role %s: %w", user.role, err)
		}
	}
	return nil
}

func insertProjectsAndWorkflows(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO projects (id, tenant_id, name, slug, status)
		VALUES ($1::uuid, $2::uuid, 'Golden Tenant Project', 'default', 'ACTIVE')
	`, ProjectID, TenantA); err != nil {
		return fmt.Errorf("insert fixture project: %w", err)
	}

	validV1 := definitionJSON("golden-builder", "Golden Builder", builderNodesV1(), builderTransitionsV1())
	validV2 := definitionJSON("golden-builder", "Golden Builder", builderNodesV2(), builderTransitionsV2())
	invalid := []byte(`{"slug":"golden-invalid","name":"Invalid Golden Graph","schemaVersion":1,"status":"DRAFT","nodes":[{"id":"invalid-state","kind":"STATE","name":"ORPHAN","requiredContext":[],"capabilities":[],"policy":{},"position":{"x":0,"y":0}}],"transitions":[],"policy":{"interruptible":"USER_REQUESTED","priority":1},"triggers":[]}`)

	workflows := []struct {
		id, slug, name, status string
		current, version       int
		draft                  []byte
		tenant                 string
	}{
		{BuilderWorkflowID, "golden-builder", "Golden Builder", "PUBLISHED", 2, 0, validV2, TenantA},
		{InvalidWorkflowID, "golden-invalid", "Invalid Golden Graph", "DRAFT", 0, 0, invalid, TenantA},
		{StaleWorkflowID, "golden-stale", "Stale Golden Graph", "DRAFT", 1, 0, validV1, TenantA},
		{RuntimeWorkflowID, "golden-runtime", "Golden Runtime", "PUBLISHED", 1, 0, validV1, TenantA},
		{SentinelWorkflowID, "sentinel-only", "Sentinel Only", "PUBLISHED", 1, 0, validV1, TenantB},
	}

	for _, workflow := range workflows {
		if _, err := tx.Exec(ctx, `
			INSERT INTO workflows
				(id, tenant_id, project_id, slug, name, status, current_version, version, draft_definition)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, $9::jsonb)
		`, workflow.id, workflow.tenant, ProjectID, workflow.slug, workflow.name, workflow.status, workflow.current, workflow.version, workflow.draft); err != nil {
			return fmt.Errorf("insert fixture workflow %s: %w", workflow.slug, err)
		}
	}
	return nil
}

func insertBuilderVersions(ctx context.Context, tx pgx.Tx) error {
	v1 := definitionJSON("golden-builder", "Golden Builder", builderNodesV1(), builderTransitionsV1())
	v2 := definitionJSON("golden-builder", "Golden Builder", builderNodesV2(), builderTransitionsV2())
	versions := []struct {
		id, definition string
		versionNo      int
		current        bool
	}{
		{BuilderVersionOneID, string(v1), 1, false},
		{BuilderVersionTwoID, string(v2), 2, true},
	}

	for _, version := range versions {
		if _, err := tx.Exec(ctx, `
			INSERT INTO workflow_versions
				(id, workflow_id, tenant_id, project_id, version_no, definition, status, is_current)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6::jsonb, 'PUBLISHED', $7)
		`, version.id, BuilderWorkflowID, TenantA, ProjectID, version.versionNo, version.definition, version.current); err != nil {
			return fmt.Errorf("insert builder version %d: %w", version.versionNo, err)
		}
	}

	if err := insertGraphProjection(ctx, tx, BuilderVersionOneID, builderNodesV1(), builderTransitionsV1()); err != nil {
		return err
	}
	return insertGraphProjection(ctx, tx, BuilderVersionTwoID, builderNodesV2(), builderTransitionsV2())
}

func insertGraphProjection(ctx context.Context, tx pgx.Tx, versionID string, nodes []fixtureNode, transitions []fixtureTransition) error {
	for _, node := range nodes {
		position, err := json.Marshal(node.Position)
		if err != nil {
			return fmt.Errorf("marshal fixture node position: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO states
				(id, workflow_version_id, key, kind, name, description, instructions,
				required_context, capabilities, policy, is_terminal, position)
			VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, '', '[]'::jsonb, '[]'::jsonb, '{}'::jsonb, $7, $8::jsonb)
		`, node.ID, versionID, node.ID, node.Kind, node.Name, node.Description, node.Terminal, position); err != nil {
			return fmt.Errorf("insert fixture state %s: %w", node.ID, err)
		}
	}
	for _, transition := range transitions {
		if _, err := tx.Exec(ctx, `
			INSERT INTO transitions
				(id, workflow_version_id, key, source_state_id, target_state_id, event, priority)
			VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5::uuid, $6, $7)
		`, transition.ID, versionID, transition.ID, transition.Source, transition.Target, transition.Event, transition.Priority); err != nil {
			return fmt.Errorf("insert fixture transition %s: %w", transition.ID, err)
		}
	}
	return nil
}

func insertRuntimeFixtures(ctx context.Context, tx pgx.Tx) error {
	for _, version := range []struct {
		id, workflow, tenant, slug, name string
		nodes                            []fixtureNode
		transitions                      []fixtureTransition
	}{
		{RuntimeVersionID, RuntimeWorkflowID, TenantA, "golden-runtime", "Golden Runtime", runtimeNodesV1(), runtimeTransitionsV1()},
		{SentinelVersionID, SentinelWorkflowID, TenantB, "sentinel-only", "Sentinel Only", sentinelNodesV1(), sentinelTransitionsV1()},
	} {
		runtimeDefinition := string(definitionJSON(version.slug, version.name, version.nodes, version.transitions))
		if _, err := tx.Exec(ctx, `
			INSERT INTO workflow_versions
				(id, workflow_id, tenant_id, project_id, version_no, definition, status, is_current)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, 1, $5::jsonb, 'PUBLISHED', true)
		`, version.id, version.workflow, version.tenant, ProjectID, runtimeDefinition); err != nil {
			return fmt.Errorf("insert runtime workflow version: %w", err)
		}
		if err := insertGraphProjection(ctx, tx, version.id, version.nodes, version.transitions); err != nil {
			return err
		}
	}

	instances := []struct {
		id, status, correlation, stateID, stateKey, stateInstanceID string
	}{
		{RunningInstanceID, "RUNNING", "golden-running", runtimeStartState, runtimeStartState, RunningStateInstanceID},
		{SuspendedInstanceID, "SUSPENDED", "golden-suspended", runtimeIntakeState, runtimeIntakeState, SuspendedStateInstanceID},
		{FailedInstanceID, "FAILED", "golden-failed", runtimeEndState, runtimeEndState, FailedStateInstanceID},
		{SentinelInstanceID, "RUNNING", "sentinel-only", sentinelStartState, sentinelStartState, SentinelStateInstanceID},
	}
	for _, instance := range instances {
		workflowTenant := TenantA
		workflowID := RuntimeWorkflowID
		workflowVersionID := RuntimeVersionID
		if instance.id == SentinelInstanceID {
			workflowTenant = TenantB
			workflowID = SentinelWorkflowID
			workflowVersionID = SentinelVersionID
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO workflow_instances
				(id, tenant_id, workflow_id, workflow_version_id, correlation_key, status, version, started_at)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, 0, NOW())
		`, instance.id, workflowTenant, workflowID, workflowVersionID, instance.correlation, instance.status); err != nil {
			return fmt.Errorf("insert fixture instance %s: %w", instance.correlation, err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO state_instances
				(id, tenant_id, workflow_instance_id, workflow_version_id, state_key, state_id, status)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6::uuid, 'ACTIVE')
		`, instance.stateInstanceID, workflowTenant, instance.id, workflowVersionID, instance.stateKey, instance.stateID); err != nil {
			return fmt.Errorf("insert fixture state instance %s: %w", instance.correlation, err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE workflow_instances
			SET current_state_instance_id = $1::uuid
			WHERE id = $2::uuid AND tenant_id = $3::uuid
		`, instance.stateInstanceID, instance.id, workflowTenant); err != nil {
			return fmt.Errorf("point fixture current state %s: %w", instance.correlation, err)
		}
		if err := insertRuntimeRecords(ctx, tx, workflowTenant, instance.id, instance.correlation); err != nil {
			return err
		}
	}
	return nil
}

func insertRuntimeRecords(ctx context.Context, tx pgx.Tx, tenantID, instanceID, correlation string) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO events
			(tenant_id, event_id, type, source, aggregate_id, workflow_instance_id, timestamp, payload, correlation_id)
		VALUES ($1::uuid, $2, 'fixture.started', 'SYSTEM', $3, $4::uuid, NOW() - INTERVAL '2 minutes', $5::jsonb, $6)
	`, tenantID, "event-"+correlation, instanceID, instanceID, `{"kind":"synthetic","sequence":"started"}`, correlation); err != nil {
		return fmt.Errorf("insert fixture event %s: %w", correlation, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO context_records (tenant_id, scope_type, scope_id, key, value, version)
		VALUES
			($1::uuid, 'WORKFLOW_INSTANCE', $2, 'runtime.safe_context', $3::jsonb, 0),
			($1::uuid, 'WORKFLOW_INSTANCE', $2, 'runtime.debug_trace', $4::jsonb, 0)
	`, tenantID, instanceID,
		`{"bookingRef":"SYNTHETIC-001","customerTier":"standard"}`,
		`{"source":"local-fixture","status":"stubbed","durationMs":12,"correlationId":"`+correlation+`","reasonCode":"FIXTURE_TRACE","providerAlias":"local-stub"}`); err != nil {
		return fmt.Errorf("insert fixture context %s: %w", correlation, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO runtime_trace_entries
			(tenant_id, workflow_instance_id, turn_id, stage, source, status,
				occurred_at, correlation_id, duration_ms, reason_code, provider_alias,
				provider_reference, summary, attributes)
		VALUES ($1::uuid, $2::uuid, $3, 'MCP_ACTIVITY', 'EXTERNAL_PROVIDER', 'SUCCEEDED',
				NOW() - INTERVAL '1 minute', $4, 12, 'FIXTURE_TRACE', 'local-stub',
				$5, 'Synthetic provider metadata only', $6::jsonb)
	`, tenantID, instanceID, "turn-"+correlation, correlation,
		"operation-"+correlation, `{"mock":true,"transport":"local"}`); err != nil {
		return fmt.Errorf("insert fixture trace %s: %w", correlation, err)
	}
	return nil
}

func insertAuditExpectations(ctx context.Context, tx pgx.Tx) error {
	rows := []struct {
		resource, action, after string
	}{
		{SuspendedInstanceID, "workflow.suspended", `{"status":"SUSPENDED","source":"fixture"}`},
		{SuspendedInstanceID, "workflow.resumed", `{"status":"RUNNING","source":"fixture"}`},
		{FailedInstanceID, "workflow.retried", `{"status":"RUNNING","source":"fixture"}`},
	}
	for _, row := range rows {
		if _, err := tx.Exec(ctx, `
			INSERT INTO audit_logs (tenant_id, actor, action, resource_type, resource_id, after)
			VALUES ($1::uuid, $2::uuid, $3, 'workflow_instance', $4, $5::jsonb)
		`, TenantA, OperatorUser, row.action, row.resource, row.after); err != nil {
			return fmt.Errorf("insert fixture audit %s: %w", row.action, err)
		}
	}
	return nil
}

func Verify(ctx context.Context, pool *pgxpool.Pool, mode string) (*Verification, error) {
	checks := make([]string, 0, 8)
	if err := expectCount(ctx, pool, `SELECT COUNT(*) FROM users WHERE id IN ($1::uuid, $2::uuid, $3::uuid, $4::uuid)`, 4, EditorUser, OperatorUser, ViewerUser, SentinelUser); err != nil {
		return nil, err
	}
	checks = append(checks, "synthetic identities")
	if err := expectCount(ctx, pool, `SELECT COUNT(*) FROM role_assignments WHERE tenant_id = $1::uuid`, 3, TenantA); err != nil {
		return nil, err
	}
	checks = append(checks, "tenant-A roles")
	if err := expectCount(ctx, pool, `SELECT COUNT(*) FROM workflow_instances WHERE tenant_id = $1::uuid`, 3, TenantA); err != nil {
		return nil, err
	}
	checks = append(checks, "tenant-A runtime scope")
	if err := expectCount(ctx, pool, `SELECT COUNT(*) FROM workflow_instances WHERE tenant_id = $1::uuid`, 1, TenantB); err != nil {
		return nil, err
	}
	checks = append(checks, "sentinel tenant scope")

	var unsafe int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM context_records
		WHERE lower(value::text) ~ '(password|secret|access_token|refresh_token|authorization|raw_prompt|raw_response|rag_document|credential)'
	`).Scan(&unsafe); err != nil {
		return nil, fmt.Errorf("check fixture safety: %w", err)
	}
	if unsafe != 0 {
		return nil, errors.New("fixture contains disallowed sensitive data")
	}
	checks = append(checks, "safe JSON fields")

	switch mode {
	case "verify", "", "reset", "seed":
	case "verify-builder-draft":
		if err := expectCount(ctx, pool, `SELECT COUNT(*) FROM workflows WHERE id = $1::uuid AND draft_definition::text ILIKE '%INTAKE_EDITED%' AND version > 0`, 1, BuilderWorkflowID); err != nil {
			return nil, err
		}
		checks = append(checks, "builder draft persisted")
	case "verify-builder-published":
		if err := expectCount(ctx, pool, `SELECT COUNT(*) FROM workflow_versions WHERE workflow_id = $1::uuid AND version_no >= 3`, 1, BuilderWorkflowID); err != nil {
			return nil, err
		}
		checks = append(checks, "builder publish persisted")
	case "verify-runtime":
		if err := expectAtLeast(ctx, pool, `SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1::uuid AND actor = $2`, 3, TenantA, OperatorUser); err != nil {
			return nil, err
		}
		checks = append(checks, "runtime audit expectations")
	case "verify-runtime-suspend":
		if err := verifyRuntimeCommand(ctx, pool, RunningInstanceID, "SUSPENDED", "workflow.suspended"); err != nil {
			return nil, err
		}
		checks = append(checks, "suspend state and audit")
	case "verify-runtime-resume":
		if err := verifyRuntimeCommand(ctx, pool, SuspendedInstanceID, "RUNNING", "workflow.resumed"); err != nil {
			return nil, err
		}
		checks = append(checks, "resume state and audit")
	case "verify-runtime-retry":
		if err := verifyRuntimeCommand(ctx, pool, FailedInstanceID, "RUNNING", "workflow.retried"); err != nil {
			return nil, err
		}
		checks = append(checks, "retry state and audit")
	default:
		return nil, fmt.Errorf("unknown fixture verification mode %q", mode)
	}
	return &Verification{Mode: mode, Checks: checks}, nil
}

func BumpBuilderVersion(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `UPDATE workflows SET version = version + 1 WHERE id = $1::uuid AND tenant_id = $2::uuid`, BuilderWorkflowID, TenantA)
	if err != nil {
		return fmt.Errorf("bump builder fixture version: %w", err)
	}
	return nil
}

func expectCount(ctx context.Context, pool *pgxpool.Pool, query string, want int, args ...string) error {
	values := make([]any, len(args))
	for i, arg := range args {
		values[i] = arg
	}
	var got int
	if err := pool.QueryRow(ctx, query, values...).Scan(&got); err != nil {
		return fmt.Errorf("fixture verification query: %w", err)
	}
	if got != want {
		return fmt.Errorf("fixture verification expected %d rows, got %d", want, got)
	}
	return nil
}

func expectAtLeast(ctx context.Context, pool *pgxpool.Pool, query string, want int, args ...string) error {
	values := make([]any, len(args))
	for i, arg := range args {
		values[i] = arg
	}
	var got int
	if err := pool.QueryRow(ctx, query, values...).Scan(&got); err != nil {
		return fmt.Errorf("fixture verification query: %w", err)
	}
	if got < want {
		return fmt.Errorf("fixture verification expected at least %d rows, got %d", want, got)
	}
	return nil
}

func verifyRuntimeCommand(ctx context.Context, pool *pgxpool.Pool, instanceID, status, action string) error {
	if err := expectCount(ctx, pool, `
		SELECT COUNT(*) FROM workflow_instances
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND status = $3
	`, 1, instanceID, TenantA, status); err != nil {
		return err
	}
	if err := expectCount(ctx, pool, `
		SELECT COUNT(*) FROM audit_logs
		WHERE tenant_id = $1::uuid AND actor = $2
		  AND resource_id = $3 AND action = $4
		  AND after->>'outcome' = 'accepted'
	`, 1, TenantA, OperatorUser, instanceID, action); err != nil {
		return err
	}
	return nil
}

type fixtureNode struct {
	ID              string         `json:"id"`
	Kind            string         `json:"kind"`
	Name            string         `json:"name"`
	Description     string         `json:"description,omitempty"`
	RequiredContext []string       `json:"requiredContext"`
	Capabilities    []string       `json:"capabilities"`
	Policy          map[string]any `json:"policy"`
	Terminal        bool           `json:"isTerminal,omitempty"`
	Position        map[string]int `json:"position"`
}

type fixtureTransition struct {
	ID       string `json:"id"`
	Source   string `json:"sourceStateId"`
	Target   string `json:"targetStateId"`
	Event    string `json:"event"`
	Priority int    `json:"priority"`
}

func builderNodesV1() []fixtureNode {
	return []fixtureNode{
		{ID: startStateV1, Kind: "START", Name: "START", Terminal: false, Position: map[string]int{"x": 0, "y": 0}},
		{ID: intakeStateV1, Kind: "STATE", Name: "INTAKE", Description: "Synthetic intake state", Position: map[string]int{"x": 220, "y": 0}},
		{ID: endStateV1, Kind: "END", Name: "END", Terminal: true, Position: map[string]int{"x": 440, "y": 0}},
	}
}

func builderNodesV2() []fixtureNode {
	return []fixtureNode{
		{ID: startStateV2, Kind: "START", Name: "START", Position: map[string]int{"x": 0, "y": 0}},
		{ID: intakeStateV2, Kind: "STATE", Name: "INTAKE", Description: "Synthetic intake state", Position: map[string]int{"x": 220, "y": 0}},
		{ID: reviewStateV2, Kind: "STATE", Name: "REVIEW", Description: "Synthetic review state", Position: map[string]int{"x": 440, "y": 0}},
		{ID: endStateV2, Kind: "END", Name: "END", Terminal: true, Position: map[string]int{"x": 660, "y": 0}},
	}
}

func builderTransitionsV1() []fixtureTransition {
	return []fixtureTransition{
		{ID: startIntakeTransitionV1, Source: startStateV1, Target: intakeStateV1, Event: "workflow.started", Priority: 1},
		{ID: intakeEndTransitionV1, Source: intakeStateV1, Target: endStateV1, Event: "fixture.completed", Priority: 1},
	}
}

func builderTransitionsV2() []fixtureTransition {
	return []fixtureTransition{
		{ID: startIntakeTransitionV2, Source: startStateV2, Target: intakeStateV2, Event: "workflow.started", Priority: 1},
		{ID: intakeReviewTransitionV2, Source: intakeStateV2, Target: reviewStateV2, Event: "fixture.review", Priority: 1},
		{ID: reviewEndTransitionV2, Source: reviewStateV2, Target: endStateV2, Event: "fixture.completed", Priority: 1},
	}
}

func runtimeNodesV1() []fixtureNode {
	return []fixtureNode{
		{ID: runtimeStartState, Kind: "START", Name: "START", Position: map[string]int{"x": 0, "y": 0}},
		{ID: runtimeIntakeState, Kind: "STATE", Name: "INTAKE", Description: "Synthetic runtime state", Position: map[string]int{"x": 220, "y": 0}},
		{ID: runtimeEndState, Kind: "END", Name: "END", Terminal: true, Position: map[string]int{"x": 440, "y": 0}},
	}
}

func runtimeTransitionsV1() []fixtureTransition {
	return []fixtureTransition{
		{ID: "47000000-0000-0000-0000-000000000001", Source: runtimeStartState, Target: runtimeIntakeState, Event: "workflow.started", Priority: 1},
		{ID: "47000000-0000-0000-0000-000000000002", Source: runtimeIntakeState, Target: runtimeEndState, Event: "fixture.completed", Priority: 1},
	}
}

func sentinelNodesV1() []fixtureNode {
	return []fixtureNode{
		{ID: sentinelStartState, Kind: "START", Name: "START", Position: map[string]int{"x": 0, "y": 0}},
		{ID: sentinelIntakeState, Kind: "STATE", Name: "INTAKE", Description: "Synthetic sentinel state", Position: map[string]int{"x": 220, "y": 0}},
		{ID: sentinelEndState, Kind: "END", Name: "END", Terminal: true, Position: map[string]int{"x": 440, "y": 0}},
	}
}

func sentinelTransitionsV1() []fixtureTransition {
	return []fixtureTransition{
		{ID: "48000000-0000-0000-0000-000000000001", Source: sentinelStartState, Target: sentinelIntakeState, Event: "workflow.started", Priority: 1},
		{ID: "48000000-0000-0000-0000-000000000002", Source: sentinelIntakeState, Target: sentinelEndState, Event: "fixture.completed", Priority: 1},
	}
}

func definitionJSON(slug, name string, nodes []fixtureNode, transitions []fixtureTransition) []byte {
	type workflowPolicy struct {
		Interruptible string `json:"interruptible"`
		Priority      int    `json:"priority"`
	}
	type workflowTrigger struct {
		Event  string `json:"event"`
		Source string `json:"source"`
	}
	normalizedNodes := make([]fixtureNode, len(nodes))
	copy(normalizedNodes, nodes)
	for i := range normalizedNodes {
		if normalizedNodes[i].RequiredContext == nil {
			normalizedNodes[i].RequiredContext = []string{}
		}
		if normalizedNodes[i].Capabilities == nil {
			normalizedNodes[i].Capabilities = []string{}
		}
		if normalizedNodes[i].Policy == nil {
			normalizedNodes[i].Policy = map[string]any{}
		}
	}
	definition := struct {
		Slug          string              `json:"slug"`
		Name          string              `json:"name"`
		Description   string              `json:"description"`
		SchemaVersion int                 `json:"schemaVersion"`
		Status        string              `json:"status"`
		ProjectID     string              `json:"projectId"`
		EntryNodeID   string              `json:"entryNodeId"`
		Nodes         []fixtureNode       `json:"nodes"`
		Transitions   []fixtureTransition `json:"transitions"`
		Policy        workflowPolicy      `json:"policy"`
		Triggers      []workflowTrigger   `json:"triggers"`
	}{
		Slug:          slug,
		Name:          name,
		Description:   "Synthetic browser golden fixture",
		SchemaVersion: 1,
		Status:        "VALID",
		ProjectID:     ProjectID,
		EntryNodeID:   nodes[0].ID,
		Nodes:         normalizedNodes,
		Transitions:   transitions,
		Policy:        workflowPolicy{Interruptible: "USER_REQUESTED", Priority: 1},
		Triggers:      []workflowTrigger{{Event: "workflow.started", Source: "event"}},
	}
	raw, _ := json.Marshal(definition)
	return raw
}
