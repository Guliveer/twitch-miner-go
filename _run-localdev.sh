#!/usr/bin/env bash
set -euo pipefail

# Build and run twitch-miner-go (local development defaults)
# Adds flags for local dev: suppress lifecycle notifications, skip unauth accounts, hide banner.
# Any additional flags are passed through.
# Usage: ./_run-localdev.sh [flags]
# Example: ./_run-localdev.sh -config configs -port 9090 -log-level debug

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

export TWITCH_MINER_RUN_LOCALDEV=1
exec "$PROJECT_DIR/_run.sh" -no-lifecycle-notify -skip-unauth -no-banner "$@"
