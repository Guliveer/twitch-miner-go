#!/usr/bin/env bash
# Generate DASHBOARD_USER and DASHBOARD_PASSWORD_SHA256 for .env
# Usage: ./scripts/gen-dashboard-auth.sh [username] [password]
#   - If no arguments, prompts interactively (password hidden)
#   - If arguments provided, uses them directly

set -euo pipefail

if [ $# -ge 2 ]; then
    USER="$1"
    PASS="$2"
elif [ $# -eq 1 ]; then
    USER="$1"
    printf "Password: "
    read -rs PASS
    echo
else
    printf "Username: "
    read -r USER
    printf "Password: "
    read -rs PASS
    echo
fi

if [ -z "$USER" ] || [ -z "$PASS" ]; then
    echo "Error: username and password cannot be empty" >&2
    exit 1
fi

HASH=$(printf '%s' "$PASS" | openssl dgst -sha256 | awk '{print $NF}')

echo ""
echo "Add these to your .env file:"
echo ""
echo "DASHBOARD_USER=${USER}"
echo "DASHBOARD_PASSWORD_SHA256=${HASH}"
echo ""
echo "Or set them as environment variables / Fly.io secrets:"
echo ""
echo "  fly secrets set DASHBOARD_USER=${USER} DASHBOARD_PASSWORD_SHA256=${HASH}"
