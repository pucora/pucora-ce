# gRPC Integration

Velonetics exposes gRPC client and server integration with KrakenD-parity configuration. Unary RPC only; catalog uses compiled `.pb` descriptor files.

Implemented by [`velonetics-grpc`](https://github.com/velonetics/velonetics-grpc) via:

- `extra_config.grpc` — service catalog and optional gRPC server
- `extra_config.backend/grpc` — gRPC upstream backends

## Catalog setup

Generate `.pb` files from `.proto`:

```bash
protoc --descriptor_set_out=flights.pb flights.proto
```

Service-level config:

```json
{
  "extra_config": {
    "grpc": {
      "catalog": ["./grpc/flights.pb", "./grpc/definitions"]
    }
  }
}
```

## gRPC client (REST → gRPC)

```json
{
  "extra_config": {
    "grpc": { "catalog": ["grpcatalog/flights.pb"] }
  },
  "endpoints": [{
    "endpoint": "/flights",
    "input_query_strings": ["*"],
    "backend": [{
      "host": ["localhost:4242"],
      "url_pattern": "/flight_finder.Flights/FindFlight",
      "extra_config": { "backend/grpc": {} }
    }]
  }]
}
```

| Rule | Value |
|------|-------|
| `host` | `host:port` only (no `http://`) |
| `url_pattern` | `/package.Service/Method` |
| Streaming | Not supported in v1 |

### Key `backend/grpc` fields

| Field | Description |
|-------|-------------|
| `use_request_body` | Fill request from HTTP body |
| `input_mapping` | Map query/placeholder params to nested fields |
| `header_mapping` | Map HTTP headers to gRPC metadata |
| `request_naming_convention` | `snake_case` (default) or `camelCase` |
| `response_naming_convention` | `snake_case` (default) or `camelCase` |
| `client_tls` | TLS client settings |
| `use_alternate_host_on_error` | Retry on alternate hosts |

## gRPC server (same port as HTTP)

```json
{
  "extra_config": {
    "grpc": {
      "catalog": ["./grpc/definitions"],
      "server": {
        "services": [{
          "name": "flight_finder.Flights",
          "methods": [{
            "name": "FindFlight",
            "input_headers": ["*"],
            "payload_params": { "page.cursor": "cursor" },
            "backend": [...]
          }]
        }]
      }
    }
  }
}
```

gRPC reflection is enabled automatically. Use `grpcurl -plaintext localhost:8080 list` to discover services.

## Local development

```bash
cd velonetics-ce-master
make test-grpc
make check-grpc-fixtures
make grpc-compose-test
```

## Limitations

- Unary RPC only (no streaming)
- Catalog requires `.pb` files (not raw `.proto` at runtime)
- Arrays of objects cannot be filled via query strings
