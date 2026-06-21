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

echo "==> Waiting for Pucora gateway (WebSocket direct probe)"
for i in $(seq 1 90); do
  if go run ./smoke/main.go -url "ws://127.0.0.1:8080/ws/echo" -message "ready" -timeout 3s >/dev/null 2>&1; then
    break
  fi
  sleep 2
  if [[ "$i" -eq 90 ]]; then
    echo "gateway WebSocket not ready" >&2
    exit 1
  fi
done

echo "==> Fetching test JWT"
TOKEN_JSON="$(curl -fsS "http://127.0.0.1:8081/token")"
TOKEN="$(printf '%s' "$TOKEN_JSON" | python3 -c 'import json,sys; print(json.load(sys.stdin)["token"])')"

echo "==> Direct WebSocket echo (via gateway)"
go run ./smoke/main.go \
  -url "ws://127.0.0.1:8080/ws/echo" \
  -message "direct-ping"

echo "==> JWT-protected WebSocket echo"
go run ./smoke/main.go \
  -url "ws://127.0.0.1:8080/ws/secure" \
  -message "secure-ping" \
  -header "Authorization: Bearer ${TOKEN}"

echo "==> Multiplex WebSocket"
go run ./smoke/main.go \
  -url "ws://127.0.0.1:8080/ws/chat/general" \
  -message "hello-room" \
  -expect-prefix "echo:"

echo "All WebSocket smoke checks passed."
