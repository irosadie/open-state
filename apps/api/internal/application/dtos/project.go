package dtos

// ProjectDTO is the response-safe projection used by tenant-scoped project
// discovery clients such as the State MCP API-key console.
type ProjectDTO struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenantId"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ProjectListDTO wraps the current tenant's projects.
type ProjectListDTO struct {
	Data []ProjectDTO `json:"data"`
}
