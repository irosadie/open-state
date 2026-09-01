package dtos

// IntentDTO is the response-safe projection of one canonical intent mapping.
// It intentionally includes the mapped workflow identity so Admin Console
// clients can navigate to Builder without resolving a slug themselves.
type IntentDTO struct {
	ID           string   `json:"id"`
	TenantID     string   `json:"tenantId"`
	ProjectID    string   `json:"projectId"`
	WorkflowID   string   `json:"workflowId"`
	Key          string   `json:"key"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Examples     []string `json:"examples"`
	WorkflowSlug string   `json:"workflowSlug"`
}

// IntentListDTO wraps a tenant/project-scoped intent catalog.
type IntentListDTO struct {
	Data []IntentDTO `json:"data"`
}
