package engine

// IntentDefinition maps a conversation intent (within a project) to a workflow.
// Hierarchy: Project → Intent → Workflow → State (PRD §40.1).
type IntentDefinition struct {
	ID           string   `json:"id"`
	ProjectID    string   `json:"projectId"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	WorkflowSlug string   `json:"workflowSlug"`
	EntryEvent   string   `json:"entryEvent,omitempty"`
	Examples     []string `json:"examples,omitempty"`
	Priority     int      `json:"priority"`
}

// IntentRegistry holds the defined intents, scoped per project.
type IntentRegistry struct {
	SchemaVersion int               `json:"schemaVersion"`
	Intents       []IntentDefinition `json:"intents"`
}

// IntentResolver maps (project, intent id) → workflow + entry event + initial state.
// It depends on a provider so it can look up the workflow definition.
type IntentResolver struct {
	registry   IntentRegistry
	workflowFn func(projectID, slug string) (*WorkflowDefinition, bool)
}

// NewIntentResolver builds a resolver backed by the given registry and workflow lookup.
func NewIntentResolver(reg IntentRegistry, lookup func(projectID, slug string) (*WorkflowDefinition, bool)) *IntentResolver {
	return &IntentResolver{registry: reg, workflowFn: lookup}
}

// ResolveIntent returns the workflow, entry event, and initial state for an
// (project, intent) pair. The workflow is resolved within the same project.
func (r *IntentResolver) ResolveIntent(projectID, intentID string) (*WorkflowDefinition, string, string, bool) {
	for _, intent := range r.registry.Intents {
		if intent.ProjectID != projectID || intent.ID != intentID {
			continue
		}
		if r.workflowFn == nil {
			return nil, "", "", false
		}
		def, ok := r.workflowFn(projectID, intent.WorkflowSlug)
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

// ListIntents returns all registered intents for a project.
func (r *IntentResolver) ListIntents(projectID string) []IntentDefinition {
	var out []IntentDefinition
	for _, intent := range r.registry.Intents {
		if intent.ProjectID == projectID {
			out = append(out, intent)
		}
	}
	return out
}
