package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

type workflowGraph struct {
	Nodes       []graphNode       `json:"nodes"`
	Transitions []graphTransition `json:"transitions"`
}

type graphNode struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	IsTerminal bool   `json:"isTerminal"`
}

type graphTransition struct {
	ID            string `json:"id"`
	SourceStateID string `json:"sourceStateId"`
	TargetStateID string `json:"targetStateId"`
}

type rawWorkflowGraph struct {
	Nodes       []json.RawMessage `json:"nodes"`
	Transitions []json.RawMessage `json:"transitions"`
}

func ensureWorkflowDefinitionJSON(raw []byte) *domain.DomainError {
	if len(bytes.TrimSpace(raw)) == 0 {
		return domain.NewValidation("definition is required")
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil || envelope == nil {
		return domain.NewValidation("definition must be a JSON object")
	}
	return nil
}

func validateWorkflowDefinition(raw []byte) *domain.DomainError {
	if err := ensureWorkflowDefinitionJSON(raw); err != nil {
		return err
	}
	var graph workflowGraph
	if err := json.Unmarshal(raw, &graph); err != nil {
		return domain.NewValidation("definition must be valid JSON")
	}

	issues := make([]dtos.WorkflowValidationIssue, 0)
	if len(graph.Nodes) == 0 {
		issues = append(issues, dtos.WorkflowValidationIssue{Code: "GRAPH_EMPTY", Message: "workflow must contain at least one node"})
	}
	nodes := make(map[string]graphNode, len(graph.Nodes))
	startIDs := make([]string, 0, 1)
	terminalIDs := make(map[string]struct{})
	for _, node := range graph.Nodes {
		if node.ID == "" {
			issues = append(issues, dtos.WorkflowValidationIssue{Code: "NODE_ID_REQUIRED", Message: "every node must have an id"})
			continue
		}
		if _, exists := nodes[node.ID]; exists {
			issues = append(issues, dtos.WorkflowValidationIssue{Code: "DUPLICATE_NODE", Message: fmt.Sprintf("node id %q is duplicated", node.ID), NodeID: node.ID})
			continue
		}
		nodes[node.ID] = node
		if node.Kind == "START" {
			startIDs = append(startIDs, node.ID)
		}
		if node.Kind == "END" || node.IsTerminal {
			terminalIDs[node.ID] = struct{}{}
		}
	}
	if len(startIDs) != 1 {
		issues = append(issues, dtos.WorkflowValidationIssue{Code: "START_COUNT", Message: "workflow must contain exactly one START node"})
	}

	adjacency := make(map[string][]string, len(nodes))
	transitionIDs := make(map[string]struct{}, len(graph.Transitions))
	for _, transition := range graph.Transitions {
		if transition.ID == "" {
			issues = append(issues, dtos.WorkflowValidationIssue{Code: "TRANSITION_ID_REQUIRED", Message: "every transition must have an id"})
			continue
		}
		if _, exists := transitionIDs[transition.ID]; exists {
			issues = append(issues, dtos.WorkflowValidationIssue{Code: "DUPLICATE_TRANSITION", Message: fmt.Sprintf("transition id %q is duplicated", transition.ID), TransitionID: transition.ID})
			continue
		}
		transitionIDs[transition.ID] = struct{}{}
		if _, exists := nodes[transition.SourceStateID]; !exists {
			issues = append(issues, dtos.WorkflowValidationIssue{Code: "INVALID_SOURCE", Message: "transition source does not reference a node", TransitionID: transition.ID})
		}
		if _, exists := nodes[transition.TargetStateID]; !exists {
			issues = append(issues, dtos.WorkflowValidationIssue{Code: "INVALID_TARGET", Message: "transition target does not reference a node", TransitionID: transition.ID})
		}
		if _, sourceExists := nodes[transition.SourceStateID]; sourceExists {
			if _, targetExists := nodes[transition.TargetStateID]; targetExists {
				adjacency[transition.SourceStateID] = append(adjacency[transition.SourceStateID], transition.TargetStateID)
			}
		}
	}

	if len(startIDs) == 1 && len(terminalIDs) > 0 {
		reachable := map[string]struct{}{startIDs[0]: {}}
		queue := []string{startIDs[0]}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			for _, next := range adjacency[current] {
				if _, seen := reachable[next]; seen {
					continue
				}
				reachable[next] = struct{}{}
				queue = append(queue, next)
			}
		}
		terminalReachable := false
		for terminalID := range terminalIDs {
			if _, ok := reachable[terminalID]; ok {
				terminalReachable = true
				break
			}
		}
		if !terminalReachable {
			issues = append(issues, dtos.WorkflowValidationIssue{Code: "NO_TERMINAL_PATH", Message: "START must reach a terminal node"})
		}
	} else if len(terminalIDs) == 0 {
		issues = append(issues, dtos.WorkflowValidationIssue{Code: "NO_TERMINAL", Message: "workflow must contain an END or terminal node"})
	}

	if len(issues) == 0 {
		return nil
	}
	details, err := json.Marshal(issues)
	if err != nil {
		return domain.NewValidation("workflow definition is invalid")
	}
	return domain.NewValidationWithDetails("workflow definition is invalid", details)
}

