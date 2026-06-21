#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "==> Waiting for mock SOAP backend"
for i in $(seq 1 30); do
  if curl -fsS "http://127.0.0.1:8081/health" >/dev/null 2>&1; then
    break
  fi
  sleep 1
  if [[ "$i" -eq 30 ]]; then
    echo "mock SOAP backend not ready" >&2
    exit 1
  fi
done

echo "==> Waiting for Pucora gateway"
for i in $(seq 1 90); do
  if curl -fsS "http://127.0.0.1:8080/country/US" >/dev/null 2>&1; then
    break
  fi
  sleep 2
  if [[ "$i" -eq 90 ]]; then
    echo "gateway not ready" >&2
    exit 1
  fi
done

echo "==> SOAP REST conversion smoke test"
RESP="$(curl -fsS "http://127.0.0.1:8080/country/US")"
echo "$RESP"
if ! printf '%s' "$RESP" | python3 -c 'import json,sys; d=json.load(sys.stdin); assert d.get("flag_url") == "http://www.example.com/flags/US.jpg", d'; then
  echo "unexpected response: $RESP" >&2
  exit 1
fi

echo "==> SOAP REST conversion for DE"
RESP="$(curl -fsS "http://127.0.0.1:8080/country/DE")"
if ! printf '%s' "$RESP" | python3 -c 'import json,sys; d=json.load(sys.stdin); assert d.get("flag_url") == "http://www.example.com/flags/DE.jpg", d'; then
  echo "unexpected response: $RESP" >&2
  exit 1
fi

echo "All SOAP smoke checks passed."
