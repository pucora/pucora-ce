#!/usr/bin/env bash
# Sync Helm chart version metadata with the gateway release version.
set -euo pipefail

VERSION="${1:-}"
CHART_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../pucora" && pwd)"
CHART_YAML="${CHART_DIR}/Chart.yaml"
VALUES_YAML="${CHART_DIR}/values.yaml"

if [[ -z "${VERSION}" ]]; then
  echo "usage: $0 <version>" >&2
  echo "example: $0 2.1.1" >&2
  exit 1
fi

VERSION="${VERSION#v}"

if [[ ! "${VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+([-.+][0-9A-Za-z.-]+)?$ ]]; then
  echo "invalid semver version: ${VERSION}" >&2
  exit 1
fi

if [[ ! -f "${CHART_YAML}" ]]; then
  echo "chart not found: ${CHART_YAML}" >&2
  exit 1
fi

tmp="$(mktemp)"
awk -v version="${VERSION}" '
  /^version:/ { print "version: " version; next }
  /^appVersion:/ { print "appVersion: \"" version "\""; next }
  { print }
' "${CHART_YAML}" > "${tmp}"
mv "${tmp}" "${CHART_YAML}"

tmp="$(mktemp)"
awk -v version="${VERSION}" '
  /^image:/ { in_image=1 }
  in_image && /^[[:space:]]+tag:/ {
    sub(/tag:.*/, "tag: \"" version "\"")
    in_image=0
  }
  /^[^[:space:]#]/ && $0 !~ /^image:/ { in_image=0 }
  { print }
' "${VALUES_YAML}" > "${tmp}"
mv "${tmp}" "${VALUES_YAML}"

echo "Synced chart version to ${VERSION}"
echo "  - ${CHART_YAML}"
echo "  - ${VALUES_YAML} (image.tag)"
