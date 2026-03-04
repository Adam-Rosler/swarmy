#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
go test ./e2e -count=1 -v
