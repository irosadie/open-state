# engine-context Specification

## Purpose
Resolve hierarchical context for the runtime engine: available vs missing,
with persistent-memory and workflow-data separation (PRD §23, §24, §36).

## Requirements

### Requirement: Hierarchical resolution

The resolver SHALL merge context across scopes with defined precedence.

#### Scenario: Scope precedence

- GIVEN context entries in tenant, conversation, workflow, state, and turn scopes
- THEN the resolver merges them
- AND a later scope (turn) overrides an earlier scope (tenant) on key conflict

### Requirement: Missing context detection

The resolver SHALL compute what a state still requires.

#### Scenario: Required context

- GIVEN a state declares `requiredContext`
- THEN the resolver returns the entries present in available context
- AND returns a `missing` list for the absent required keys

### Requirement: Memory vs workflow split

The resolver SHALL keep persistent memory separate from transient workflow data.

#### Scenario: Separate scopes

- GIVEN customer memory (e.g. `customer.address`) and workflow data (e.g.
  `booking.date`)
- THEN the resolver exposes them as separate scopes
- AND removing workflow data does not remove memory

### Requirement: Sensitivity flag

The resolver SHALL mark sensitive entries for later PII redaction.

#### Scenario: Sensitive key

- GIVEN an entry marked sensitive (e.g. payment token)
- THEN it carries a `sensitive` flag
- AND downstream (MCP/LLM) may redact it (Epic #4)
