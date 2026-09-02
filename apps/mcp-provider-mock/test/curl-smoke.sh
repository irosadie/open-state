#!/usr/bin/env bash

set -euo pipefail

app_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
bun_bin="$(command -v bun)"
base_port="${MCP_PROVIDER_MOCK_PORT:-$((20000 + RANDOM % 10000))}"
log_file="$(mktemp)"
server_pid=""

cleanup() {
  if [[ -n "$server_pid" ]]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  rm -f "$log_file"
}
trap cleanup EXIT

post_mcp() {
  local port="$1"
  local payload="$2"
  curl --fail --silent --show-error \
    -H "Accept: application/json, text/event-stream" \
    -H "Content-Type: application/json" \
    -H "MCP-Protocol-Version: 2025-11-25" \
    --data "$payload" \
    "http://127.0.0.1:${port}/mcp"
}

call_tool() {
  local port="$1"
  local tool="$2"
  local arguments="$3"
  local request
  request="$(jq -nc --arg tool "$tool" --argjson arguments "$arguments" '{jsonrpc:"2.0",id:3,method:"tools/call",params:{name:$tool,arguments:$arguments}}')"
  post_mcp "$port" "$request"
}

wait_until_ready() {
  local port="$1"
  for _ in $(seq 1 80); do
    if curl --fail --silent "http://127.0.0.1:${port}/health/ready" >/dev/null; then
      return
    fi
    sleep 0.05
  done

  cat "$log_file" >&2
  return 1
}

