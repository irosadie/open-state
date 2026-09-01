package capability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
)

// fixtureFile is the on-disk shape of testdata/capability_fixtures.json.
type fixtureFile struct {
	Capabilities map[string]fixtureEntry `json:"capabilities"`
}

type fixtureEntry struct {
	Event string         `json:"event"`
	Data  map[string]any `json:"data"`
}

// JSONFileProvider is a test/sandbox capability provider that reads preset
// responses from a JSON fixtures file instead of calling a real MCP/HTTP
// backend. Useful for flow testing without any live third-party dependency.
//
// If the requested capability name is not found in the file, it falls back to
// the MockProvider behavior (echo + from_mock flag) so the flow keeps running.
type JSONFileProvider struct {
	fixtures map[string]fixtureEntry
}

// NewJSONFileProvider loads the fixtures file at path and returns a provider.
// Returns an error if the file cannot be read or parsed.
func NewJSONFileProvider(path string) (*JSONFileProvider, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("jsonfile_provider: open %q: %w", path, err)
	}
	defer f.Close()

	var ff fixtureFile
	if err := json.NewDecoder(f).Decode(&ff); err != nil {
		return nil, fmt.Errorf("jsonfile_provider: decode %q: %w", path, err)
	}
	return &JSONFileProvider{fixtures: ff.Capabilities}, nil
}

// Invoke implements domaincap.CapabilityProvider.
// It looks up inv.Name in the fixtures map and returns the preset response.
// Unknown capabilities fall back to mock-echo behavior.
func (p *JSONFileProvider) Invoke(_ context.Context, inv domaincap.Invocation) (domaincap.InvocationResult, error) {
	entry, ok := p.fixtures[inv.Name]
	if !ok {
		// Fallback: echo like MockProvider so unknown capabilities don't block the flow.
		event := "capability.success"
		return domaincap.InvocationResult{
			Data: map[string]any{
				"capability": inv.Name,
				"mock":       true,
				"echo":       inv.Payload,
				"action_id":  inv.ActionID,
				"note":       "capability not found in fixtures — using mock echo",
			},
			FromMock:        true,
			Duration:        time.Millisecond * 5,
			CapabilityEvent: &event,
		}, nil
	}

	event := entry.Event
	// Merge fixture data with echoed payload so the LLM sees both
	// the preset response AND the inputs it sent.
	data := make(map[string]any, len(entry.Data)+2)
	for k, v := range entry.Data {
		data[k] = v
	}
	data["_input"] = inv.Payload
	data["_capability"] = inv.Name

	return domaincap.InvocationResult{
		Data:            data,
		FromMock:        true,
		Duration:        time.Millisecond * 10,
		CapabilityEvent: &event,
	}, nil
}

// JSONFileProviderResolver resolves every capability to the same JSONFileProvider.
type JSONFileProviderResolver struct {
	provider *JSONFileProvider
}

// NewJSONFileProviderResolver returns a resolver backed by the given fixtures file.
// Falls back to MockProviderResolver if the file cannot be loaded.
func NewJSONFileProviderResolver(path string) (*JSONFileProviderResolver, error) {
	p, err := NewJSONFileProvider(path)
	if err != nil {
		return nil, err
	}
	return &JSONFileProviderResolver{provider: p}, nil
}

// ResolveProvider implements domaincap.ProviderResolver.
func (r *JSONFileProviderResolver) ResolveProvider(_ *domaincap.ResolvedCapability) domaincap.CapabilityProvider {
	return r.provider
}
