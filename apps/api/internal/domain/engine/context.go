package engine

// ContextScope identifies where a context entry lives (PRD §23).
type ContextScope string

const (
	ScopeTenant       ContextScope = "TENANT"
	ScopeConversation ContextScope = "CONVERSATION"
	ScopeWorkflow     ContextScope = "WORKFLOW"
	ScopeState        ContextScope = "STATE"
	ScopeTurn         ContextScope = "TURN"
	ScopeMemory       ContextScope = "MEMORY"
)

// scopeOrder defines precedence: later scope overrides earlier on key conflict.
var scopeOrder = []ContextScope{
	ScopeTenant,
	ScopeConversation,
	ScopeWorkflow,
	ScopeState,
	ScopeTurn,
}

// ContextEntry is a single key/value in a scope.
type ContextEntry struct {
	Scope     ContextScope `json:"scope"`
	Key       string       `json:"key"`
	Value     any          `json:"value"`
	Sensitive bool         `json:"sensitive,omitempty"`
}

// ResolvedContext is the output of the context resolver.
type ResolvedContext struct {
	// Available is the merged context (later scopes override earlier).
	Available map[string]any `json:"available"`
	// Missing lists required keys that are absent.
	Missing []string `json:"missing"`
	// Memory holds persistent customer data (PRD §24).
	Memory map[string]any `json:"memory"`
	// WorkflowData holds transient workflow data (PRD §24).
	WorkflowData map[string]any `json:"workflowData"`
}

// ContextResolver merges hierarchical context and detects missing required
// context for a state. It is domain-pure (values are passed in, not fetched).
type ContextResolver struct {
	scopes map[ContextScope]map[string]ContextEntry
}

// NewContextResolver builds a resolver.
func NewContextResolver() *ContextResolver {
	return &ContextResolver{scopes: map[ContextScope]map[string]ContextEntry{}}
}

// Set stores a context entry in a scope.
func (r *ContextResolver) Set(scope ContextScope, key string, value any, sensitive bool) {
	if r.scopes[scope] == nil {
		r.scopes[scope] = map[string]ContextEntry{}
	}
	r.scopes[scope][key] = ContextEntry{Scope: scope, Key: key, Value: value, Sensitive: sensitive}
}

// Clear removes all scopes.
func (r *ContextResolver) Clear() {
	r.scopes = map[ContextScope]map[string]ContextEntry{}
}

// Resolve merges context by precedence and computes missing required context.
func (r *ContextResolver) Resolve(requiredContext []string) *ResolvedContext {
	available := map[string]any{}
	memory := map[string]any{}
	workflow := map[string]any{}

	// merge by precedence (later scope wins); memory is a base source too
	allScopes := append([]ContextScope{ScopeMemory}, scopeOrder...)
	for _, scope := range allScopes {
		for k, e := range r.scopes[scope] {
			available[k] = e.Value
		}
	}
	// memory & workflow are separate from the merge above
	for k, e := range r.scopes[ScopeMemory] {
		memory[k] = e.Value
	}
	for _, scope := range []ContextScope{ScopeWorkflow, ScopeState, ScopeTurn} {
		for k, e := range r.scopes[scope] {
			workflow[k] = e.Value
		}
	}

	missing := []string{}
	for _, req := range requiredContext {
		if _, ok := available[req]; !ok {
			missing = append(missing, req)
		}
	}

	return &ResolvedContext{
		Available:    available,
		Missing:      missing,
		Memory:       memory,
		WorkflowData: workflow,
	}
}

// Entry returns a single context entry from any scope.
func (r *ContextResolver) Entry(scope ContextScope, key string) (ContextEntry, bool) {
	e, ok := r.scopes[scope][key]
	return e, ok
}
