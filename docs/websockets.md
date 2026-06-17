# WebSockets

Velonetics supports bidirectional communication using the [WebSocket protocol (RFC-6455)](https://datatracker.ietf.org/doc/html/rfc6455). Clients connect to the gateway over WebSocket; the gateway connects to one or more backend hosts using `ws://` or `wss://`.

This feature is implemented by the [`velonetics-websocket`](https://github.com/velonetics/velonetics-websocket) module and is configured per endpoint via `extra_config.websocket`.

## Operating modes

| Mode | Config | Client ↔ Gateway | Gateway ↔ Backend | Best for |
|------|--------|------------------|-------------------|----------|
| **Multiplexing** (default) | `enable_direct_communication: false` | One WebSocket per client | **One shared** WebSocket per endpoint | Chat rooms, fan-out, many clients |
| **Direct** | `enable_direct_communication: true` | One WebSocket per client | **One WebSocket per client** | Transparent proxy, binary streams, subprotocols |

Multiplexing is recommended when many clients talk to the same backend service. Direct mode is simpler but opens more backend connections.

## Quick start (direct echo)

**1. Backend echo server** (example using `websocat` or any WS server on port 8081):

```bash
# any WebSocket echo server listening on ws://127.0.0.1:8081/
```

**2. Gateway config** (`velonetics.json`):

```json
{
  "version": 3,
  "port": 8080,
  "endpoints": [
    {
      "endpoint": "/ws/echo",
      "method": "GET",
      "backend": [
        {
          "host": ["ws://127.0.0.1:8081"],
          "url_pattern": "/",
          "disable_host_sanitize": true
        }
      ],
      "extra_config": {
        "telemetry/usage": { "enabled": false },
        "websocket": {
          "enable_direct_communication": true,
          "max_message_size": 4096
        }
      }
    }
  ]
}
```

**3. Run and test:**

```bash
make build
./velonetics run -c velonetics.json

# In another terminal (requires websocat or similar):
websocat ws://localhost:8080/ws/echo
```

Messages you send are forwarded to the backend as-is and responses are returned unchanged.

## Requirements

Every WebSocket endpoint must satisfy:

1. **`extra_config.websocket`** — present (an empty object `{}` is valid).
2. **Backend `host`** — must use `ws://` or `wss://` (not `http://`).
3. **`disable_host_sanitize: true`** on the backend block.
4. **`method: "GET"`** — WebSocket upgrades use HTTP GET (the router forces GET when `websocket` is configured).

Optional but common:

- **`auth/validator`** — JWT is checked on the **HTTP upgrade request** only (not on each frame).
- **`input_headers`** inside `websocket` — headers forwarded to the backend (see below).

## Configuration reference

All settings live under `endpoints[].extra_config.websocket`:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enable_direct_communication` | bool | `false` | `true` = direct proxy; `false` = multiplexing |
| `input_headers` | `[]string` | `[]` | Headers allowed to the backend. Use `"*"` for all (not recommended). Never forwards `Upgrade`, `Connection`, or `Sec-WebSocket-*` |
| `subprotocols` | `[]string` | `[]` | Subprotocols offered to clients (**direct mode**) |
| `max_message_size` | int | `512` | Max message size in bytes; oversized messages disconnect the client |
| `message_buffer_size` | int | `256` | Per-client outbound queue size (**multiplex**) |
| `max_retries` | int | `0` | Backend reconnect attempts; `0` = unlimited |
| `backoff_strategy` | string | `fallback` | `linear`, `linear-jitter`, `exponential`, `exponential-jitter`, `fallback` |
| `connect_event` | bool | `false` | Notify backend when a client connects (**multiplex**) |
| `disconnect_event` | bool | `false` | Notify backend when a client disconnects (**multiplex**) |
| `return_error_details` | bool | `false` | Send `{"error":"..."}` to client when backend write fails |
| `disable_otel_metrics` | bool | `false` | Disable OpenTelemetry WebSocket metrics |
| `ping_period` | duration | `54s` | Client ping interval |
| `pong_wait` | duration | `60s` | Pong timeout |
| `write_wait` | duration | `10s` | Max time to write a frame to the client |
| `read_buffer_size` | int | `1024` | Reserved for future use; not applied by the current WebSocket driver |
| `write_buffer_size` | int | `1024` | Reserved for future use; not applied by the current WebSocket driver |
| `timeout` | duration | `5m` | Per-read timeout on client and backend connections |

Durations use Go syntax: `500ms`, `10s`, `5m`, `1h`.

## Multiplexing mode

### Architecture

```
 Client A ──WS──┐
 Client B ──WS──┼── Velonetics Gateway ── single WS ── Backend
 Client C ──WS──┘
```

Many clients share one backend WebSocket. The gateway wraps traffic in a JSON **envelope** on the backend leg only. Clients send and receive **plain** WebSocket frames.

### Backend handshake

When the gateway opens the backend connection it sends:

```json
{"msg":"Velonetics WS proxy starting"}
```

The backend **must** reply with the plain text string:

```
OK
```

Until `OK` is received, the gateway will not forward client traffic.

### Envelope format (gateway ↔ backend)

**Client → backend** (gateway wraps each client message):

```json
{
  "url": "/ws/chat/general",
  "session": {
    "uuid": "0b251b07-5611-49e5-b69f-cf2cb8d339d6",
    "Room": "general"
  },
  "body": "SGVsbG8gV29ybGQh"
}
```

| Field | Description |
|-------|-------------|
| `url` | Client request path on the gateway |
| `session` | Always includes `uuid` (assigned by the gateway). Endpoint placeholders appear with the **first letter capitalized** (e.g. `{room}` → `Room`) |
| `body` | Base64-encoded payload from the client |

**Backend → clients** (routing):

| Payload | Delivery |
|---------|----------|
| `{ "body": "<base64>" }` only | **Broadcast** to all clients on the endpoint |
| `{ "url": "/ws/chat/general", "body": "<base64>" }` | **Multicast** to clients on that path |
| `{ "session": { "uuid": "..." }, "body": "<base64>" }` | **Unicast** to one client |

Invalid JSON or unknown shape from the backend is **broadcast** as raw bytes to all clients.

### Example multiplex config

```json
{
  "endpoint": "/ws/{room}",
  "method": "GET",
  "input_query_strings": ["*"],
  "backend": [
    {
      "host": ["ws://127.0.0.1:8081"],
      "url_pattern": "/ws",
      "disable_host_sanitize": true
    }
  ],
  "extra_config": {
    "websocket": {
      "input_headers": ["Authorization", "Cookie"],
      "connect_event": true,
      "disconnect_event": true,
      "max_message_size": 3200000,
      "message_buffer_size": 256,
      "max_retries": 0,
      "backoff_strategy": "exponential"
    }
  }
}
```

### Minimal multiplex backend (Go sketch)

```go
conn, _ := websocket.Accept(w, r, nil)
// 1. Handshake
_, msg, _ := conn.Read(ctx)
// expect: {"msg":"Velonetics WS proxy starting"}
conn.Write(ctx, websocket.MessageText, []byte("OK"))

// 2. Read envelopes from gateway
_, data, _ := conn.Read(ctx)
var env struct {
    URL     string                 `json:"url"`
    Session map[string]interface{} `json:"session"`
    Body    string                 `json:"body"`
}
json.Unmarshal(data, &env)
payload, _ := base64.StdEncoding.DecodeString(env.Body)

// 3. Reply to one client (unicast)
reply, _ := json.Marshal(map[string]interface{}{
    "session": map[string]string{"uuid": env.Session["uuid"].(string)},
    "body":    base64.StdEncoding.EncodeToString([]byte("pong")),
})
conn.Write(ctx, websocket.MessageText, reply)
```

## Direct mode

- No handshake and no envelope — binary and text frames pass through transparently.
- Supports `subprotocols` in config.
- If the backend connection fails, the client connection is closed.
- Set `"enable_direct_communication": true` in `websocket` config.

## Authentication (JWT)

Add `auth/validator` on the same endpoint as for HTTP. The token is validated **once**, on the WebSocket upgrade request (typically `Authorization: Bearer …`).

```json
"extra_config": {
  "auth/validator": {
    "alg": "RS256",
    "jwk_url": "https://your-idp/.well-known/jwks.json",
    "audience": ["your-api"],
    "roles": ["user"]
  },
  "websocket": {
    "input_headers": ["Authorization"]
  }
}
```

After a successful upgrade, frames are not re-validated. Issue short-lived tokens or reconnect when tokens expire.

## Reconnection and backoff

If the backend WebSocket drops, the gateway retries using `max_retries` and `backoff_strategy`:

| Strategy | Delay between retries |
|----------|------------------------|
| `fallback` | 1 second (constant) |
| `linear` | `retry` seconds |
| `linear-jitter` | `retry ± 33%` random |
| `exponential` | `2^retry` seconds |
| `exponential-jitter` | `2^retry ± 33%` random |

- `max_retries: 0` — retry forever (typical for multiplex).
- `max_retries: 1` — one retry, then give up (more common in direct mode).

While the backend is down, client messages may be queued (multiplex, up to `message_buffer_size`). With `return_error_details: true`, clients can receive `{"error":"empty connection"}`.

## Logging

| Level | Meaning |
|-------|---------|
| `WARNING` | Connectivity issues (read errors, backend closed) |
| `ERROR` | Reconnect failures |
| `CRITICAL` | Retries exhausted; backend unavailable |

Log prefix: `[SERVICE: Websocket]`.

## Metrics (OpenTelemetry)

Unless `disable_otel_metrics` is set:

| Metric | Description |
|--------|-------------|
| `velonetics.websocket.connections` | Active client connections |
| `velonetics.websocket.messages.in` | Messages from clients |
| `velonetics.websocket.messages.out` | Messages to clients |
| `velonetics.websocket.reconnects` | Backend reconnect attempts |

## What is not supported

- **Socket.IO** — not compatible with plain RFC-6455 multiplexing. Socket.IO requires direct mode and a dedicated URL pattern; not officially supported in CE.
- **Per-frame JWT** — auth applies at upgrade only.
- **HTTP proxy pipeline** — WebSocket endpoints bypass the normal request/response proxy; backend middleware (Martian, circuit breaker, etc.) does not run on WS traffic.

## Testing

Module tests (sibling repo or `go.work` workspace):

```bash
cd ../velonetics-websocket
go test ./...
```

Or from the repository root (with `go.work` at the workspace parent, see `scripts/init-workspace.sh`):

```bash
make test-websocket
```

Sample configs (validate with `./velonetics check -c …`):

| Fixture | Description |
|---------|-------------|
| [`tests/fixtures/ws_direct.json`](../../tests/fixtures/ws_direct.json) | Direct echo proxy |
| [`tests/fixtures/ws_multiplex.json`](../../tests/fixtures/ws_multiplex.json) | Multiplex chat-style endpoint |
| [`tests/fixtures/ws_jwt.json`](../../tests/fixtures/ws_jwt.json) | JWT-protected direct WebSocket |

```bash
make check-fixtures
```

Validate config:

```bash
./velonetics check -c your-config.json
```

## Docker

### Compose stack (gateway + mock backend)

For a full local stack with direct, multiplex, and JWT-protected endpoints:

```bash
make ws-compose-up      # start mock backend + gateway
make ws-compose-smoke   # run end-to-end WebSocket checks
make ws-compose-down    # tear down
```

See [examples/websocket/README.md](../examples/websocket/README.md).

### Single container

```bash
make docker
docker run -p 8080:8080 \
  -v $(pwd)/velonetics-ws.json:/etc/velonetics/velonetics.json \
  velonetics/velonetics:2.0.0 run -c /etc/velonetics/velonetics.json
```

Ensure the container can reach backend `ws://` hosts (use host networking or service names in Docker Compose).

## Related files

| Path | Purpose |
|------|---------|
| [`velonetics-websocket`](https://github.com/velonetics/velonetics-websocket) | Implementation module (`../velonetics-websocket` in workspace) |
| [`velonetics-schema` v2.13 websocket.json](https://github.com/velonetics/velonetics-schema/blob/v2.0.1/v2.13/websocket.json) | JSON Schema |
| [`handler_factory.go`](../handler_factory.go) | Gateway wiring (WebSocket → JWT handler chain) |
| [`lura` router/gin/router.go](https://github.com/velonetics/lura/blob/v2.0.1/router/gin/router.go) | GET-only registration for WS endpoints |
| [`tests/fixtures/ws_direct.json`](../tests/fixtures/ws_direct.json) | Direct mode sample |
| [`tests/fixtures/ws_multiplex.json`](../tests/fixtures/ws_multiplex.json) | Multiplex sample |
| [`tests/fixtures/ws_jwt.json`](../tests/fixtures/ws_jwt.json) | JWT + direct mode sample |
| [`examples/websocket/`](../examples/websocket/) | Docker Compose stack + smoke tests |
| [`docs/WEBSOCKET_RELEASE.md`](WEBSOCKET_RELEASE.md) | Release notes and publish commands |
