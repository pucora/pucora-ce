#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=common.sh
source "$DIR/common.sh"

wait_for "http://127.0.0.1:8081/health" "mock JWT"
wait_grpc

echo "==> gRPC JWT smoke (unauthenticated should fail)"
set +e
unauth=$(call_find_flight 2>&1)
code=$?
set -e
if [[ $code -eq 0 ]]; then
  echo "expected unauthenticated call to fail, got: $unauth"
  exit 1
fi
echo "OK: unauthenticated rejected"

token_json=$(curl -fsS "http://127.0.0.1:8081/token")
token=$(echo "$token_json" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [[ -z "$token" ]]; then
  echo "failed to fetch token: $token_json"
  exit 1
fi

echo "==> gRPC JWT smoke (authenticated)"
resp=$(call_find_flight -H "authorization: Bearer ${token}")
echo "$resp" | grep -q "FL-001" || { echo "expected FL-001: $resp"; exit 1; }
echo "OK: $resp"
