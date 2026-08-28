package dtos

import "encoding/json"

// Workflow Builder API request/response DTOs (PRD 146). All operations are
// tenant+project scoped; tenant comes from the X-Tenant-ID header, never the
// body (PRD §74, §96).

// CreateWorkflowRequest is the payload to create a workflow definition draft.
type CreateWorkflowRequest struct {
	ProjectID   string `json:"projectId"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateWorkflowRequest is the payload to update a workflow draft's mutable
// fields using optimistic concurrency (PRD §31).
type UpdateWorkflowRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     int    `json:"version"`
}

// PublishWorkflowRequest is the payload to publish a workflow definition to an
// immutable, current version (PRD §3.3, §9, §55, §65, §69). Definition holds the
// full WorkflowDefinition envelope as JSONB (PRD §68).
type PublishWorkflowRequest struct {
	Version    int             `json:"version"`
	Definition json.RawMessage `json:"definition"`
}

// WorkflowDTO is the serializable workflow definition root returned to callers.
type WorkflowDTO struct {
	ID             string  `json:"id"`
	TenantID       string  `json:"tenantId"`
	ProjectID      string  `json:"projectId"`
	Slug           string  `json:"slug"`
	Name           string  `json:"name"`
	Description    *string `json:"description"`
	Status         string  `json:"status"`
	CurrentVersion int     `json:"currentVersion"`
	Version        int     `json:"version"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

// WorkflowListDTO wraps a tenant/project-scoped workflow list.
type WorkflowListDTO struct {
	Data []WorkflowDTO `json:"data"`
}

// WorkflowVersionDTO is the serializable immutable workflow version.
type WorkflowVersionDTO struct {
	ID         string `json:"id"`
	WorkflowID string `json:"workflowId"`
	VersionNo  int    `json:"versionNo"`
	Status     string `json:"status"`
	IsCurrent  bool   `json:"isCurrent"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}
