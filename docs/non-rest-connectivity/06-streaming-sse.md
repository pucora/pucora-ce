# HTTP Streaming & Server-Sent Events (SSE)

**Edition:** Enterprise only  
**Namespace:** None — uses standard endpoint/backend config with `no-op` encoding  
**Official docs:** [HTTP Streaming and SSE](https://www.krakend.io/docs/enterprise/endpoints/streaming/)

## What it does

Proxies long-lived HTTP connections where the backend streams data chunks over time. SSE (Server-Sent Events) is a subtype using `text/event-stream` for unidirectional server→client push.

## Key capabilities

- Transparent streaming proxy (no buffering or transformation)
- Works for video, logs, LLM token streams, progressive loading, long-polling
- SSE support for notifications, live feeds, AI generative models
- Request validation (auth, IP filtering) still applies

## Limitations

- **No response manipulation** — KrakenD acts as proxy only
- No merging of multiple backend responses
- Client must handle reconnection (KrakenD does not auto-reconnect clients)

## How it works

```
Client ──persistent HTTP──► KrakenD ──persistent HTTP──► Backend
         chunks forwarded as-is, connection stays open
```

Unlike regular REST endpoints, the connection remains open until timeout or client disconnect.

## Configuration

### HTTP streaming

Three requirements:

1. `output_encoding: "no-op"` on endpoint
2. `encoding: "no-op"` on backend
3. Long `timeout` on the **endpoint** (not service level)

```json
{
  "endpoint": "/video/{id}",
  "method": "GET",
  "timeout": "5m",
  "output_encoding": "no-op",
  "input_headers": ["Content-Type"],
  "backend": [{
    "url_pattern": "/video/{id}.mkv",
    "encoding": "no-op",
    "host": ["https://videos.example.com/"]
  }]
}
```

### SSE example

SSE uses the same pattern. Pass `Content-Type` and set appropriate timeout:

```json
{
  "endpoint": "/weather-stream",
  "method": "POST",
  "timeout": "300s",
  "output_encoding": "no-op",
  "input_headers": ["Content-Type"],
  "backend": [{
    "host": ["https://weather.example.com"],
    "url_pattern": "/api/agents/weatherAgent/stream",
    "encoding": "no-op"
  }]
}
```

### AI / LLM streaming (Gemini example)

```json
{
  "endpoint": "/stream",
  "method": "POST",
  "output_encoding": "no-op",
  "input_headers": ["Content-Type"],
  "backend": [{
    "url_pattern": "/v1beta/models/gemini-2.0-flash:streamGenerateContent?alt=sse&key=xxx",
    "method": "POST",
    "encoding": "no-op",
    "host": ["https://generativelanguage.googleapis.com/"]
  }]
}
```

## SSE vs WebSockets

| | SSE | WebSockets |
|---|-----|------------|
| Direction | Server → client only | Bidirectional |
| Protocol | Plain HTTP | WS upgrade (RFC-6455) |
| Complexity | Lower | Higher |
| Use case | Notifications, feeds | Chat, gaming, IoT |

## Operational considerations

- Set timeout on the **streaming endpoint only** — avoid long timeouts at service level
- Streaming connections consume resources continuously
- Redeployments may wait for connection draining; use `max_shutdown_wait_time` to force-kill
- Infrastructure must support persistent HTTP connections (load balancers, proxies)

## Pucora CE status

**Implemented** in Pucora CE via `no-op` encoding with flush-aware streaming proxy. See [pucora-ce/docs/streaming-sse.md](../../pucora-ce/docs/streaming-sse.md) and `examples/streaming/` for configuration, Docker demo, and audit rules (5.2.4–5.2.6).
