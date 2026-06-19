#!/usr/bin/env bash
# Install the Pucora Helm chart on a Kind cluster and verify health.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CHART_DIR="${ROOT_DIR}/deploy/helm/pucora"
RELEASE_NAME="${RELEASE_NAME:-pucora-ci}"
NAMESPACE="${NAMESPACE:-default}"
CLUSTER_NAME="${CLUSTER_NAME:-pucora-helm-test}"
IMAGE_REPO="${IMAGE_REPO:-pucora-ci}"
IMAGE_TAG="${IMAGE_TAG:-test}"
SKIP_BUILD="${SKIP_BUILD:-0}"
SKIP_CLUSTER_CREATE="${SKIP_CLUSTER_CREATE:-0}"

log() {
  echo "[helm-cluster-test] $*"
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

require_cmd kubectl
require_cmd helm
require_cmd docker
require_cmd kind

if [[ "${SKIP_CLUSTER_CREATE}" != "1" ]]; then
  if ! kind get clusters 2>/dev/null | grep -qx "${CLUSTER_NAME}"; then
    log "creating kind cluster ${CLUSTER_NAME}"
    kind create cluster --name "${CLUSTER_NAME}" --wait 120s
  else
    log "reusing kind cluster ${CLUSTER_NAME}"
    kind export kubeconfig --name "${CLUSTER_NAME}"
  fi
fi

if [[ "${SKIP_BUILD}" != "1" ]]; then
  log "building gateway image ${IMAGE_REPO}:${IMAGE_TAG}"
  docker build \
    --build-arg GOLANG_VERSION="$(awk -F' := ' '/^GOLANG_VERSION := / { print $2; exit }' "${ROOT_DIR}/Makefile")" \
    --build-arg ALPINE_VERSION="$(awk -F' := ' '/^ALPINE_VERSION := / { print $2; exit }' "${ROOT_DIR}/Makefile")" \
    --build-arg VERSION="$(awk -F' := ' '/^VERSION := / { print $2; exit }' "${ROOT_DIR}/Makefile")" \
    -t "${IMAGE_REPO}:${IMAGE_TAG}" \
    "${ROOT_DIR}"
  kind load docker-image "${IMAGE_REPO}:${IMAGE_TAG}" --name "${CLUSTER_NAME}"
fi

log "installing chart release ${RELEASE_NAME}"
helm upgrade --install "${RELEASE_NAME}" "${CHART_DIR}" \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --set image.repository="${IMAGE_REPO}" \
  --set image.tag="${IMAGE_TAG}" \
  --set image.pullPolicy=Never \
  --set replicaCount=1 \
  --set tests.enabled=true \
  --set tests.configCheck=true \
  --wait \
  --timeout 5m

SERVICE="${RELEASE_NAME}-pucora"
log "waiting for deployment rollout"
kubectl rollout status "deployment/${SERVICE}" -n "${NAMESPACE}" --timeout=5m

log "checking /__health via in-cluster curl"
kubectl run "pucora-health-check-$RANDOM" \
  --namespace "${NAMESPACE}" \
  --rm \
  --restart=Never \
  --image=curlimages/curl:8.5.0 \
  --command -- \
  curl -sf "http://${SERVICE}.${NAMESPACE}.svc.cluster.local:8080/__health"

log "running helm test"
helm test "${RELEASE_NAME}" --namespace "${NAMESPACE}" --timeout 5m

log "cluster install test passed"
