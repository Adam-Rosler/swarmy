#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="${SWARMY_REPO_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
BIN_PATH="$REPO_ROOT/bin/swarmy"

if ! command -v go >/dev/null 2>&1; then
  echo "error: go is required but not found on PATH" >&2
  exit 1
fi

needs_rebuild() {
  if [ ! -x "$BIN_PATH" ]; then
    return 0
  fi

  if find "$REPO_ROOT" -type f \( -name '*.go' -o -name 'go.mod' -o -name 'go.sum' \) -newer "$BIN_PATH" -print -quit | grep -q .; then
    return 0
  fi

  return 1
}

if needs_rebuild; then
  mkdir -p "$REPO_ROOT/bin"
  (cd "$REPO_ROOT" && go build -o "$BIN_PATH" ./cmd/swarmy)
fi

exec "$BIN_PATH" "$@"
