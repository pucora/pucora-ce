#!/usr/bin/env bash
# Publish all Pucora workspace modules with new tags (correct github.com/pucora/* paths).
# Run from pucora-ce: ./scripts/publish-all-rebrand-tags.sh [--dry-run]
set -euo pipefail

DRY_RUN=false
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=true

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKSPACE="$(cd "${ROOT}/.." && pwd)"
PUBLISH="${ROOT}/scripts/publish-fork-module.sh"
PUBLISHED_FILE="$(mktemp)"
trap 'rm -f "$PUBLISHED_FILE"' EXIT

export GOPRIVATE=github.com/pucora/*
export GOSUMDB=off

# module-dir:new-tag (dependency order)
ORDER="
binder:v1.0.1
flatmap:v1.0.1
lru:v1.0.1
httpcache:v1.0.1
go-auth0:v2.0.1
lura:v2.0.9
pucora-gologging:v2.0.1
pucora-gelf:v2.0.1
pucora-xml:v2.0.1
pucora-rss:v2.0.1
pucora-koanf:v1.0.1
pucora-flexibleconfig:v2.0.1
pucora-jsonschema:v2.0.1
pucora-httpcache:v2.0.1
pucora-httpsecure:v2.0.1
pucora-cors:v2.0.1
pucora-botdetector:v2.0.1
pucora-influx:v2.0.1
pucora-martian:v2.0.1
pucora-metrics:v2.0.1
pucora-circuitbreaker:v3.0.1
pucora-ratelimit:v3.0.1
pucora-cel:v2.0.1
pucora-lua:v2.0.1
pucora-opencensus:v2.0.1
pucora-otel:v1.0.1
pucora-oauth2-clientcredentials:v2.0.1
bloomfilter:v2.0.1
pucora-jose:v2.0.1
pucora-logstash:v2.0.1
pucora-lambda:v2.0.1
pucora-amqp:v2.0.4
pucora-audit:v1.0.2
pucora-cobra:v2.0.1
pucora-pubsub:v2.0.6
pucora-usage:v2.0.1
pucora-grpc:v2.0.8
pucora-soap:v2.2.3
pucora-websocket:v2.0.8
"

bump_local_deps() {
	local dir="$1"
	local gomod="${WORKSPACE}/${dir}/go.mod"
	local mod_path dep_tag
	[[ -f "$gomod" ]] || return 0
	while IFS=' ' read -r mod_path dep_tag; do
		[[ -n "$mod_path" ]] || continue
		if grep -q "${mod_path}" "$gomod" 2>/dev/null; then
			( cd "${WORKSPACE}/${dir}" && GOWORK=off go get "${mod_path}@${dep_tag}" 2>/dev/null ) || true
		fi
	done < "$PUBLISHED_FILE"
}

publish_one() {
	local dir="$1" tag="$2"
	local gomod="${WORKSPACE}/${dir}/go.mod"
	if [[ ! -f "$gomod" ]]; then
		echo "SKIP (no go.mod): ${dir}"
		return 0
	fi

	local module_path
	module_path="$(awk '/^module / {print $2; exit}' "$gomod")"

	echo ""
	echo "========================================"
	echo "Publishing ${dir} @ ${tag} (${module_path})"
	echo "========================================"

	bump_local_deps "$dir"

	if $DRY_RUN; then
		"$PUBLISH" "$dir" "$tag" --dry-run
	else
		"$PUBLISH" "$dir" "$tag"
	fi
	echo "${module_path} ${tag}" >> "$PUBLISHED_FILE"
}

echo "Workspace: ${WORKSPACE}"
echo "Dry run: ${DRY_RUN}"

failed=""
count=0
while IFS= read -r spec; do
	[[ -n "$spec" ]] || continue
	dir="${spec%%:*}"
	tag="${spec#*:}"
	count=$((count + 1))
	if ! publish_one "$dir" "$tag"; then
		failed="${failed} ${dir}:${tag}"
	fi
done <<EOF
${ORDER}
EOF

echo ""
echo "========================================"
echo "Publish summary (${count} modules)"
echo "========================================"
cat "$PUBLISHED_FILE" | while read -r line; do echo "  ${line}"; done
if [[ -n "${failed}" ]]; then
	echo "Failed:${failed}"
	exit 1
fi

if ! $DRY_RUN; then
	echo ""
	echo "Next: ./scripts/bump-ce-deps.sh && GOWORK=off go build ./cmd/pucora-ce"
fi
