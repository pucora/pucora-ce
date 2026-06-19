#!/usr/bin/env bash
# Generate go.work at the Pucora workspace root (parent of this repo).
#
# Usage (from pucora-ce or workspace root):
#   ./scripts/init-workspace.sh
#
# Creates ../../go.work relative to pucora-ce, listing every sibling
# Go module found on disk. Local builds then use sibling repos instead of
# published GitHub tags. CI and solo CE clones use go.mod require only.
#
# Do NOT add replace ../... blocks to committed go.mod files — they break
# standalone clones. go.work is local-only and must not be committed to
# individual module repos.
#
set -euo pipefail

CE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKSPACE_ROOT="$(cd "${CE_ROOT}/.." && pwd)"
WORK_FILE="${WORKSPACE_ROOT}/go.work"

GO_VERSION="$(grep '^go ' "${CE_ROOT}/go.mod" | awk '{print $2}')"

# Sibling module directories (relative to workspace root).
MODULE_DIRS=(
	pucora-ce
	binder
	bloomfilter
	flatmap
	go-auth0
	httpcache
	lru
	pucora-amqp
	pucora-audit
	pucora-botdetector
	pucora-cel
	pucora-circuitbreaker
	pucora-cobra
	pucora-cors
	pucora-flexibleconfig
	pucora-gelf
	pucora-gologging
	pucora-httpcache
	pucora-httpsecure
	pucora-influx
	pucora-jose
	pucora-jsonschema
	pucora-koanf
	pucora-lambda
	pucora-logstash
	pucora-lua
	lura
	pucora-martian
	pucora-metrics
	pucora-oauth2-clientcredentials
	pucora-opencensus
	pucora-otel
	pucora-pubsub
	pucora-ratelimit
	pucora-rss
	pucora-soap
	pucora-grpc
	pucora-usage
	pucora-websocket
	pucora-xml
)

{
	echo "go ${GO_VERSION}"
	echo ""
	echo "use ("
	for dir in "${MODULE_DIRS[@]}"; do
		if [[ -f "${WORKSPACE_ROOT}/${dir}/go.mod" ]]; then
			echo "	./${dir}"
		fi
	done
	echo ")"
} > "${WORK_FILE}"

echo "==> Wrote ${WORK_FILE}"
echo "    Run builds from ${CE_ROOT} or any module in the workspace."
echo "    Delete go.work to use published modules from GitHub."
