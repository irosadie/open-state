package dtos

// AuditEntryDTO is a safe, external-facing representation of an audit entry
// (PRD 50). It never exposes internal fields or raw secret data.
type AuditEntryDTO struct {
	ID            string          `json:"id"`
	TenantID      string          `json:"tenantId"`
	Actor         string          `json:"actor"`
	Action        string          `json:"action"`
	ResourceType  string          `json:"resourceType"`
	ResourceID    string          `json:"resourceId"`
	Before        *map[string]any `json:"before,omitempty"`
	After         *map[string]any `json:"after,omitempty"`
	CorrelationID *string         `json:"correlationId,omitempty"`
	OccurredAt    string          `json:"occurredAt"`
}

// AuditPageDTO is the paginated envelope for audit listing (PRD 50).
type AuditPageDTO struct {
	Data     []AuditEntryDTO `json:"data"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
	Total    int64           `json:"total"`
	HasNext  bool            `json:"hasNext"`
}
