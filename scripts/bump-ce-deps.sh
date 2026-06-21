#!/usr/bin/env bash
# Bump pucora-ce go.mod to latest rebrand publish tags.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
export GOWORK=off GOSUMDB=off GOPRIVATE=github.com/pucora/*

declare -a DEPS=(
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
)

for dep in "${DEPS[@]}"; do
	echo "go get ${dep}"
	go get "$dep" || echo "WARN: failed ${dep}"
done

go mod tidy
echo "==> Updated $(pwd)/go.mod and go.sum"
