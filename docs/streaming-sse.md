# HTTP Streaming & Server-Sent Events (SSE)

Velonetics supports transparent HTTP streaming and Server-Sent Events (SSE) using the same **no-op proxy** model as [KrakenD Enterprise](https://www.krakend.io/docs/enterprise/endpoints/streaming/). The gateway forwards response chunks as they arrive without buffering or transforming the payload.

Unlike WebSockets, no separate module or `extra_config` namespace is required — streaming is enabled through endpoint and backend encoding settings.

## When to use streaming vs WebSockets

| | SSE / HTTP streaming | WebSockets |
|---|---------------------|------------|
| Direction | Server → client | Bidirectional |
| Protocol | Plain HTTP | RFC-6455 upgrade |
| Config | `no-op` encoding | `extra_config.websocket` |
| Complexity | Lower | Higher |
| Best for | Notifications, LLM tokens, logs, video chunks | Chat, gaming, binary bidirectional |

## Quick start

**1. Run the example stack:**

```bash
make sse-compose-test
```

**2. Or manually:**

```bash
cd examples/streaming
docker compose up --build
curl -N http://localhost:8080/events
```

You should see SSE events arrive incrementally (`event-1`, `event-2`, …) rather than all at once when the stream ends.

## Configuration

Streaming endpoints use standard endpoint/backend settings. Three requirements:

1. **`output_encoding: "no-op"`** on the endpoint
2. **`encoding: "no-op"`** on the backend
3. **Long `timeout` on the endpoint** (not at service level)

### SSE example

```json
{
  "version": 3,
  "port": 8080,
  "write_timeout": "0s",
  "endpoints": [
    {
      "endpoint": "/weather-stream",
      "method": "POST",
      "timeout": "300s",
      "output_encoding": "no-op",
      "input_headers": ["Content-Type"],
      "backend": [
        {
          "host": ["https://weather.example.com"],
          "url_pattern": "/api/agents/weatherAgent/stream",
          "encoding": "no-op"
        }
      ]
    }
  ]
}
```

### Generic HTTP streaming (video)

```json
{
  "endpoint": "/video/{id}",
  "method": "GET",
  "timeout": "5m",
  "output_encoding": "no-op",
  "input_headers": ["Content-Type"],
  "backend": [
    {
      "host": ["https://videos.example.com"],
      "url_pattern": "/video/{id}.mkv",
      "encoding": "no-op"
    }
  ]
}
```

### LLM / Gemini SSE

```json
{
  "endpoint": "/stream",
  "method": "POST",
  "timeout": "300s",
  "output_encoding": "no-op",
  "input_headers": ["Content-Type"],
  "backend": [
    {
      "host": ["https://generativelanguage.googleapis.com"],
      "url_pattern": "/v1beta/models/gemini-2.0-flash:streamGenerateContent?alt=sse&key=xxx",
      "method": "POST",
      "encoding": "no-op"
    }
  ]
}
```

## Configuration reference

| Setting | Level | Description |
|---------|-------|-------------|
| `output_encoding` | Endpoint | Must be `"no-op"` |
| `encoding` | Backend | Must be `"no-op"` |
| `timeout` | Endpoint | Max stream duration (e.g. `"300s"`, `"5m"`) |
| `input_headers` | Endpoint | Forward headers like `Content-Type` to backend |
| `write_timeout` | Service | Set `"0s"` when streaming endpoints exist |
| `max_shutdown_wait_time` | Service | Grace period before force-killing streams on shutdown |

**Important:** Keep the root service `timeout` short. Set long timeouts only on streaming endpoints.

## How it works

```
Client ──persistent HTTP──► Velonetics ──persistent HTTP──► Backend
         chunks forwarded as-is, connection stays open
```

Velonetics uses `NoOpHTTPResponseParser` to pipe the backend response body directly to the client via flush-aware streaming copy. Request validation (JWT, rate limits, IP filtering) still applies before the proxy runs.

## Compatible middleware

| Component | Streaming-safe? |
|-----------|-----------------|
| JWT (`auth/validator`) | Yes (request phase) |
| Rate limiting | Yes |
| Bot detector | Yes |
| CORS | Yes |
| Lua post-response | **No** |
| Response JSON schema | **No** |
| Response body modifiers | **No** |
| Multi-backend merge | **No** |

Run `velonetics audit` to catch common misconfigurations (rules 5.2.4–5.2.6).

## Operational considerations

- **Connection persistence:** Streaming connections are long-lived. Ensure load balancers and reverse proxies disable response buffering and allow idle connections.
- **No payload processing:** Backends must emit valid stream formats (`text/event-stream` for SSE).
- **Redeployments:** Long sessions delay shutdown. Use `max_shutdown_wait_time` to cap graceful drain.
- **Client reconnection:** Velonetics does not auto-reconnect clients; implement retry logic in the client.
- **Resource usage:** Each stream holds open connections on the gateway and backend.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| Events arrive all at once at end | Proxy or LB buffering | Ensure `no-op` encoding; set `write_timeout: "0s"`; disable LB buffering |
| Connection drops early | Timeout too short | Increase endpoint `timeout`; check service `write_timeout` |
| Empty response | Wrong backend encoding | Set backend `encoding: "no-op"` |
| 401 before stream | JWT on protected endpoint | Pass valid `Authorization` header |

## See also

- [examples/streaming](../examples/streaming/) — Docker Compose demo
- [WebSockets](websockets.md) — bidirectional real-time
- [KrakenD streaming docs](https://www.krakend.io/docs/enterprise/endpoints/streaming/)
