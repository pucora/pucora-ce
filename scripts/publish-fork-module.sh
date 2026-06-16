#!/usr/bin/env bash
# Publish a forked Go module from forks/<name>/ to github.com/velonetics/<name>
#
# Usage:
#   ./scripts/publish-fork-module.sh velonetics-websocket v2.0.1
#   ./scripts/publish-fork-module.sh velonetics-websocket v2.0.1 --dry-run
#
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "usage: $0 <fork-dir-name> <tag> [--dry-run]" >&2
  echo "example: $0 velonetics-websocket v2.0.1" >&2
  exit 2
fi

FORK_NAME="$1"
TAG="$2"
DRY_RUN=false
if [[ "${3:-}" == "--dry-run" ]]; then
  DRY_RUN=true
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="${ROOT}/forks/${FORK_NAME}"
REMOTE="git@github.com:velonetics/${FORK_NAME}.git"

if [[ ! -d "$SRC" ]]; then
  echo "fork not found: $SRC" >&2
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
echo "==> Staging ${FORK_NAME}"
rsync -a --exclude '.git' "$SRC/" "$STAGE/"

if [[ -f "$ROOT/LICENSE" ]]; then
  cp "$ROOT/LICENSE" "$STAGE/LICENSE"
fi

cd "$STAGE"
echo "==> Preparing standalone go.mod"
awk '/^replace /,/^\)/{next} 1' go.mod > go.mod.standalone
mv go.mod.standalone go.mod
GOPROXY=direct go mod tidy

if $DRY_RUN; then
  echo "==> Dry run complete. Module staged at ${STAGE}"
  trap - EXIT
  exit 0
fi

if ! gh repo view "velonetics/${FORK_NAME}" >/dev/null 2>&1; then
  echo "==> Creating github.com/velonetics/${FORK_NAME}"
  gh repo create "velonetics/${FORK_NAME}" --public \
    --description "Velonetics CE module: ${FORK_NAME}"
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
  git commit -m "Release ${FORK_NAME} ${TAG}"
fi

echo "==> Pushing main"
git push origin main

if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "==> Tag ${TAG} already exists locally, updating"
  git tag -d "$TAG" >/dev/null 2>&1 || true
fi
git tag -a "$TAG" -m "${FORK_NAME} ${TAG}"
git push origin "$TAG" --force

echo "==> Published ${REMOTE} @ ${TAG}"
echo "Next: bump github.com/velonetics/${FORK_NAME}/v2 in velonetics-ce go.mod if needed."