run_scenario() {
  local scenario="$1"
  local port="$2"
  local tool="$3"
  local assertion="$4"
  local arguments="${5:-}"
  if [[ -z "$arguments" ]]; then
    arguments="{}"
  fi

  MCP_PROVIDER_MOCK_PORT="$port" \
    MCP_PROVIDER_MOCK_SCENARIO="$app_dir/fixtures/$scenario.json" \
    "$bun_bin" "$app_dir/src/index.ts" >"$log_file" 2>&1 &
  server_pid="$!"
  wait_until_ready "$port"

  post_mcp "$port" '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"curl-smoke","version":"1.0.0"}}}' \
    | jq -e '.result.serverInfo.name != ""' >/dev/null
  post_mcp "$port" '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' \
    | jq -e --arg tool "$tool" '.result.tools | any(.name == $tool)' >/dev/null
  call_tool "$port" "$tool" "$arguments" \
    | jq -e "$assertion" >/dev/null

  case "$scenario" in
    padel)
      booking="$(call_tool "$port" "padel.create_booking" '{"venue_id":"padel-senayan","date":"2026-09-01","time":"18:00"}')"
      jq -e '.result.structuredContent.bookingReference == "PAD-0001"' <<<"$booking" >/dev/null
      payment="$(call_tool "$port" "padel.payment.create" '{"booking_id":"PAD-0001"}')"
      payment_id="$(jq -r '.result.structuredContent.payment_id' <<<"$payment")"
      call_tool "$port" "padel.payment.verify" "$(jq -nc --arg payment_id "$payment_id" '{payment_id:$payment_id}')" \
        | jq -e '.result.structuredContent.status == "PAID"' >/dev/null
      call_tool "$port" "padel.court.availability" '{"venue_id":"padel-senayan","date":"2026-09-01"}' \
        | jq -e '.result.structuredContent.available_slots | map(.time) == ["19:00"]' >/dev/null
      call_tool "$port" "padel.notification.send" '{"booking_id":"PAD-0001","message":"Your padel booking is confirmed."}' \
        | jq -e '.result.structuredContent.status == "DELIVERED"' >/dev/null
      ;;
    food-order)
      cart="$(call_tool "$port" "food.cart.add" '{"menu_id":"menu-001","quantity":2}')"
      cart_id="$(jq -r '.result.structuredContent.cart_id' <<<"$cart")"
      call_tool "$port" "food.cart.add" "$(jq -nc --arg cart_id "$cart_id" '{cart_id:$cart_id,menu_id:"menu-004",quantity:1}')" >/dev/null
      call_tool "$port" "food.cart.get" "$(jq -nc --arg cart_id "$cart_id" '{cart_id:$cart_id}')" \
        | jq -e '.result.structuredContent.total == 78000' >/dev/null
      order="$(call_tool "$port" "food.order.create" "$(jq -nc --arg cart_id "$cart_id" '{cart_id:$cart_id,delivery_address:"Sudirman Street No. 1, Jakarta"}')")"
      order_id="$(jq -r '.result.structuredContent.order_id' <<<"$order")"
      payment="$(call_tool "$port" "food.payment.create" "$(jq -nc --arg order_id "$order_id" '{order_id:$order_id}')")"
      payment_id="$(jq -r '.result.structuredContent.payment_id' <<<"$payment")"
      call_tool "$port" "food.payment.verify" "$(jq -nc --arg payment_id "$payment_id" '{payment_id:$payment_id}')" \
        | jq -e '.result.structuredContent.status == "PAID"' >/dev/null
      call_tool "$port" "food.order.track" "$(jq -nc --arg order_id "$order_id" '{order_id:$order_id}')" \
        | jq -e '.result.structuredContent.status == "ON_DELIVERY"' >/dev/null
      ;;
    doctor)
      reservation="$(call_tool "$port" "booking.reserve" '{"schedule_id":"sch-001"}')"
      reservation_id="$(jq -r '.result.structuredContent.reservation_id' <<<"$reservation")"
      appointment="$(call_tool "$port" "booking.confirm" "$(jq -nc --arg reservation_id "$reservation_id" '{reservation_id:$reservation_id}')")"
      booking_id="$(jq -r '.result.structuredContent.booking_id' <<<"$appointment")"
      payment="$(call_tool "$port" "doctor.payment.create" "$(jq -nc --arg booking_id "$booking_id" '{booking_id:$booking_id}')")"
      payment_id="$(jq -r '.result.structuredContent.payment_id' <<<"$payment")"
      call_tool "$port" "doctor.payment.verify" "$(jq -nc --arg payment_id "$payment_id" '{payment_id:$payment_id}')" \
        | jq -e '.result.structuredContent.status == "PAID"' >/dev/null
      call_tool "$port" "doctor.notification.send" "$(jq -nc --arg booking_id "$booking_id" '{booking_id:$booking_id,message:"Your doctor appointment is confirmed."}')" \
        | jq -e '.result.structuredContent.status == "DELIVERED"' >/dev/null
      call_tool "$port" "doctor.schedule" '{}' \
        | jq -e '.result.structuredContent.schedules[0].available == false' >/dev/null
      call_tool "$port" "booking.cancel" "$(jq -nc --arg booking_id "$booking_id" '{booking_id:$booking_id}')" >/dev/null
      call_tool "$port" "doctor.schedule" '{}' \
        | jq -e '.result.structuredContent.schedules[0].available == true' >/dev/null
      ;;
    doctor-happy)
      schedule="$(call_tool "$port" "schedule.check" '{}')"
      schedule_id="$(jq -r '.result.structuredContent.schedule_id' <<<"$schedule")"
      jq -e '.result.structuredContent.schedule_available == true' <<<"$schedule" >/dev/null
      call_tool "$port" "queue.check" '{}' \
        | jq -e '.result.structuredContent.queue_available == true' >/dev/null

      reservation="$(call_tool "$port" "booking.reserve" "$(jq -nc --arg schedule_id "$schedule_id" '{schedule_id:$schedule_id}')")"
      reservation_id="$(jq -r '.result.structuredContent.reservation_id' <<<"$reservation")"
      appointment="$(call_tool "$port" "booking.confirm" "$(jq -nc --arg reservation_id "$reservation_id" '{reservation_id:$reservation_id}')")"
      booking_id="$(jq -r '.result.structuredContent.booking_id' <<<"$appointment")"
      payment="$(call_tool "$port" "doctor.payment.create" "$(jq -nc --arg booking_id "$booking_id" '{booking_id:$booking_id}')")"
      payment_id="$(jq -r '.result.structuredContent.payment_id' <<<"$payment")"
      call_tool "$port" "doctor.payment.verify" "$(jq -nc --arg payment_id "$payment_id" '{payment_id:$payment_id}')" \
        | jq -e '.result.structuredContent.status == "PAID"' >/dev/null
      call_tool "$port" "doctor.notification.send" "$(jq -nc --arg booking_id "$booking_id" '{booking_id:$booking_id,message:"Your doctor appointment is confirmed."}')" \
        | jq -e '.result.structuredContent.status == "DELIVERED"' >/dev/null
      call_tool "$port" "booking.cancel" "$(jq -nc --arg booking_id "$booking_id" '{booking_id:$booking_id}')" \
        | jq -e '.result.structuredContent.status == "CANCELLED"' >/dev/null
      call_tool "$port" "doctor.schedule" '{}' \
        | jq -e --arg schedule_id "$schedule_id" '.result.structuredContent.schedules | any(.[]; .schedule_id == $schedule_id and .available == true)' >/dev/null
      ;;
  esac

  cleanup
  trap cleanup EXIT
  server_pid=""
}