func compareWorkflowDefinitions(workflowID string, baseVersion, targetVersion int, baseRaw, targetRaw []byte) (*dtos.WorkflowDiffDTO, error) {
	var base, target rawWorkflowGraph
	if err := json.Unmarshal(baseRaw, &base); err != nil {
		return nil, domain.NewValidation("base version definition is invalid")
	}
	if err := json.Unmarshal(targetRaw, &target); err != nil {
		return nil, domain.NewValidation("target version definition is invalid")
	}
	return &dtos.WorkflowDiffDTO{
		WorkflowID:    workflowID,
		BaseVersion:   baseVersion,
		TargetVersion: targetVersion,
		Nodes:         compareGraphItems(base.Nodes, target.Nodes),
		Transitions:   compareGraphItems(base.Transitions, target.Transitions),
	}, nil
}

type graphItemID struct {
	ID string `json:"id"`
}

func compareGraphItems(base, target []json.RawMessage) dtos.WorkflowDiffGroup {
	baseItems := indexGraphItems(base)
	targetItems := indexGraphItems(target)
	group := dtos.WorkflowDiffGroup{
		Added:   make([]dtos.WorkflowDiffItem, 0),
		Removed: make([]dtos.WorkflowDiffItem, 0),
		Changed: make([]dtos.WorkflowDiffItem, 0),
	}
	for id, raw := range targetItems {
		old, exists := baseItems[id]
		if !exists {
			group.Added = append(group.Added, dtos.WorkflowDiffItem{ID: id, Definition: raw})
			continue
		}
		fields := changedFields(old, raw)
		if len(fields) > 0 {
			group.Changed = append(group.Changed, dtos.WorkflowDiffItem{ID: id, ChangedFields: fields})
		}
	}
	for id, raw := range baseItems {
		if _, exists := targetItems[id]; !exists {
			group.Removed = append(group.Removed, dtos.WorkflowDiffItem{ID: id, Definition: raw})
		}
	}
	sort.Slice(group.Added, func(i, j int) bool { return group.Added[i].ID < group.Added[j].ID })
	sort.Slice(group.Removed, func(i, j int) bool { return group.Removed[i].ID < group.Removed[j].ID })
	sort.Slice(group.Changed, func(i, j int) bool { return group.Changed[i].ID < group.Changed[j].ID })
	return group
}

func indexGraphItems(items []json.RawMessage) map[string]json.RawMessage {
	indexed := make(map[string]json.RawMessage, len(items))
	for _, raw := range items {
		var item graphItemID
		if err := json.Unmarshal(raw, &item); err == nil && item.ID != "" {
			indexed[item.ID] = raw
		}
	}
	return indexed
}

func changedFields(base, target json.RawMessage) []string {
	var baseFields, targetFields map[string]json.RawMessage
	if json.Unmarshal(base, &baseFields) != nil || json.Unmarshal(target, &targetFields) != nil {
		return []string{"definition"}
	}
	keys := make(map[string]struct{}, len(baseFields)+len(targetFields))
	for key := range baseFields {
		keys[key] = struct{}{}
	}
	for key := range targetFields {
		keys[key] = struct{}{}
	}
	changed := make([]string, 0)
	for key := range keys {
		if !jsonValuesEqual(baseFields[key], targetFields[key]) {
			changed = append(changed, key)
		}
	}
	sort.Strings(changed)
	return changed
}

func jsonValuesEqual(base, target json.RawMessage) bool {
	var baseValue, targetValue any
	if json.Unmarshal(base, &baseValue) != nil || json.Unmarshal(target, &targetValue) != nil {
		return bytes.Equal(bytes.TrimSpace(base), bytes.TrimSpace(target))
	}
	return reflect.DeepEqual(baseValue, targetValue)
}
