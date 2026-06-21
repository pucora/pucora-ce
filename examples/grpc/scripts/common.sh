#!/usr/bin/env bash
set -euo pipefail

GRPCURL="${GRPCURL:-grpcurl}"
GATEWAY="${GATEWAY:-127.0.0.1:8080}"
REQUEST_JSON='{"where":{"latitude":40.7,"longitude":-74.0}}'
PROTOSET="${PROTOSET:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/mock-backend/flights.pb}"

wait_for() {
  local url="$1"
  local label="$2"
  for _ in $(seq 1 30); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      echo "OK: $label ready"
      return 0
    fi
    sleep 2
  done
  echo "timeout waiting for $label at $url"
  exit 1
}

wait_grpc() {
  for _ in $(seq 1 30); do
    if "$GRPCURL" -plaintext -protoset "$PROTOSET" "$GATEWAY" list >/dev/null 2>&1; then
      echo "OK: gRPC gateway ready"
      return 0
    fi
    sleep 2
  done
  echo "timeout waiting for gRPC on $GATEWAY"
  exit 1
}

call_find_flight() {
  if [ $# -eq 0 ]; then
    "$GRPCURL" -plaintext -protoset "$PROTOSET" -d "$REQUEST_JSON" \
      "$GATEWAY" flight_finder.Flights/FindFlight
    return
  fi
  "$GRPCURL" -plaintext -protoset "$PROTOSET" "$@" -d "$REQUEST_JSON" \
    "$GATEWAY" flight_finder.Flights/FindFlight
}
