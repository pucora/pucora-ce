# HTTP Streaming & SSE Example

Local Docker Compose stack demonstrating KrakenD-parity HTTP streaming through Velonetics using `no-op` encoding.

## Quick start

```bash
cd examples/streaming
docker compose up --build
```

In another terminal:

```bash
curl -N http://localhost:8080/events
curl -N http://localhost:8080/chunked
```

Or from the CE repo root:

```bash
make sse-compose-test
```

## Endpoints

| Gateway path | Backend | Description |
|--------------|---------|-------------|
| `GET /events` | `GET /events` | SSE (`text/event-stream`), 6 events every 500ms |
| `GET /chunked` | `GET /chunked` | Raw byte stream with 300ms between chunks |

## Configuration

Streaming requires:

1. `output_encoding: "no-op"` on the endpoint
2. `encoding: "no-op"` on the backend
3. Long `timeout` on the **endpoint** (not service level)
4. `write_timeout: "0s"` at service level for long-lived connections

See [docs/streaming-sse.md](../../docs/streaming-sse.md) for full documentation.

## Phase 1 baseline notes

| Check | Status |
|-------|--------|
| no-op proxy pipes backend body | Works (pre-existing Lura path) |
| SSE incremental delivery | Required flush-aware `StreamCopy` (Phase 2) |
| Endpoint timeout closes stream | Works via request context |
| Service `write_timeout` | Must be `0s` for streams longer than default |
| Multi-backend + no-op | Incompatible — audit rule 5.2.4 |
| Response Lua / modifiers | Incompatible — audit rule 5.2.6 |
