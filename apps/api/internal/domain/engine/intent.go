package engine

// IntentDefinition maps a conversation intent to a workflow (PRD §40.1).
type IntentDefinition struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	WorkflowSlug string   `json:"workflowSlug"`
	EntryEvent   string   `json:"entryEvent,omitempty"`
	Examples     []string `json:"examples,omitempty"`
	Priority     int      `json:"priority"`
}

// IntentRegistry holds the defined intents for a tenant.
type IntentRegistry struct {
	SchemaVersion int               `json:"schemaVersion"`
	Intents       []IntentDefinition `json:"intents"`
}

// IntentResolver maps intent id → workflow + entry event.
// It depends on a provider so it can look up the workflow definition.
type IntentResolver struct {
	registry   IntentRegistry
	workflowFn func(slug string) (*WorkflowDefinition, bool)
}

// NewIntentResolver builds a resolver backed by the given registry and workflow lookup.
func NewIntentResolver(reg IntentRegistry, lookup func(slug string) (*WorkflowDefinition, bool)) *IntentResolver {
	return &IntentResolver{registry: reg, workflowFn: lookup}
}

// ResolveIntent returns the workflow, entry event, and initial state for an intent id.
func (r *IntentResolver) ResolveIntent(intentID string) (*WorkflowDefinition, string, string, bool) {
	for _, intent := range r.registry.Intents {
		if intent.ID != intentID {
			continue
		}
		if r.workflowFn == nil {
			return nil, "", "", false
		}
		def, ok := r.workflowFn(intent.WorkflowSlug)
		if !ok {
			return nil, "", "", false
		}
		entry := intent.EntryEvent
		if entry == "" {
			entry = "workflow.started"
		}
		return def, entry, def.EntryNodeID, true
	}
	return nil, "", "", false
}

// ListIntents returns all registered intents.
func (r *IntentResolver) ListIntents() []IntentDefinition {
	return r.registry.Intents
}
