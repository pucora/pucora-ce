#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "==> Waiting for mock GraphQL backend"
for i in $(seq 1 30); do
  if curl -fsS "http://127.0.0.1:4000/health" >/dev/null 2>&1; then
    break
  fi
  sleep 1
  if [[ "$i" -eq 30 ]]; then
    echo "mock backend not ready" >&2
    exit 1
  fi
done

echo "==> Waiting for Pucora gateway"
for i in $(seq 1 60); do
  if curl -fsS -o /dev/null -w "%{http_code}" -X POST "http://127.0.0.1:8080/graphql" \
    -H "Content-Type: application/json" \
    -d '{"query":"{ ping }"}' | grep -q 200; then
    break
  fi
  sleep 2
  if [[ "$i" -eq 60 ]]; then
    echo "gateway not ready" >&2
    exit 1
  fi
done

echo "==> Mode 1: REST to GraphQL mutation adapter"
RESP="$(curl -fsS -X POST "http://127.0.0.1:8080/review/1500" \
  -H "Content-Type: application/json" \
  -d '{"review":{"stars":5,"commentary":"great movie"}}')"
echo "$RESP" | grep -q '"stars":5'

echo "==> Mode 1: REST to GraphQL query adapter (GET)"
RESP="$(curl -fsS "http://127.0.0.1:8080/hero/JEDI")"
echo "$RESP" | grep -q '"name":"Luke"'

echo "==> Mode 2: GraphQL proxy passthrough"
RESP="$(curl -fsS -X POST "http://127.0.0.1:8080/graphql" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ ping }","operationName":"Ping"}')"
echo "$RESP" | grep -q '"ping":"pong"'

echo "==> Mode 3: GraphQL federation (parallel subgraphs)"
RESP="$(curl -fsS "http://127.0.0.1:8080/user-data/42")"
echo "$RESP" | grep -q '"user"'
echo "$RESP" | grep -q '"user_metadata"'
echo "$RESP" | grep -q '"Alice"'
echo "$RESP" | grep -q '"en-US"'

echo "All GraphQL smoke checks passed."
