#!/usr/bin/env bash
# Clone published sibling modules next to pucora-ce for CI make targets that use ../ paths.
#
# Usage (from pucora-ce root, e.g. GitHub Actions after checkout):
#   ./scripts/clone-ci-siblings.sh
#   ./scripts/clone-ci-siblings.sh auth
#   ./scripts/clone-ci-siblings.sh endpoints
#
# Local monorepo checkouts already have siblings on disk; existing dirs are skipped.
set -euo pipefail

CE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PARENT="$(cd "${CE_ROOT}/.." && pwd)"
SCOPE="${1:-all}"

clone() {
  local repo="$1"
  local tag="$2"
  local dest="${PARENT}/${repo}"
  if [[ -d "${dest}/.git" || -f "${dest}/go.mod" ]]; then
    echo "==> ${repo} already present at ${dest}, skipping"
    return 0
  fi
  echo "==> Cloning ${repo}@${tag} into ${dest}"
  git clone --depth 1 --branch "${tag}" "https://github.com/pucora/${repo}.git" "${dest}"
}

clone_auth() {
  clone pucora-apikeys v2.0.0
  clone pucora-basicauth v2.0.0
  clone pucora-gcp-auth v2.0.0
  clone pucora-aws-sigv4 v2.0.0
  clone pucora-ntlm v2.0.0
  clone pucora-revoker v2.0.0
  clone pucora-jwk-aggregator v1.0.0
}

clone_endpoints() {
  clone pucora-jmespath v1.0.0
  clone pucora-response-body v1.0.0
  clone pucora-request-body v1.0.0
  clone pucora-jsonschema v2.0.2
  clone pucora-security-policies v1.0.0
  clone pucora-wildcard v1.0.0
  clone pucora-ratelimit v3.0.2
  clone pucora-openapi v1.0.0
  clone pucora-postman v1.0.0
  clone pucora-middleware-plugin v1.0.0
}

case "${SCOPE}" in
  auth) clone_auth ;;
  endpoints) clone_endpoints ;;
  all)
    clone_auth
    clone_endpoints
    ;;
  *)
    echo "usage: $0 [auth|endpoints|all]" >&2
    exit 2
    ;;
esac

echo "==> Sibling modules ready under ${PARENT}"
