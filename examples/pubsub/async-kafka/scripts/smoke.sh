#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

GATEWAY="${GATEWAY:-http://127.0.0.1:8080}"
WEBHOOK="${WEBHOOK:-http://127.0.0.1:8081}"
for i in $(seq 1 90); do
  if curl -fsS "${GATEWAY}/__health" >/dev/null 2>&1; then
    break
  fi
  sleep 2
  if [[ "$i" -eq 90 ]]; then
    echo "gateway not ready" >&2
    exit 1
  fi
done

echo "==> Waiting for mock webhook"
for i in $(seq 1 30); do
  if curl -fsS "${WEBHOOK}/health" >/dev/null 2>&1; then
    break
  fi
  sleep 1
  if [[ "$i" -eq 30 ]]; then
    echo "mock webhook not ready" >&2
    exit 1
  fi
done

echo "==> Producing Kafka message"
docker compose exec -T redpanda rpk topic create events >/dev/null 2>&1 || true
printf '%s\n' '{"event":"async-smoke"}' | docker compose exec -T redpanda rpk topic produce events -k smoke-key >/dev/null

echo "==> Waiting for async agent delivery"
for i in $(seq 1 30); do
  if resp="$(curl -fsS "${WEBHOOK}/last" 2>/dev/null)"; then
    echo "$resp"
    if printf '%s' "$resp" | python3 -c 'import json,sys; d=json.load(sys.stdin); assert d.get("event") == "async-smoke", d'; then
      echo "All async/kafka smoke checks passed."
      exit 0
    fi
  fi
  sleep 1
done
echo "async/kafka smoke test failed" >&2
exit 1
