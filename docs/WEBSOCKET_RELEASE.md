# WebSocket release notes (CE integration)

## Module published

- Repository: https://github.com/velonetics/velonetics-websocket
- Tags: `v2.0.0`, `v2.0.1`
- Go module: `github.com/velonetics/velonetics-websocket/v2`

## CE integration

| Path | Purpose |
|------|---------|
| `forks/velonetics-websocket/` | Local fork (gitignored; publish with `scripts/publish-fork-module.sh`) |
| `handler_factory.go` | Wires WebSocket handler before JWT |
| `docs/websockets.md` | Configuration reference |
| `examples/websocket/` | Docker Compose stack + smoke tests |
| `tests/fixtures/ws_*.json` | Config fixtures |
| `velonetics-ws.json` | Minimal direct-mode sample config |

## Commands

```bash
make test-websocket      # unit tests in forks/velonetics-websocket
make check-fixtures      # validate ws_*.json + velonetics-ws.json
make ws-compose-test     # Docker Compose end-to-end smoke
./scripts/publish-fork-module.sh velonetics-websocket v2.0.x
```

## CI

- `.github/workflows/go.yml` — build, fixture validation, published-module build
- `.github/workflows/websocket-compose.yml` — Docker Compose smoke
