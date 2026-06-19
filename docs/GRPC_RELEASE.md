# gRPC release notes (CE integration)

## Module published

- Repository: https://github.com/pucora/pucora-grpc
- Tag: `v2.0.1` (parity gap closure)
- Go module: `github.com/pucora/pucora-grpc/v2`

### v2.0.1 highlights

- Server `fillResponse` handles `resp.Data` and `resp.Io` (JSON + raw protobuf)
- gRPC-to-gRPC passthrough for single `backend/grpc` method backends
- Per-method JWT via `auth/validator` on `grpc.server` methods
- Server OTel overrides (`disable_metrics`, `disable_traces`)
- TLS + h2c on same port via cmux (reuses gateway `tls` config)
- Client: `read_buffer_size`, `request_naming_convention`, connection-state host retry
- Startup guard: `output_encoding: no-op` rejected for gRPC client endpoints

## CE release

- Version: `v2.0.2`
- Requires: `github.com/pucora/pucora-grpc/v2 v2.0.1`

## CE integration

| Path | Purpose |
|------|---------|
| `../pucora-grpc` | Sibling module (use `go.work` locally; publish with `scripts/publish-fork-module.sh`) |
| `backend_factory.go` | Wires `backend/grpc` client factory |
| `executor.go` | Catalog bootstrap + gRPC server on gateway port + JWT rejecter |
| `docs/grpc.md` | Configuration reference |
| `examples/grpc/` | Docker Compose stacks (client, server, mixed, JWT) |
| `tests/fixtures/grpc_*.json` | Config fixtures |

## Local workspace

Clone CE and sibling modules under one parent directory, then:

```bash
./scripts/init-workspace.sh   # writes ../go.work
make build
```

Without `go.work`, builds use published module versions from `go.mod` (no `replace` for gRPC).

## Commands

```bash
make test-grpc           # unit tests in ../pucora-grpc (local monorepo)
make check-grpc-fixtures # validate grpc_*.json
make grpc-compose-test   # Docker Compose end-to-end (client + server + mixed + JWT)
./scripts/publish-fork-module.sh pucora-grpc v2.0.x
```

## CI

- `go.yml`: module tests + fixture validation
- `grpc-compose.yml`: full compose smoke on path changes

## GOSUMDB

If `sum.golang.org` has not indexed a newly published tag yet, CI may need `GOSUMDB=off` temporarily. Remove once the module is indexed.

## KrakenD parity (v2)

- Unary RPC only (no streaming)
- `.pb` catalog (not raw `.proto` at runtime)
- gRPC client: REST → upstream gRPC → JSON
- gRPC server: same port as HTTP via `cmux`, reflection, JWT, passthrough
- TLS ingress shares gateway `tls` configuration
