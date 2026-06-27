#!/usr/bin/env bash
set -euo pipefail

# Open the Config Editor
# Usage: ./edit-config.sh [--config DIR] [--port PORT] [--tui] [--no-browser]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="$SCRIPT_DIR/config-editor"

if [ ! -f "$BINARY" ]; then
  echo "Building config-editor..."
  cd "$SCRIPT_DIR"
  go build -o config-editor ./cmd/config-editor
fi

exec "$BINARY" "$@"
