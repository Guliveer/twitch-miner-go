#!/usr/bin/env bash
set -euo pipefail

# Build and run twitch-miner-go
# Usage: ./run.sh [flags]
# Example: ./run.sh -config configs -port 8080 -log-level debug

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

cd "$PROJECT_DIR"

VERSION=$(cat "$PROJECT_DIR/VERSION" 2>/dev/null || echo "dev")
GIT_COMMIT=$(git -C "$PROJECT_DIR" rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS="-X github.com/Guliveer/twitch-miner-go/internal/version.Number=${VERSION} -X github.com/Guliveer/twitch-miner-go/internal/version.GitCommit=${GIT_COMMIT}"

echo "Building twitch-miner-go v${VERSION}..."
go build -ldflags "${LDFLAGS}" -o twitch-miner-go ./cmd/twitch-miner-go

echo "Starting twitch-miner-go..."
./twitch-miner-go "$@"
