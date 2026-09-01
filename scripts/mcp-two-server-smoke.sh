#!/usr/bin/env bash

set -euo pipefail

state_url="${STATE_MCP_URL:-http://127.0.0.1:8030}"
provider_url="${PROVIDER_MCP_URL:-http://127.0.0.1:8031}"
state_token="${STATE_MCP_TOKEN:?Set STATE_MCP_TOKEN to an osk_... State MCP API key}"
intent="${INTENT:-BOOKING_PADEL}"
instance_id="${INSTANCE_ID:?Set INSTANCE_ID to a running workflow instance}"
event_type="${EVENT_TYPE:?Set EVENT_TYPE to the desired transition event}"
provider_args="${PROVIDER_ARGS_JSON:-{}}"
protocol_version="${MCP_PROTOCOL_VERSION:-2025-06-18}"

state_call() {
  curl --fail --silent --show-error \
    -H "Authorization: Bearer ${state_token}" \
    -H "Accept: application/json, text/event-stream" \
    -H "Content-Type: application/json" \
    -H "MCP-Protocol-Version: ${protocol_version}" \
    --data "$1" \
    "${state_url}/mcp"
}

provider_call() {
  curl --fail --silent --show-error \
    -H "Accept: application/json, text/event-stream" \
    -H "Content-Type: application/json" \
    -H "MCP-Protocol-Version: ${protocol_version}" \
    --data "$1" \
    "${provider_url}/mcp"
}

text_payload() {
  jq -c '.result.content[0].text | fromjson'
}

echo "Checking both MCP identities..."
state_init="$(state_call '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"'"${protocol_version}"'","capabilities":{},"clientInfo":{"name":"two-mcp-smoke","version":"1.0.0"}}}')"
provider_init="$(provider_call '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"'"${protocol_version}"'","capabilities":{},"clientInfo":{"name":"two-mcp-smoke","version":"1.0.0"}}}')"
jq -e '.result.serverInfo.name == "openstate" and (.result.instructions | contains("report_capability_result"))' <<<"${state_init}" >/dev/null
provider_name="$(jq -r '.result.serverInfo.name' <<<"${provider_init}")"
jq -e '.result.serverInfo.name != "openstate"' <<<"${provider_init}" >/dev/null

state_call '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' \
  | jq -e '.result.tools | any(.name == "resolve_intent") and any(.name == "report_capability_result") and any(.name == "propose_event")' >/dev/null
provider_tools="$(provider_call '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}')"

resolved="$(state_call "$(jq -nc --arg intent "${intent}" '{jsonrpc:"2.0",id:3,method:"tools/call",params:{name:"resolve_intent",arguments:{intent:$intent}}}')" | text_payload)"
current="$(state_call "$(jq -nc --arg instance "${instance_id}" '{jsonrpc:"2.0",id:4,method:"tools/call",params:{name:"get_current_state",arguments:{instance:$instance}}}')" | text_payload)"
requirement="$(jq -e '.requiredCapabilities[] | select(.status != "MISSING_MAPPING")' <<<"${current}")"
capability="$(jq -r '.capability' <<<"${requirement}")"
provider_server="$(jq -r '.providerServer' <<<"${requirement}")"
provider_tool="$(jq -r '.tool' <<<"${requirement}")"
jq -e --arg tool "${provider_tool}" '.result.tools | any(.name == $tool)' <<<"${provider_tools}" >/dev/null

blocked="$(state_call "$(jq -nc --arg instance "${instance_id}" --arg type "${event_type}" '{jsonrpc:"2.0",id:5,method:"tools/call",params:{name:"propose_event",arguments:{instance:$instance,type:$type}}}')" | text_payload)"
jq -e '.ok == false and (.message | contains("capability requirement not satisfied"))' <<<"${blocked}" >/dev/null

provider_result="$(provider_call "$(jq -nc --arg tool "${provider_tool}" --argjson arguments "${provider_args}" '{jsonrpc:"2.0",id:6,method:"tools/call",params:{name:$tool,arguments:$arguments}}')")"
normalized_result="$(jq -c '.result.structuredContent // {}' <<<"${provider_result}")"
report_args="$(jq -nc --arg project "${PROJECT_ID:-}" --arg instance "${instance_id}" --arg state "$(jq -r '.stateId // .stateKey' <<<"${current}")" --arg capability "${capability}" --arg server "${provider_server}" --arg tool "${provider_tool}" --arg correlation "${CORRELATION_ID:-two-mcp-smoke}" --arg key "${IDEMPOTENCY_KEY:-two-mcp-smoke-${instance_id}-${capability}}" --argjson result "${normalized_result}" '{project:$project,instance:$instance,state:$state,capability:$capability,providerServer:$server,providerTool:$tool,correlationId:$correlation,idempotencyKey:$key,status:"SUCCEEDED",result:$result} | with_entries(select(.value != ""))')"
bad_report_args="$(jq -c '.providerTool = (.providerTool + ".undeclared")' <<<"${report_args}")"
state_call "$(jq -nc --argjson arguments "${bad_report_args}" '{jsonrpc:"2.0",id:7,method:"tools/call",params:{name:"report_capability_result",arguments:$arguments}}')" \
  | text_payload | jq -e '.ok == false' >/dev/null
reported="$(state_call "$(jq -nc --argjson arguments "${report_args}" '{jsonrpc:"2.0",id:8,method:"tools/call",params:{name:"report_capability_result",arguments:$arguments}}')" | text_payload)"
jq -e '.accepted == true' <<<"${reported}" >/dev/null

transition_args="$(jq -nc --arg instance "${instance_id}" --arg type "${event_type}" '{instance:$instance,type:$type}')"
transitioned="$(state_call "$(jq -nc --argjson arguments "${transition_args}" '{jsonrpc:"2.0",id:9,method:"tools/call",params:{name:"propose_event",arguments:$arguments}}')" | text_payload)"
jq -e '.ok == true' <<<"${transitioned}" >/dev/null

echo "Two-MCP smoke passed: openstate + ${provider_name}; ${capability} used ${provider_server}/${provider_tool}."
