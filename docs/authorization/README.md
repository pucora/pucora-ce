# Pucora Authorization (Enterprise Parity)

Implementation reference for KrakenD Enterprise authentication and authorization features in Pucora CE.

> **Source docs:** [docs/authorization](../../docs/authorization/)

## Feature index

| # | Feature | Edition | Namespace | Module | Doc |
|---|---------|---------|-----------|--------|-----|
| 1 | API Keys | Enterprise | `auth/api-keys` | `pucora-apikeys` | [01-api-keys.md](01-api-keys.md) |
| 2 | Basic Auth | Enterprise | `auth/basic` | `pucora-basicauth` | [02-basic-auth.md](02-basic-auth.md) |
| 3 | Revoke Server | Enterprise | `auth/revoker` + `pucora revoker` | `pucora-revoker` | [03-revoke-server.md](03-revoke-server.md) |
| 4 | Multiple IdP | Enterprise | `plugin/http-server` → `jwk-aggregator` | `pucora-jwk-aggregator` | [04-multiple-idp.md](04-multiple-idp.md) |
| 5 | GCP Auth | Enterprise | `auth/gcp` | `pucora-gcp-auth` | [05-gcp-auth.md](05-gcp-auth.md) |
| 6 | AWS SigV4 | Enterprise | `auth/aws-sigv4` | `pucora-aws-sigv4` | [06-aws-sigv4.md](06-aws-sigv4.md) |
| 7 | NTLM | Enterprise | `auth/ntlm` | `pucora-ntlm` | [07-ntlm.md](07-ntlm.md) |

All seven Enterprise-only authorization features are implemented in Pucora CE with schema parity.

**Production go-live:** [GO-LIVE.md](GO-LIVE.md) — checklist, auth matrix template, and `pucora audit` gates.

## Test fixtures

See [pucora-ce/tests/fixtures/](../tests/fixtures/):

- `api_keys.json` — service keys + protected endpoint
- `basic_auth.json` — inline bcrypt user + protected endpoint
- `backend_gcp.json` — GCP ID token injection
- `backend_sigv4.json` — AWS SigV4 signing
- `backend_ntlm.json` — NTLM backend client
- `backend_client_tls.json` — per-backend mTLS via `backend/http/client`
- `revoke_server.json` — standalone revoker config
- `multi_idp.json` — jwk-aggregator plugin + JWT validator

Validate fixtures:

```bash
make -C pucora-ce check-fixtures-auth
```

Run module unit tests:

```bash
make -C pucora-ce test-auth
```

## Build jwk-aggregator plugin

```bash
make -C pucora-ce jwk-aggregator-plugin
```

Produces `pucora-ce/plugins/jwk-aggregator.so`.

## Run revoke server

```bash
pucora revoker -c tests/fixtures/revoke_server.json
```
