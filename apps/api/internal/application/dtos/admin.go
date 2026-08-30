package dtos

import "encoding/json"

// TenantDTO is the safe current-tenant profile returned to the console.
type TenantDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// UpdateTenantRequest contains the editable current-tenant profile fields.
// Pointers preserve PATCH semantics: omitted fields retain their current value.
type UpdateTenantRequest struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
}

type TenantMembershipDTO struct {
	RoleAssignmentID string  `json:"roleAssignmentId"`
	UserID           string  `json:"userId"`
	TenantID         string  `json:"tenantId"`
	Role             string  `json:"role"`
	Email            string  `json:"email"`
	Name             string  `json:"name"`
	Status           string  `json:"status"`
	Photo            *string `json:"photo,omitempty"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}

type TenantMembershipPageDTO struct {
	Data     []TenantMembershipDTO `json:"data"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
	Total    int64                 `json:"total"`
	HasNext  bool                  `json:"hasNext"`
}

type UpdateMembershipRoleRequest struct {
	Role string `json:"role"`
}

// InstanceDTO is the stable console representation of a workflow instance.
type InstanceDTO struct {
	ID                     string  `json:"id"`
	TenantID               string  `json:"tenantId"`
	WorkflowID             string  `json:"workflowId"`
	WorkflowVersionID      string  `json:"workflowVersionId"`
	CorrelationKey         *string `json:"correlationKey,omitempty"`
	Status                 string  `json:"status"`
	Version                int     `json:"version"`
	CurrentStateInstanceID *string `json:"currentStateInstanceId,omitempty"`
	StartedAt              *string `json:"startedAt,omitempty"`
	CompletedAt            *string `json:"completedAt,omitempty"`
	ExpiresAt              *string `json:"expiresAt,omitempty"`
	CreatedAt              string  `json:"createdAt"`
	UpdatedAt              string  `json:"updatedAt"`
}

type EventDTO struct {
	ID                 string          `json:"id"`
	TenantID           string          `json:"tenantId"`
	EventID            string          `json:"eventId"`
	Type               string          `json:"type"`
	Source             string          `json:"source"`
	AggregateID        *string         `json:"aggregateId,omitempty"`
	WorkflowInstanceID *string         `json:"workflowInstanceId,omitempty"`
	Sequence           int64           `json:"sequence"`
	Timestamp          string          `json:"timestamp"`
	Payload            json.RawMessage `json:"payload"`
	CorrelationID      *string         `json:"correlationId,omitempty"`
	CausationID        *string         `json:"causationId,omitempty"`
	CreatedAt          string          `json:"createdAt"`
}

type EventPageDTO struct {
	Data     []EventDTO `json:"data"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Total    int64      `json:"total"`
	HasNext  bool       `json:"hasNext"`
}
