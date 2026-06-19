# Non-REST Connectivity — Bugfix Changelog

Pucora CE parity fixes across websocket, gRPC, pubsub/kafka, SOAP, AMQP, and lura (GraphQL/streaming). Module versions vendored in `pucora-ce` as of this changelog.

## Published module versions (CE `go.mod`)

| Module | Version | Areas |
|--------|---------|--------|
| `github.com/pucora/lura/v2` | **v2.0.7** | GraphQL GET dedup, streaming timeout (gin + mux), nil headers |
| `github.com/pucora/pucora-amqp/v2` | **v2.0.3** | Async probe, auto_ack rejection, consumer race, QueueBind errors, connection leak |
| `github.com/pucora/pucora-grpc/v2` | **v2.0.7** | fillResponse Io, multiplex shutdown, cookie JWT from metadata |
| `github.com/pucora/pucora-pubsub/v2` | **v2.0.5** | Kafka async commit/retry, startup probe, HTTP subscriber pending offsets, format-before-commit |
| `github.com/pucora/pucora-soap/v2` | **v2.2.2** | Validate without watcher leak, `key_password` for encrypted X509 keys |
| `github.com/pucora/pucora-websocket/v2` | **v2.0.7** | Hub lifecycle, flush/requeue, outbox warnings, disconnect event, binary frames, test registry |

## Fixes by area

### WebSocket (`pucora-websocket`)

- Hub backend uses gateway lifecycle context (not per-client request ctx) for read/reconnect
- `flushAllPending` requeues all unsent messages on backend failure
- `deliverToClient` logs when client outbox is full (no silent drop)
- `toString(float64)` session matching for JSON numeric params
- `disconnect_event` sent on lifecycle context
- Binary client frames preserved in hub `writePump`
- `ResetHubRegistry` cancels hub lifecycle in tests

### gRPC (`pucora-grpc`)

- `fillResponse` closes `resp.Io` on both Data and Io paths
- Multiplex server shuts down all listeners on `Serve` error
- gRPC server JWT: cookies from metadata (`cookie`, `grpcgateway-cookie`, `cookie_key`)
- Client `ErrorProxy` for misconfiguration (prior pass)

### Pub/Sub / Kafka (`pucora-pubsub`)

- Kafka async: commit only on pipeline success; retry pending offset (no implicit commit via later offset)
- Kafka async: sequential processing; startup probe is config-only
- Kafka HTTP subscriber: format before commit; pending message on failure; propagate commit errors
- Go Cloud pubsub subscriber: format before ack
- Subscriber decode errors skip commit (prior pass)

### SOAP (`pucora-soap`)

- `ValidateConfig` does not start template watcher goroutines
- WS-Security X509: `key_password` decrypts encrypted PEM private keys
- Startup validation + `ErrorProxy` (prior pass)

### AMQP (`pucora-amqp`)

- Async agent: config-only startup probe (no message consumption)
- Async agent: reject `auto_ack` (prevents silent loss on pipeline failure)
- Sync consumer: thread-safe delivery channel on reconnect
- Sync consumer: fail on `QueueBind` errors
- Close AMQP connection when `Channel()` fails after `Dial`

### Lura core (`lura`)

- GraphQL GET: `Set` query params (no duplicate `query=` keys)
- GraphQL: nil `Headers` guard on GET/POST
- GraphQL: empty variable placeholder no longer panics at startup
- Gin + mux: streaming endpoints skip endpoint `timeout` on request context

### CE integration (`pucora-ce`)

- Async agents and Gin router share `errgroup` context (router stops when agents fail)
- Kafka + SOAP startup validation in `executor.go`

## Upgrade CE

```bash
cd pucora-ce
GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod tidy
GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod vendor
GOWORK=off go build ./cmd/pucora-ce/
```

Publish sibling modules with `scripts/publish-fork-module.sh` when cutting releases from the monorepo.
