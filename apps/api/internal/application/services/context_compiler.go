package services

import (
	"context"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	domainrag "github.com/irosadie/open-state/api/internal/domain/rag"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

// ContextCompiler compiles the minimal per-turn context for an LLM/RAG client
// (PRD 22). It composes runtime context, persistent memory, workflow data, and
// RAG retrievals, then applies PII redaction last (PRD 90). It depends only on
// domain ports + repository interfaces (ADR-001, PRD 169).
type ContextCompiler struct {
	repo       repositories.IContextRepository
	ragProvider domainrag.RAGProvider
	redactor   domainrag.Redactor
}

// NewContextCompiler builds a ContextCompiler.
func NewContextCompiler(
	repo repositories.IContextRepository,
	ragProvider domainrag.RAGProvider,
	redactor domainrag.Redactor,
) *ContextCompiler {
	return &ContextCompiler{repo: repo, ragProvider: ragProvider, redactor: redactor}
}

// CompileArgs selects the scopes to compile context for.
type CompileArgs struct {
	TenantID   string
	ConversationID string
	// WorkflowInstanceID scopes runtime context to a workflow instance (optional).
	WorkflowInstanceID string
	// OwnerType/OwnerID select persistent memory (e.g. CUSTOMER / customer-id).
	OwnerType string
	OwnerID   string
	// Query optionally retrieves RAG knowledge (PRD 171).
	Query string
}

// Compile builds the compiled context for a turn.
func (c *ContextCompiler) Compile(ctx context.Context, args CompileArgs) (*dtos.CompiledContext, error) {
	out := &dtos.CompiledContext{
		Available: map[string]any{},
		Missing:   []string{},
		Memory:    map[string]any{},
		Workflow:  map[string]any{},
	}

	// Runtime context for the conversation (and instance if provided).
	if err := c.loadScope(ctx, args.TenantID, entities.ContextScopeConversation, args.ConversationID, out.Available); err != nil {
		return nil, err
	}
	if args.WorkflowInstanceID != "" {
		if err := c.loadScope(ctx, args.TenantID, entities.ContextScopeWorkflowInstance, args.WorkflowInstanceID, out.Workflow); err != nil {
			return nil, err
		}
	}

	// Persistent memory (kept separate from workflow runtime data, PRD §24).
	if args.OwnerType != "" && args.OwnerID != "" {
		if err := c.loadMemory(ctx, args.TenantID, args.OwnerType, args.OwnerID, out.Memory); err != nil {
			return nil, err
		}
	}

	// RAG knowledge (PRD 171).
	if c.ragProvider != nil && args.Query != "" {
		retrievals, err := c.loadRetrievals(ctx, args.Query)
		if err != nil {
			return nil, err
		}
		out.Retrieval = retrievals
	}

	// Redact text values last (PRD 90).
	if err := c.redact(out); err != nil {
		return nil, err
	}

	return out, nil
}

func (c *ContextCompiler) loadScope(ctx context.Context, tenantID string, scopeType entities.ContextScopeType, scopeID string, target map[string]any) error {
	records, err := c.repo.ListContextByScope(ctx, tenantID, scopeType, scopeID)
	if err != nil {
		return err
	}
	for i := range records {
		target[records[i].Key] = rawToAny(records[i].Value)
	}
	return nil
}

func (c *ContextCompiler) loadMemory(ctx context.Context, tenantID, ownerType, ownerID string, target map[string]any) error {
	refs, err := c.repo.ListMemoryByOwner(ctx, tenantID, ownerType, ownerID)
	if err != nil {
		return err
	}
	for i := range refs {
		target[refs[i].Name] = rawToAny(refs[i].Value)
	}
	return nil
}

func (c *ContextCompiler) loadRetrievals(ctx context.Context, query string) ([]map[string]any, error) {
	retrieval, err := c.ragProvider.Retrieve(ctx, query)
	if err != nil {
		return nil, err
	}
	if retrieval == nil || retrieval.Text == "" {
		return nil, nil
	}
	item := map[string]any{"text": retrieval.Text}
	for k, v := range retrieval.Metadata {
		item[k] = v
	}
	return []map[string]any{item}, nil
}

func (c *ContextCompiler) redact(out *dtos.CompiledContext) error {
	if c.redactor == nil {
		return nil
	}

	redacted := false
	redactSection := func(section map[string]any) error {
		for key, value := range section {
			s, ok := value.(string)
			if !ok || s == "" {
				continue
			}
			red, err := c.redactor.Redact(context.Background(), s)
			if err != nil {
				return err
			}
			if red != s {
				section[key] = red
				redacted = true
			}
		}
		return nil
	}

	if err := redactSection(out.Available); err != nil {
		return err
	}
	if err := redactSection(out.Memory); err != nil {
		return err
	}
	if err := redactSection(out.Workflow); err != nil {
		return err
	}

	out.Redacted = redacted
	return nil
}
