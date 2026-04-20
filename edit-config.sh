#!/usr/bin/env bash
set -euo pipefail

# Open the Config Editor GUI in browser
# Usage: ./edit-config.sh [--port PORT] [--config DIR]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/tools/config-editor"

if ! command -v node &>/dev/null; then
  echo "Node.js is required but not installed."
  echo "Download it from https://nodejs.org/"
  exit 1
fi

if [ ! -d "node_modules" ]; then
  echo "Installing dependencies..."
  npm install --silent
fi

node server.js "$@"
