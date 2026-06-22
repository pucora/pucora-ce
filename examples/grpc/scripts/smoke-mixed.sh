#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=common.sh
source "$DIR/common.sh"

wait_for "http://127.0.0.1:8080/__health" "gateway"
wait_grpc
echo "==> gRPC mixed mode smoke"
resp=$(curl -fsS "http://127.0.0.1:8080/flights?lat=40.7&lon=-74.0")
echo "$resp" | grep -q "FL-001" || { echo "REST client failed: $resp"; exit 1; }
echo "OK REST: $resp"
grpc_resp=$(call_find_flight)
echo "$grpc_resp" | grep -q "FL-001" || { echo "gRPC server failed: $grpc_resp"; exit 1; }
echo "OK gRPC: $grpc_resp"
