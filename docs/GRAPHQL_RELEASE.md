# GraphQL release notes (CE integration)

## Lura module published

- Repository: https://github.com/pucora/lura
- Tag: `v2.0.4` (GraphQL namespace fix, method inheritance, GET query URL merge)
- Go module: `github.com/pucora/lura/v2`
- Config namespace: `backend/graphql` (alias for `transport/http/client/graphql`)

## CE integration

| Path | Purpose |
|------|---------|
| `cmd/pucora-ce/main.go` | `ExtraConfigAlias` for `backend/graphql` |
| `docs/graphql.md` | Configuration reference (3 KrakenD modes) |
| `examples/graphql/` | Docker Compose stack + smoke tests |
| `tests/fixtures/graphql_*.json` | Config fixtures |
| `tests/graphql_integration_test.go` | Adapter, proxy, federation, JWT, rate limit |

## Modes (KrakenD parity)

1. **REST → GraphQL adapter** — `backend/graphql` builds upstream query/mutation from REST
2. **GraphQL proxy** — forward client GraphQL unchanged (no `backend/graphql`)
3. **Simple federation** — parallel backends with `group` + `backend/graphql` each

## Commands

```bash
make test-graphql
make check-fixtures-graphql
make graphql-compose-test
./scripts/publish-fork-module.sh lura v2.0.4
```

## Dependency

CE `go.mod` requires `github.com/pucora/lura/v2 v2.0.4` or newer with GraphQL fixes.
