#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUNNER="$REPO_ROOT/scripts/swarmy-auto.sh"
RC_PATH="${SHELL_RC_PATH:-$HOME/.bashrc}"

if ! command -v go >/dev/null 2>&1; then
  echo "error: go is required but not found on PATH" >&2
  exit 1
fi

chmod +x "$RUNNER"
mkdir -p "$REPO_ROOT/bin"
(cd "$REPO_ROOT" && go build -o "$REPO_ROOT/bin/swarmy" ./cmd/swarmy)

mkdir -p "$(dirname "$RC_PATH")"
touch "$RC_PATH"

start_marker="# >>> swarmy alias >>>"
end_marker="# <<< swarmy alias <<<"
alias_line="alias swarmy='$RUNNER'"

tmp_file="$(mktemp)"
awk -v s="$start_marker" -v e="$end_marker" '
BEGIN { skip=0 }
$0 == s { skip=1; next }
$0 == e { skip=0; next }
!skip { print }
' "$RC_PATH" > "$tmp_file"

{
  cat "$tmp_file"
  echo "$start_marker"
  echo "$alias_line"
  echo "$end_marker"
} > "$RC_PATH"
rm -f "$tmp_file"

echo "swarmy installed."
echo "Alias added to: $RC_PATH"
echo "Run: source $RC_PATH"
echo "Then use: swarmy --task \"...\" --agents \"codex:1,claude:1,gemini:1\""
