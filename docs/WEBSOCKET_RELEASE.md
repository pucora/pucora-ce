# WebSocket release notes (CE integration)

## Module published

- Repository: https://github.com/pucora/pucora-websocket
- Tags: `v2.0.0`, `v2.0.1`
- Go module: `github.com/pucora/pucora-websocket/v2`
- Router support: https://github.com/pucora/lura `v2.0.1` (GET + NoopProxy for WS endpoints)

## CE integration

| Path | Purpose |
|------|---------|
| `../pucora-websocket` | Sibling module (use `go.work`; publish with `scripts/publish-fork-module.sh`) |
| `handler_factory.go` | Wires WebSocket handler before JWT |
| `docs/websockets.md` | Configuration reference |
| `examples/websocket/` | Docker Compose stack + smoke tests |
| `tests/fixtures/ws_*.json` | Config fixtures |
| `pucora-ws.json` | Minimal direct-mode sample config |

## Local workspace

Clone CE and sibling modules under one parent directory, then:

```bash
./scripts/init-workspace.sh   # writes ../go.work
make build
```

Without `go.work`, builds use published module versions from `go.mod`.

## Commands

```bash
make test-websocket      # unit tests in ../pucora-websocket
make check-fixtures      # validate ws_*.json + pucora-ws.json
make ws-compose-test     # Docker Compose end-to-end smoke
./scripts/publish-fork-module.sh pucora-websocket v2.0.x
./scripts/publish-fork-module.sh lura v2.0.x
```

## CI

- `.github/workflows/go.yml` — build, fixture validation, published-module build
- `.github/workflows/websocket-compose.yml` — Docker Compose smoke