run_timeout_scenario() {
  local port="$1"
  local request

  MCP_PROVIDER_MOCK_PORT="$port" \
    MCP_PROVIDER_MOCK_SCENARIO="$app_dir/fixtures/doctor-timeout.json" \
    "$bun_bin" "$app_dir/src/index.ts" >"$log_file" 2>&1 &
  server_pid="$!"
  wait_until_ready "$port"

  post_mcp "$port" '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"curl-smoke","version":"1.0.0"}}}' \
    | jq -e '.result.serverInfo.name != ""' >/dev/null
  post_mcp "$port" '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' \
    | jq -e '.result.tools | any(.name == "schedule.check")' >/dev/null

  request="$(jq -nc '{jsonrpc:"2.0",id:3,method:"tools/call",params:{name:"schedule.check",arguments:{}}}')"
  if curl --max-time 0.05 --silent --show-error \
    -H "Accept: application/json, text/event-stream" \
    -H "Content-Type: application/json" \
    -H "MCP-Protocol-Version: 2025-11-25" \
    --data "$request" \
    "http://127.0.0.1:${port}/mcp" >/dev/null; then
    echo "doctor timeout scenario unexpectedly completed within the client timeout" >&2
    return 1
  fi

  sleep 0.3
  call_tool "$port" "schedule.check" '{}' \
    | jq -e '.result.structuredContent.schedule_id == "sch-timeout-001"' >/dev/null

  cleanup
  trap cleanup EXIT
  server_pid=""
}

run_state_gateway_scenario() {
  if [[ -z "${STATE_MCP_TOKEN:-}" || -z "${STATE_MCP_TENANT:-}" || -z "${STATE_MCP_PROJECT:-}" || -z "${STATE_MCP_PROVIDER_ALIAS:-}" || -z "${STATE_MCP_INSTANCE:-}" || -z "${STATE_MCP_CAPABILITY:-}" ]]; then
    echo "Skipping optional State MCP gateway smoke test; set STATE_MCP_TOKEN, STATE_MCP_TENANT, STATE_MCP_PROJECT, STATE_MCP_PROVIDER_ALIAS, STATE_MCP_INSTANCE, and STATE_MCP_CAPABILITY to enable it."
    return
  fi

  local state_url="${STATE_MCP_URL:-http://127.0.0.1:8030}"
  local state_token="$STATE_MCP_TOKEN"
  local state_tenant="$STATE_MCP_TENANT"
  local state_project="$STATE_MCP_PROJECT"
  local state_provider_alias="$STATE_MCP_PROVIDER_ALIAS"
  local state_instance="$STATE_MCP_INSTANCE"
  local state_capability="$STATE_MCP_CAPABILITY"
  local state_payload="${STATE_MCP_PAYLOAD_JSON:-}"
  local response
  local arguments
  if [[ -z "$state_payload" ]]; then
    state_payload="{}"
  fi

  state_call() {
    curl --fail --silent --show-error \
      -H "Authorization: Bearer ${state_token}" \
      -H "Accept: application/json, text/event-stream" \
      -H "Content-Type: application/json" \
      -H "MCP-Protocol-Version: 2025-11-25" \
      --data "$1" \
      "${state_url}/mcp"
  }

  response="$(state_call '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"doctor-provider-curl-smoke","version":"1.0.0"}}}')"
  jq -e '.result.serverInfo.name == "openstate" and (.result.instructions | contains("enforced gateway"))' <<<"$response" >/dev/null
  state_call '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' \
    | jq -e '.result.tools | any(.name == "invoke_capability") and all(.name != "report_capability_result")' >/dev/null

  arguments="$(jq -nc --arg project "$state_project" '{project:$project}')"
  state_call "$(jq -nc --argjson arguments "$arguments" '{jsonrpc:"2.0",id:3,method:"tools/call",params:{name:"list_intents",arguments:$arguments}}')" \
    | jq -e '.result.content[0].text | fromjson' >/dev/null

  arguments="$(jq -nc --arg instance "$state_instance" --arg capability "$state_capability" --argjson payload "$state_payload" '{instance:$instance,capability:$capability,correlationId:"doctor-provider-curl-smoke",idempotencyKey:("doctor-provider-curl-smoke-" + $instance + "-" + $capability),payload:$payload}')"
  response="$(state_call "$(jq -nc --argjson arguments "$arguments" '{jsonrpc:"2.0",id:4,method:"tools/call",params:{name:"invoke_capability",arguments:$arguments}}')")"
  jq -e '.result.content[0].text | fromjson | .ok == true and .invoked == true' <<<"$response" >/dev/null
  echo "State MCP gateway smoke passed for ${state_capability} (${state_tenant}/${state_project}, provider ${state_provider_alias})."
}

