# Production authorization go-live checklist

Use this checklist before exposing a Pucora gateway to production traffic. It complements `pucora check` (syntax) and `pucora audit` (security recommendations).

## Automated gates (run in CI)

```bash
# Syntax + schema
pucora check -c your-prod-config.json

# Security audit — fail on HIGH and CRITICAL
pucora audit -c your-prod-config.json -s CRITICAL,HIGH

# Reference baseline (must stay clean)
pucora audit -c tests/fixtures/prod_auth_baseline.json -s CRITICAL,HIGH

# Auth integration smoke (after build)
go test ./tests -run 'TestAPIKeys|TestBasicAuth|TestOAuth2|TestAWSSigV4' -count=1
```

## Auth matrix (document per environment)

For each endpoint, record:

| Endpoint | Client auth | Roles/scopes | Backend auth | Public? |
|----------|-------------|--------------|--------------|---------|
| `/public/health` | None | — | None | Yes |
| `/api/{resource}` | JWT (IdP JWKS) | `user` | OAuth2 client credentials | No |

Store this matrix in your runbook or internal wiki and keep it aligned with config changes.

## Manual checks enterprises expect

### Secrets and config

- [ ] No API keys, passwords, OAuth `client_secret`, or GCP SA JSON committed to git
- [ ] Inject secrets via Vault, K8s Secrets, or your platform's secret store
- [ ] API keys use `hash: "sha256"` (or stronger), not `plain`
- [ ] Revoke server has `revoke_server_api_key` set and is reachable only on a private network

### JWT / IdP

- [ ] `disable_jwk_security` is **not** set in production
- [ ] `operation_debug` is **not** enabled on validators
- [ ] `audience`, `issuer`, and `roles`/`scopes` match your IdP token layout
- [ ] JWK cache TTL tested against your IdP key rotation policy
- [ ] Clock sync (NTP) on all gateway nodes

### Transport

- [ ] TLS terminates at gateway or a trusted LB — document which
- [ ] `allow_insecure_connections` is false
- [ ] Backend mTLS cert paths exist in the runtime image and rotation is owned by ops
- [ ] Service mTLS (`enable_mtls`) only if all clients can present certs

### Abuse protection

- [ ] Rate limits on authenticated and public routes (`qos/ratelimit/*`)
- [ ] Bot detector or WAF in front for internet-facing APIs
- [ ] Circuit breakers on upstream backends

### Revocation

- [ ] Bloom filter / revoke server HA documented
- [ ] Revoke drill: revoke token → 401 on all nodes → survives gateway restart
- [ ] False-positive rate for bloom filter (`n`, `p`) accepted by security team

### Observability

- [ ] Auth failures logged without raw tokens or API keys
- [ ] Metrics/alerts on 401/403 spikes and upstream auth errors (OAuth2, GCP, SigV4)

### Staging smoke (real dependencies)

- [ ] Valid IdP JWT accepted; expired/wrong-audience rejected
- [ ] OAuth2 client credentials reach backend with expected Bearer token
- [ ] AWS SigV4 or GCP ID token works against a staging upstream with real IAM
- [ ] NTLM only if explicitly approved — not default for new services

## Audit rules for auth (section 1.3)

| Rule | Severity | Trigger |
|------|----------|---------|
| 1.3.1 | HIGH | `disable_jwk_security: true` on any JWT validator |
| 1.3.2 | HIGH | `operation_debug: true` on any JWT validator |
| 1.3.3 | CRITICAL | `auth/revoker` without `revoke_server_api_key` |
| 1.3.4 | MEDIUM | API keys with `hash: plain` or unset |
| 1.3.5 | HIGH | API keys with `strategy: query_string` |

Exclude rules only with documented exceptions:

```bash
pucora audit -c prod.json -s CRITICAL,HIGH -i 1.1.1,1.1.2
```

## Related docs

- [Authorization features](../authorization/README.md)
- [Pucora audit tool](https://pucora.io/docs/configuration/audit/)
