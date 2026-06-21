#!/usr/bin/env bash
# Verify Helm chart version metadata matches the expected gateway version.
set -euo pipefail

EXPECTED="${1:-}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CHART_DIR="${ROOT_DIR}/deploy/helm/pucora"
CHART_YAML="${CHART_DIR}/Chart.yaml"
VALUES_YAML="${CHART_DIR}/values.yaml"
MAKEFILE="${ROOT_DIR}/Makefile"

read_makefile_version() {
  awk -F' := ' '/^VERSION := / { print $2; exit }' "${MAKEFILE}"
}

read_chart_field() {
  local field="$1"
  awk -v field="${field}" '
    $1 == field ":" {
      value=$2
      gsub(/^"/, "", value)
      gsub(/"$/, "", value)
      print value
      exit
    }
  ' "${CHART_YAML}"
}

read_values_image_tag() {
  awk '
    /^image:/ { in_image=1; next }
    in_image && /^[[:space:]]+tag:/ {
      tag=$2
      gsub(/^"/, "", tag)
      gsub(/"$/, "", tag)
      print tag
      exit
    }
    /^[^[:space:]#]/ && $0 !~ /^image:/ { in_image=0 }
  ' "${VALUES_YAML}"
}

if [[ -z "${EXPECTED}" ]]; then
  EXPECTED="$(read_makefile_version)"
fi

EXPECTED="${EXPECTED#v}"

MAKEFILE_VERSION="$(read_makefile_version)"
CHART_VERSION="$(read_chart_field version)"
CHART_APP_VERSION="$(read_chart_field appVersion)"
VALUES_TAG="$(read_values_image_tag)"

fail=0
check() {
  local name="$1"
  local actual="$2"
  local want="$3"
  if [[ "${actual}" != "${want}" ]]; then
    echo "version mismatch: ${name}=${actual}, expected ${want}" >&2
    fail=1
  fi
}

check "Makefile VERSION" "${MAKEFILE_VERSION}" "${EXPECTED}"
check "Chart.yaml version" "${CHART_VERSION}" "${EXPECTED}"
check "Chart.yaml appVersion" "${CHART_APP_VERSION}" "${EXPECTED}"
check "values.yaml image.tag" "${VALUES_TAG}" "${EXPECTED}"

if [[ "${fail}" -ne 0 ]]; then
  echo "run: make sync-chart-version" >&2
  exit 1
fi

echo "Chart versions are aligned at ${EXPECTED}"
