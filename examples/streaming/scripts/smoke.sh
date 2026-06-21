#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "==> Waiting for mock backend"
for i in $(seq 1 30); do
  if curl -fsS "http://127.0.0.1:8081/health" >/dev/null 2>&1; then
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
  if curl -fsS "http://127.0.0.1:8080/__health" >/dev/null 2>&1; then
    break
  fi
  sleep 2
  if [[ "$i" -eq 60 ]]; then
    echo "gateway not ready" >&2
    exit 1
  fi
done

echo "==> SSE incremental delivery through gateway"
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

(
  curl -fsS -N --max-time 8 "http://127.0.0.1:8080/events" >"$TMP" &
  CURL_PID=$!
  sleep 2
  if ! grep -q 'event-1' "$TMP" 2>/dev/null; then
    echo "first SSE event not received within 2s (buffering or proxy failure)" >&2
    kill "$CURL_PID" 2>/dev/null || true
    exit 1
  fi
  wait "$CURL_PID" || true
)

if ! grep -qi 'text/event-stream' "$TMP" && ! grep -q 'event-1' "$TMP"; then
  echo "SSE response missing expected content (got: $(head -c 200 "$TMP"))" >&2
  exit 1
fi

EVENT_COUNT="$(grep -c '^data:' "$TMP" || true)"
if [[ "$EVENT_COUNT" -lt 3 ]]; then
  echo "expected at least 3 SSE events, got ${EVENT_COUNT}" >&2
  exit 1
fi

echo "==> Chunked stream through gateway"
CHUNK_TMP="$(mktemp)"
trap 'rm -f "$TMP" "$CHUNK_TMP"' EXIT

(
  curl -fsS -N --max-time 5 "http://127.0.0.1:8080/chunked" >"$CHUNK_TMP" &
  CURL_PID=$!
  sleep 1
  if ! grep -q 'alpha-' "$CHUNK_TMP" 2>/dev/null; then
    echo "first chunk not received within 1s" >&2
    kill "$CURL_PID" 2>/dev/null || true
    exit 1
  fi
  wait "$CURL_PID" || true
)

if ! grep -q 'charlie-' "$CHUNK_TMP"; then
  echo "chunked stream incomplete: $(cat "$CHUNK_TMP")" >&2
  exit 1
fi

echo "All HTTP streaming / SSE smoke checks passed."
