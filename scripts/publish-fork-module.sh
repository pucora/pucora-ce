#!/usr/bin/env bash
# Publish a Go module from a sibling repo to github.com/pucora/<name>
#
# Usage:
#   ./scripts/publish-fork-module.sh pucora-websocket v2.0.1
#   ./scripts/publish-fork-module.sh pucora-websocket v2.0.1 --dry-run
#   ./scripts/publish-fork-module.sh lura v2.0.1   # publishes ../pucora-lura
#
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "usage: $0 <module-dir-name> <tag> [--dry-run]" >&2
  echo "example: $0 pucora-websocket v2.0.1" >&2
  exit 2
fi

MODULE_NAME="$1"
TAG="$2"
DRY_RUN=false
if [[ "${3:-}" == "--dry-run" ]]; then
  DRY_RUN=true
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKSPACE="$(cd "${ROOT}/.." && pwd)"

# GitHub repo name (lura vs lura).
case "${MODULE_NAME}" in
  lura) GH_NAME="lura" ;;
  binder|bloomfilter|flatmap|go-auth0|httpcache|lru) GH_NAME="${MODULE_NAME}" ;;
  pucora-*) GH_NAME="${MODULE_NAME}" ;;
  *) GH_NAME="${MODULE_NAME}" ;;
esac

# Local directory (lura for lura).
case "${MODULE_NAME}" in
  lura) LOCAL_NAME="lura" ;;
  *) LOCAL_NAME="${MODULE_NAME}" ;;
esac

SRC="${WORKSPACE}/${LOCAL_NAME}"
REMOTE="git@github.com:pucora/${GH_NAME}.git"

if [[ ! -d "$SRC" ]]; then
  echo "module not found: $SRC" >&2
  echo "Clone sibling repos under ${WORKSPACE} or pass an existing module dir name." >&2
  exit 1
fi

if [[ ! -f "$SRC/go.mod" ]]; then
  echo "missing go.mod in $SRC" >&2
  exit 1
fi

BASE="$(mktemp -d)"
STAGE="${BASE}/stage"
REPO="${BASE}/repo"
cleanup() { rm -rf "$BASE"; }
trap cleanup EXIT

mkdir -p "$STAGE"
echo "==> Staging ${LOCAL_NAME} -> pucora/${GH_NAME}"
rsync -a --exclude '.git' "$SRC/" "$STAGE/"

if [[ -f "$ROOT/LICENSE" ]]; then
  cp "$ROOT/LICENSE" "$STAGE/LICENSE"
fi

cd "$STAGE"
echo "==> Preparing standalone go.mod"
awk '/^replace /,/^\)/{next} 1' go.mod > go.mod.standalone
mv go.mod.standalone go.mod
if ! GOPROXY=direct GOSUMDB=off GOPRIVATE=github.com/pucora/* go mod tidy; then
  echo "ERROR: go mod tidy failed for ${GH_NAME} — fix deps in local ${LOCAL_NAME}/go.mod first" >&2
  exit 1
fi

if $DRY_RUN; then
  echo "==> Dry run complete. Module staged at ${STAGE}"
  trap - EXIT
  exit 0
fi

if ! gh repo view "pucora/${GH_NAME}" >/dev/null 2>&1; then
  echo "==> Creating github.com/pucora/${GH_NAME}"
  gh repo create "pucora/${GH_NAME}" --public \
    --description "Pucora CE module: ${GH_NAME}"
fi

echo "==> Syncing with ${REMOTE}"
git clone "$REMOTE" "$REPO"
find "$REPO" -mindepth 1 -maxdepth 1 ! -name '.git' -exec rm -rf {} +
rsync -a "$STAGE"/ "$REPO"/

cd "$REPO"
git add -A
if git diff --cached --quiet; then
  echo "==> No file changes since last publish"
else
  git commit -m "Release ${GH_NAME} ${TAG}"
fi

echo "==> Pushing main"
git push origin main

if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "==> Tag ${TAG} already exists locally, updating"
  git tag -d "$TAG" >/dev/null 2>&1 || true
fi
git tag -a "$TAG" -m "${GH_NAME} ${TAG}"
git push origin "$TAG" --force

echo "==> Published ${REMOTE} @ ${TAG}"
echo "Next: bump github.com/pucora/${GH_NAME}/v2 in pucora-ce go.mod if needed."
