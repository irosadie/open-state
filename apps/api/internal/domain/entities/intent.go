package entities

import "time"

// Intent is a canonical, tenant/project-scoped routing choice that maps to a
// published workflow. Examples are natural-language utterances supplied to an
// external LLM for classification.
type Intent struct {
	ID           string
	TenantID     string
	ProjectID    string
	WorkflowID   string
	Key          string
	Name         string
	Description  string
	Examples     []string
	WorkflowSlug string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
