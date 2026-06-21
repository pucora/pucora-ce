#!/usr/bin/env bash
# Update every workspace module to use rebrand publish tags (run before publish-all).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKSPACE="$(cd "${ROOT}/.." && pwd)"

export GOWORK=off GOSUMDB=off GOPRIVATE=github.com/pucora/* GOPROXY=direct

# All target tags (dependency-safe versions with github.com/pucora/* module paths)
TAGS="
github.com/pucora/binder@v1.0.1
github.com/pucora/bloomfilter/v2@v2.0.1
github.com/pucora/flatmap@v1.0.1
github.com/pucora/go-auth0/v2@v2.0.1
github.com/pucora/httpcache@v1.0.1
github.com/pucora/lru@v1.0.1
github.com/pucora/lura/v2@v2.0.9
github.com/pucora/pucora-amqp/v2@v2.0.4
github.com/pucora/pucora-audit@v1.0.2
github.com/pucora/pucora-botdetector/v2@v2.0.1
github.com/pucora/pucora-cel/v2@v2.0.1
github.com/pucora/pucora-circuitbreaker/v3@v3.0.1
github.com/pucora/pucora-cobra/v2@v2.0.1
github.com/pucora/pucora-cors/v2@v2.0.1
github.com/pucora/pucora-flexibleconfig/v2@v2.0.1
github.com/pucora/pucora-gelf/v2@v2.0.1
github.com/pucora/pucora-gologging/v2@v2.0.1
github.com/pucora/pucora-grpc/v2@v2.0.8
github.com/pucora/pucora-httpcache/v2@v2.0.1
github.com/pucora/pucora-httpsecure/v2@v2.0.1
github.com/pucora/pucora-influx/v2@v2.0.1
github.com/pucora/pucora-jose/v2@v2.0.1
github.com/pucora/pucora-jsonschema/v2@v2.0.1
github.com/pucora/pucora-koanf@v1.0.1
github.com/pucora/pucora-lambda/v2@v2.0.1
github.com/pucora/pucora-logstash/v2@v2.0.1
github.com/pucora/pucora-lua/v2@v2.0.1
github.com/pucora/pucora-martian/v2@v2.0.1
github.com/pucora/pucora-metrics/v2@v2.0.1
github.com/pucora/pucora-oauth2-clientcredentials/v2@v2.0.1
github.com/pucora/pucora-opencensus/v2@v2.0.1
github.com/pucora/pucora-otel@v1.0.1
github.com/pucora/pucora-pubsub/v2@v2.0.6
github.com/pucora/pucora-ratelimit/v3@v3.0.1
github.com/pucora/pucora-rss/v2@v2.0.1
github.com/pucora/pucora-soap/v2@v2.2.3
github.com/pucora/pucora-usage/v2@v2.0.1
github.com/pucora/pucora-websocket/v2@v2.0.8
github.com/pucora/pucora-xml/v2@v2.0.1
"

MODULE_DIRS="
binder bloomfilter flatmap go-auth0 httpcache lru lura
pucora-amqp pucora-audit pucora-botdetector pucora-cel pucora-circuitbreaker
pucora-cobra pucora-configurator pucora-cors pucora-flexibleconfig pucora-gelf
pucora-gologging pucora-grpc pucora-httpcache pucora-httpsecure pucora-influx
pucora-jose pucora-jsonschema pucora-koanf pucora-lambda pucora-logstash
pucora-lua pucora-martian pucora-metrics pucora-oauth2-clientcredentials
pucora-opencensus pucora-otel pucora-pubsub pucora-ratelimit pucora-rss
pucora-soap pucora-usage pucora-websocket pucora-xml pucora-ce
"

echo "Syncing pucora deps across workspace: ${WORKSPACE}"

for dir in $MODULE_DIRS; do
  gomod="${WORKSPACE}/${dir}/go.mod"
  [[ -f "$gomod" ]] || continue
  echo ""
  echo "==> ${dir}"
  (
    cd "${WORKSPACE}/${dir}"
    for tag in $TAGS; do
      [[ -n "$tag" ]] || continue
      mod="${tag%@*}"
      if grep -q "$mod" go.mod 2>/dev/null; then
        go get "$tag" 2>&1 | tail -1 || true
      fi
    done
    go mod tidy 2>&1 | tail -1 || echo "WARN: tidy failed for ${dir}"
  )
done

echo ""
echo "==> Workspace dep sync complete"
