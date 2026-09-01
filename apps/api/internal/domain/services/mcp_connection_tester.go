package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// MCPConnectionTester is the transport-neutral handshake port. Handshake stays
// intentionally separate from explicit tool discovery.
type MCPConnectionTester interface {
	Handshake(ctx context.Context, connection *entities.MCPConnection) (MCPHandshakeResult, error)
}

type MCPHandshakeResult struct {
	Ready     bool
	ErrorCode string
}

// MCPToolDiscoverer is called only by an explicit catalog refresh. Implementations
// must initialize the session and request tools/list; they must never call tools/call.
type MCPToolDiscoverer interface {
	DiscoverTools(ctx context.Context, connection *entities.MCPConnection) (MCPToolDiscoveryResult, error)
}

// MCPResolvedToolCaller is the provider execution port used by the enforced
// gateway. The caller receives a connection and tool resolved by OpenState's
// project binding chain; it must not accept an endpoint, credential, or
// caller-selected tool name separately.
type MCPResolvedToolCaller interface {
	InvokeTool(ctx context.Context, connection *entities.MCPConnection, tool *entities.MCPDiscoveredTool, payload map[string]any, timeout time.Duration) (MCPToolCallResult, error)
}

type MCPToolCallResult struct {
	Data     map[string]any
	Duration time.Duration
}

type MCPToolDiscoveryResult struct {
	Tools     []MCPToolDefinition
	ErrorCode string
}

type MCPToolDefinition struct {
	Name        string
	Title       string
	Description string
	InputSchema json.RawMessage
	Annotations json.RawMessage
}