smoke_domain="${MCP_PROVIDER_MOCK_SMOKE_DOMAIN:-all}"
case "$smoke_domain" in
  all)
    run_scenario "padel" "$base_port" "padel.court.search" '.result.structuredContent.courts[0].name == "GOR Senayan Court A"'
    run_scenario "food-order" "$((base_port + 1))" "food.menu.list" '.result.structuredContent.menu[0].name == "Special Fried Rice"'
    ;;
  doctor)
    ;;
  *)
    echo "MCP_PROVIDER_MOCK_SMOKE_DOMAIN must be all or doctor" >&2
    exit 1
    ;;
esac

run_scenario "doctor" "$((base_port + 2))" "doctor.search" '.result.structuredContent.doctors[0].specialization == "Internal Medicine"'
run_scenario "doctor-happy" "$((base_port + 3))" "doctor.lookup" '.result.structuredContent.doctor.id == "doc-001"'
run_scenario "doctor-no-results" "$((base_port + 4))" "doctor.search" '.result.structuredContent.doctors | length == 0'
run_scenario "doctor-unavailable" "$((base_port + 5))" "schedule.check" '.result.structuredContent.schedule_available == false'
run_scenario "doctor-queue-full" "$((base_port + 6))" "queue.check" '.result.structuredContent.queue_available == false'
run_scenario "doctor-conflict" "$((base_port + 7))" "booking.reserve" '.result.isError == true and ((.result.content[0].text | fromjson).code == "schedule_conflict")' '{"schedule_id":"sch-conflict-001"}'
run_scenario "doctor-payment-failed" "$((base_port + 8))" "doctor.payment.create" '.result.isError == true and ((.result.content[0].text | fromjson).code == "payment_declined")' '{"booking_id":"BKGD-0001"}'
run_scenario "doctor-notification-failed" "$((base_port + 9))" "doctor.notification.send" '.result.isError == true and ((.result.content[0].text | fromjson).code == "notification_delivery_failed")' '{"booking_id":"BKGD-0001","message":"Your doctor appointment is confirmed."}'
run_scenario "doctor-provider-error" "$((base_port + 10))" "doctor.lookup" '.result.isError == true and ((.result.content[0].text | fromjson).code == "doctor_catalog_unavailable")'
run_scenario "doctor-invalid-output" "$((base_port + 11))" "doctor.lookup" '.result.structuredContent.doctor == "this should be an object"'
run_timeout_scenario "$((base_port + 12))"
run_state_gateway_scenario

printf 'MCP curl smoke tests passed.\n'
