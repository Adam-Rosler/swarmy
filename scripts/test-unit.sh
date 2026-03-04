#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
go test ./internal/... -count=1 -v
