#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOOKS_DIR="$SCRIPT_DIR/githooks"

if [ ! -d "$HOOKS_DIR" ]; then
    echo "Error: hooks directory not found at $HOOKS_DIR"
    exit 1
fi

git config core.hooksPath "$HOOKS_DIR"
echo "Git hooks installed. Using hooks from: $HOOKS_DIR"
echo ""
echo "Active hooks:"
echo "  - pre-commit  : Blocks commit if local branch is behind remote (pull first)"
echo "  - commit-msg  : Validates Conventional Commits format on each commit"
echo "  - pre-push    : Validates all commits follow Conventional Commits before push"
