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

  MCP_PROVIDER_MOCK_PORT="$port" \
    MCP_PROVIDER_MOCK_SCENARIO="$app_dir/fixtures/$scenario.json" \
    "$bun_bin" "$app_dir/src/index.ts" >"$log_file" 2>&1 &
  server_pid="$!"
  wait_until_ready "$port"

  post_mcp "$port" '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"curl-smoke","version":"1.0.0"}}}' \
    | jq -e '.result.serverInfo.name != ""' >/dev/null
  post_mcp "$port" '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' \
    | jq -e --arg tool "$tool" '.result.tools | any(.name == $tool)' >/dev/null
  call_tool "$port" "$tool" '{}' \
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
  esac

  cleanup
  trap cleanup EXIT
  server_pid=""
}

run_scenario "padel" "$base_port" "padel.court.search" '.result.structuredContent.courts[0].name == "GOR Senayan Court A"'
run_scenario "food-order" "$((base_port + 1))" "food.menu.list" '.result.structuredContent.menu[0].name == "Special Fried Rice"'
run_scenario "doctor" "$((base_port + 2))" "doctor.search" '.result.structuredContent.doctors[0].specialization == "Internal Medicine"'

printf 'MCP curl smoke tests passed.\n'
