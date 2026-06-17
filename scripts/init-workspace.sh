#!/usr/bin/env bash
# Generate go.work at the Velonetics workspace root (parent of this repo).
#
# Usage (from velonetics-ce-master or workspace root):
#   ./scripts/init-workspace.sh
#
# Creates ../../go.work relative to velonetics-ce-master, listing every sibling
# Go module found on disk. Local builds then use sibling repos instead of
# published GitHub tags. CI and solo CE clones continue to use go.mod require.
#
set -euo pipefail

CE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKSPACE_ROOT="$(cd "${CE_ROOT}/.." && pwd)"
WORK_FILE="${WORKSPACE_ROOT}/go.work"

GO_VERSION="$(grep '^go ' "${CE_ROOT}/go.mod" | awk '{print $2}')"

# Sibling module directories (relative to workspace root).
MODULE_DIRS=(
	velonetics-ce-master
	binder
	bloomfilter
	flatmap
	go-auth0
	httpcache
	lru
	velonetics-amqp
	velonetics-audit
	velonetics-botdetector
	velonetics-cel
	velonetics-circuitbreaker
	velonetics-cobra
	velonetics-cors
	velonetics-flexibleconfig
	velonetics-gelf
	velonetics-gologging
	velonetics-httpcache
	velonetics-httpsecure
	velonetics-influx
	velonetics-jose
	velonetics-jsonschema
	velonetics-koanf
	velonetics-lambda
	velonetics-logstash
	velonetics-lua
	velonetics-lura
	velonetics-martian
	velonetics-metrics
	velonetics-oauth2-clientcredentials
	velonetics-opencensus
	velonetics-otel
	velonetics-pubsub
	velonetics-ratelimit
	velonetics-rss
	velonetics-soap
	velonetics-grpc
	velonetics-usage
	velonetics-websocket
	velonetics-xml
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
