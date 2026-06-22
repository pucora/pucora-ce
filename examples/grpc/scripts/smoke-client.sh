#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=common.sh
source "$DIR/common.sh"

wait_for "http://127.0.0.1:8080/__health" "gateway"

echo "==> REST gRPC client smoke"
resp=$(curl -fsS "http://127.0.0.1:8080/flights?lat=40.7&lon=-74.0")
echo "$resp" | grep -q "FL-001" || { echo "expected flight id in response: $resp"; exit 1; }
echo "OK: $resp"
