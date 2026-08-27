# mcp/server-runtime Specification

## Purpose
Define a standalone MCP server that exposes the Orchestrator runtime to external LLM/AI
systems: intent resolution, active workflow lookup, compiled context, and authorized
capability invocation — over the Model Context Protocol.

## Requirements

### Requirement: MCP server startup

The platform SHALL run a standalone MCP server reachable over Streamable HTTP.

#### Scenario: Server exposes tools

- **WHEN** the MCP server starts
- **THEN** it connects and declares tools for intent resolution, active workflow,
  context retrieval, and capability invocation

### Requirement: Intent resolution tool

The MCP server SHALL expose an intent-resolution tool that returns the classified
intent, its workflow, and current state to the LLM.

#### Scenario: Resolve intent

- **WHEN** the LLM calls the intent-resolution tool with an intent
- **THEN** the server returns the mapped workflow and the current state for that intent

### Requirement: Active workflow lookup

The MCP server SHALL expose a tool that returns the active workflow and current state
for a conversation.

#### Scenario: Get active workflow

- **WHEN** the LLM calls the active-workflow tool for a conversation
- **THEN** the server returns the active workflow, current state, and allowed events
- **AND** returns an indication when no active workflow exists

### Requirement: Context retrieval tool

The MCP server SHALL expose a tool that returns the compiled runtime context for a turn.

#### Scenario: Get context

- **WHEN** the LLM calls the context tool
- **THEN** the server returns tenant context, active workflow, current state, state
  purpose and instructions, available and missing context, allowed events and
  transitions, available capabilities, and relevant memory

### Requirement: Capability invocation tool

The MCP server SHALL let the LLM request an authorized capability and return the
normalized result.

#### Scenario: Invoke authorized capability

- **WHEN** the LLM requests a capability that is authorized for the context
- **THEN** the server runs the execution security chain
- **AND** returns the normalized capability result

#### Scenario: Invoke unauthorized capability

- **WHEN** the LLM requests a capability not authorized for the context
- **THEN** the server rejects the request with an authorization result
- **AND** does NOT invoke the capability

### Requirement: Tool capability filtering

The MCP server SHALL expose only capabilities allowed by tenant, workflow, state, and
policy.

#### Scenario: No full registry exposure

- **WHEN** the server presents available capabilities to the LLM
- **THEN** it includes only those allowed by the resolved binding
- **AND** never exposes the complete global MCP registry
