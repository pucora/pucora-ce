# gRPC release notes (CE integration)

## Module published

- Repository: https://github.com/velonetics/velonetics-grpc
- Tag: `v2.0.0`
- Go module: `github.com/velonetics/velonetics-grpc/v2`

## CE integration

| Path | Purpose |
|------|---------|
| `../velonetics-grpc` | Sibling module (use `go.work` locally; publish with `scripts/publish-fork-module.sh`) |
| `backend_factory.go` | Wires `backend/grpc` client factory |
| `executor.go` | Catalog bootstrap + gRPC server on gateway port |
| `docs/grpc.md` | Configuration reference |
| `examples/grpc/` | Docker Compose stack + smoke tests |
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
make test-grpc           # unit tests in ../velonetics-grpc (local monorepo)
make check-grpc-fixtures # validate grpc_*.json
make grpc-compose-test   # Docker Compose end-to-end smoke
./scripts/publish-fork-module.sh velonetics-grpc v2.0.x
```

## KrakenD parity (v1)

- Unary RPC only (no streaming)
- `.pb` catalog (not raw `.proto` at runtime)
- gRPC client: REST → upstream gRPC → JSON
- gRPC server: same port as HTTP via `cmux`, reflection enabled
