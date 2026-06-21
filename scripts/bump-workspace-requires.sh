#!/usr/bin/env bash
# Bump stale velonetics-era require versions in all workspace go.mod files.
set -euo pipefail

WORKSPACE="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

for gomod in "$WORKSPACE"/*/go.mod; do
  [[ -f "$gomod" ]] || continue
  sed -i '' \
    -e 's|github.com/pucora/lura/v2 v2.0.[0-8]|github.com/pucora/lura/v2 v2.0.9|g' \
    -e 's|github.com/pucora/flatmap v1.0.0|github.com/pucora/flatmap v1.0.1|g' \
    -e 's|github.com/pucora/binder v1.0.0|github.com/pucora/binder v1.0.1|g' \
    -e 's|github.com/pucora/lru v1.0.0|github.com/pucora/lru v1.0.1|g' \
    -e 's|github.com/pucora/httpcache v1.0.0|github.com/pucora/httpcache v1.0.1|g' \
    -e 's|github.com/pucora/go-auth0/v2 v2.0.0|github.com/pucora/go-auth0/v2 v2.0.1|g' \
    -e 's|github.com/pucora/bloomfilter/v2 v2.0.0|github.com/pucora/bloomfilter/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-gologging/v2 v2.0.0|github.com/pucora/pucora-gologging/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-jose/v2 v2.0.0|github.com/pucora/pucora-jose/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-koanf v1.0.0|github.com/pucora/pucora-koanf v1.0.1|g' \
    -e 's|github.com/pucora/pucora-gelf/v2 v2.0.0|github.com/pucora/pucora-gelf/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-xml/v2 v2.0.0|github.com/pucora/pucora-xml/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-rss/v2 v2.0.0|github.com/pucora/pucora-rss/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-flexibleconfig/v2 v2.0.0|github.com/pucora/pucora-flexibleconfig/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-jsonschema/v2 v2.0.0|github.com/pucora/pucora-jsonschema/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-httpcache/v2 v2.0.0|github.com/pucora/pucora-httpcache/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-httpsecure/v2 v2.0.0|github.com/pucora/pucora-httpsecure/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-cors/v2 v2.0.0|github.com/pucora/pucora-cors/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-botdetector/v2 v2.0.0|github.com/pucora/pucora-botdetector/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-martian/v2 v2.0.0|github.com/pucora/pucora-martian/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-metrics/v2 v2.0.0|github.com/pucora/pucora-metrics/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-influx/v2 v2.0.0|github.com/pucora/pucora-influx/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-circuitbreaker/v3 v3.0.0|github.com/pucora/pucora-circuitbreaker/v3 v3.0.1|g' \
    -e 's|github.com/pucora/pucora-ratelimit/v3 v3.0.0|github.com/pucora/pucora-ratelimit/v3 v3.0.1|g' \
    -e 's|github.com/pucora/pucora-cel/v2 v2.0.0|github.com/pucora/pucora-cel/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-lua/v2 v2.0.0|github.com/pucora/pucora-lua/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-opencensus/v2 v2.0.0|github.com/pucora/pucora-opencensus/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-otel v1.0.0|github.com/pucora/pucora-otel v1.0.1|g' \
    -e 's|github.com/pucora/pucora-oauth2-clientcredentials/v2 v2.0.0|github.com/pucora/pucora-oauth2-clientcredentials/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-logstash/v2 v2.0.0|github.com/pucora/pucora-logstash/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-lambda/v2 v2.0.0|github.com/pucora/pucora-lambda/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-amqp/v2 v2.0.[0-3]|github.com/pucora/pucora-amqp/v2 v2.0.4|g' \
    -e 's|github.com/pucora/pucora-audit v1.0.0|github.com/pucora/pucora-audit v1.0.2|g' \
    -e 's|github.com/pucora/pucora-audit v1.0.1|github.com/pucora/pucora-audit v1.0.2|g' \
    -e 's|github.com/pucora/pucora-cobra/v2 v2.0.0|github.com/pucora/pucora-cobra/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-pubsub/v2 v2.0.[0-5]|github.com/pucora/pucora-pubsub/v2 v2.0.6|g' \
    -e 's|github.com/pucora/pucora-usage/v2 v2.0.0|github.com/pucora/pucora-usage/v2 v2.0.1|g' \
    -e 's|github.com/pucora/pucora-grpc/v2 v2.0.[0-7]|github.com/pucora/pucora-grpc/v2 v2.0.8|g' \
    -e 's|github.com/pucora/pucora-soap/v2 v2.2.[0-2]|github.com/pucora/pucora-soap/v2 v2.2.3|g' \
    -e 's|github.com/pucora/pucora-websocket/v2 v2.0.[0-7]|github.com/pucora/pucora-websocket/v2 v2.0.8|g' \
    "$gomod"
done

echo "==> Bumped require versions in workspace go.mod files"
