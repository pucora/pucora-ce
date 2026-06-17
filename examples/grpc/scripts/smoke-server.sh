#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=common.sh
source "$DIR/common.sh"

wait_grpc
echo "==> gRPC server smoke"
"$GRPCURL" -plaintext -protoset "$PROTOSET" "$GATEWAY" list | grep -q "flight_finder.Flights" || {
  echo "expected flight_finder.Flights in service list"
  exit 1
}
resp=$(call_find_flight)
echo "$resp" | grep -q "FL-001" || { echo "expected FL-001 in gRPC response: $resp"; exit 1; }
echo "OK: $resp"
