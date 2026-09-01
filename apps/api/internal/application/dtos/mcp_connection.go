package dtos

// MCP connection DTOs intentionally omit credential values and protected
// credential references. A credentialReference is an opaque reference resolved
// by the server's secret infrastructure and is accepted only on writes.
type CreateMCPConnectionRequest struct {
	Name                string   `json:"name"`
	Alias               string   `json:"alias"`
	Transport           string   `json:"transport"`
	Endpoint            string   `json:"endpoint"`
	StdioProfile        string   `json:"stdioProfile"`
	StdioArgs           []string `json:"stdioArgs"`
	AuthType            string   `json:"authType"`
	CredentialReference string   `json:"credentialReference"`
}

type UpdateMCPConnectionRequest = CreateMCPConnectionRequest

type MCPConnectionDTO struct {
	ID                string   `json:"id"`
	TenantID          string   `json:"tenantId"`
	ProjectID         string   `json:"projectId"`
	Name              string   `json:"name"`
	Alias             string   `json:"alias"`
	Transport         string   `json:"transport"`
	Endpoint          *string  `json:"endpoint"`
	StdioProfile      *string  `json:"stdioProfile"`
	StdioArgs         []string `json:"stdioArgs"`
	AuthType          string   `json:"authType"`
	CredentialStatus  string   `json:"credentialStatus"`
	Status            string   `json:"status"`
	LastTestStatus    string   `json:"lastTestStatus"`
	LastTestErrorCode *string  `json:"lastTestErrorCode"`
	LastTestedAt      *string  `json:"lastTestedAt"`
	CreatedAt         string   `json:"createdAt"`
	UpdatedAt         string   `json:"updatedAt"`
}

type MCPConnectionListDTO struct {
	Data []MCPConnectionDTO `json:"data"`
}
