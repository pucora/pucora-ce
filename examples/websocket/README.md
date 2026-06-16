# WebSocket local stack (Docker Compose)

Runs Velonetics CE with a mock backend that provides:

- **HTTP** `GET /jwk/symmetric` — HS256 JWK for JWT validation
- **HTTP** `GET /token` — sample bearer token for smoke tests
- **WebSocket** `/echo` — direct echo (used by `/ws/echo` and `/ws/secure`)
- **WebSocket** `/ws` — multiplex backend with Velonetics handshake + envelope replies

## Quick start

From the repository root:

```bash
make ws-compose-up
```

`mock-backend` is vendored for offline Docker builds. Re-vendor after dependency changes:

```bash
cd examples/websocket/mock-backend && go mod vendor
```

In another terminal:

```bash
make ws-compose-smoke
```

Or manually:

```bash
cd examples/websocket
docker compose up --build
./scripts/smoke.sh
```

## Endpoints

| Gateway URL | Mode | Auth |
|-------------|------|------|
| `ws://localhost:8080/ws/echo` | Direct echo | None |
| `ws://localhost:8080/ws/chat/{room}` | Multiplex | None |
| `ws://localhost:8080/ws/secure` | Direct echo | JWT (`auth/validator`) |

### Get a test JWT

```bash
curl -s http://127.0.0.1:8081/token | jq .
```

Use the `value` field as the `Authorization` header on the WebSocket upgrade.

### Manual test with websocat

```bash
# Direct
websocat ws://127.0.0.1:8080/ws/echo

# JWT (replace TOKEN)
TOKEN=$(curl -s http://127.0.0.1:8081/token | jq -r .token)
websocat -H="Authorization: Bearer $TOKEN" ws://127.0.0.1:8080/ws/secure
```

## Files

| File | Purpose |
|------|---------|
| `docker-compose.yml` | Gateway + mock backend |
| `velonetics.json` | Gateway config (service DNS names) |
| `mock-backend/` | Echo + multiplex + JWK server |
| `smoke/` | Tiny Go WebSocket client for CI/local checks |
| `scripts/smoke.sh` | End-to-end smoke script |

## Tear down

```bash
cd examples/websocket && docker compose down
```

## Related docs

- [docs/websockets.md](../../docs/websockets.md) — full WebSocket configuration reference
- [tests/fixtures/ws_jwt.json](../../tests/fixtures/ws_jwt.json) — JWT fixture for `127.0.0.1` (non-Docker)
