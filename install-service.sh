#!/usr/bin/env bash
set -euo pipefail

# ─────────────────────────────────────────────────
# Twitch Miner Go — service installer (systemd / OpenRC)
# Usage: sudo ./install-service.sh
# ─────────────────────────────────────────────────

readonly DEFAULT_SERVICE_NAME="twitch-miner-go"
readonly DEFAULT_INSTALL_DIR="/usr/local/bin"
readonly DEFAULT_CONFIG_DIR="/etc/${DEFAULT_SERVICE_NAME}/configs"
readonly DEFAULT_DATA_DIR="/var/lib/${DEFAULT_SERVICE_NAME}"
readonly DEFAULT_ENV_FILE="/etc/${DEFAULT_SERVICE_NAME}/.env"
readonly DEFAULT_PORT="8080"
readonly DEFAULT_LOG_LEVEL="INFO"
readonly DEFAULT_OPENRC_LOG_DIR="/var/log"

# Detected init system: "systemd" or "openrc" (set by detect_init_system)
INIT_SYSTEM=""
# Config mode: "copy" or "symlink" (set by wizard)
CONFIG_MODE="copy"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ── Helpers ──────────────────────────────────────

info()  { echo -e "${GREEN}[+]${NC} $*"; }
warn()  { echo -e "${YELLOW}[!]${NC} $*"; }
error() { echo -e "${RED}[x]${NC} $*" >&2; }
ask()   { echo -en "${CYAN}[?]${NC} $1 ${BOLD}[$2]${NC}: "; }

# Prompt with a default value. Empty input = default.
prompt() {
    local var_name="$1" prompt_text="$2" default="$3"
    ask "$prompt_text" "$default"
    local input
    read -r input
    eval "$var_name=\"${input:-$default}\""
}

confirm() {
    local prompt_text="$1" default="${2:-y}"
    ask "$prompt_text (y/n)" "$default"
    local input
    read -r input
    input="${input:-$default}"
    [[ "${input,,}" == "y" || "${input,,}" == "yes" ]]
}

banner() {
    echo -e "${BOLD}"
    echo "╔═══════════════════════════════════════════════╗"
    echo "║  Twitch Miner Go — Service Installer          ║"
    echo "║  (systemd / OpenRC)                           ║"
    echo "╚═══════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# ── Preflight checks ────────────────────────────

detect_init_system() {
    if command -v systemctl &>/dev/null && systemctl --version &>/dev/null; then
        INIT_SYSTEM="systemd"
    elif command -v rc-service &>/dev/null; then
        INIT_SYSTEM="openrc"
        # Require supervise-daemon for process supervision (OpenRC >= 0.21)
        if ! command -v supervise-daemon &>/dev/null; then
            error "OpenRC detected but 'supervise-daemon' is missing (requires OpenRC >= 0.21)."
            error "Please upgrade OpenRC: apk upgrade openrc"
            exit 1
        fi
    else
        error "No supported init system found (systemd or OpenRC)."
        exit 1
    fi
}

preflight() {
    if [[ "$(uname -s)" != "Linux" ]]; then
        error "This script only works on Linux."
        exit 1
    fi

    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root (use sudo)."
        exit 1
    fi

    detect_init_system
    info "Detected init system: ${BOLD}${INIT_SYSTEM}${NC}"
}

# ── Find or build the binary ────────────────────

resolve_binary() {
    # 1. Pre-built binary next to this script
    if [[ -x "${SCRIPT_DIR}/twitch-miner-go" ]]; then
        BINARY_SOURCE="${SCRIPT_DIR}/twitch-miner-go"
        return
    fi

    # 2. Already installed
    if command -v twitch-miner-go &>/dev/null; then
        BINARY_SOURCE="$(command -v twitch-miner-go)"
        return
    fi

    # 3. Offer to build from source
    if [[ -f "${SCRIPT_DIR}/go.mod" ]] && command -v go &>/dev/null; then
        if confirm "Binary not found. Build from source?"; then
            info "Building..."
            local version git_commit ldflags
            version=$(cat "${SCRIPT_DIR}/VERSION" 2>/dev/null || echo "dev")
            git_commit=$(git -C "$SCRIPT_DIR" rev-parse --short HEAD 2>/dev/null || echo "unknown")
            ldflags="-X github.com/Guliveer/twitch-miner-go/internal/version.Number=${version} -X github.com/Guliveer/twitch-miner-go/internal/version.GitCommit=${git_commit}"
            go build -ldflags "${ldflags}" -o "${SCRIPT_DIR}/twitch-miner-go" "${SCRIPT_DIR}/cmd/twitch-miner-go"
            BINARY_SOURCE="${SCRIPT_DIR}/twitch-miner-go"
            return
        fi
    fi

    error "Cannot find twitch-miner-go binary. Build it first with ./run.sh or place it next to this script."
    exit 1
}

# ── Wizard ───────────────────────────────────────

wizard() {
    banner
    info "This wizard will set up twitch-miner-go as a service (${INIT_SYSTEM}).\n"

    prompt SERVICE_NAME  "Service name"                "$DEFAULT_SERVICE_NAME"

    # Compute defaults based on chosen service name
    local def_config_dir="/etc/${SERVICE_NAME}/configs"
    local def_data_dir="/var/lib/${SERVICE_NAME}"
    local def_env_file="/etc/${SERVICE_NAME}/.env"

    prompt INSTALL_DIR   "Binary install directory"    "$DEFAULT_INSTALL_DIR"
    prompt CONFIG_DIR    "Config directory"            "$def_config_dir"
    prompt DATA_DIR      "Data directory (cookies)"    "$def_data_dir"
    prompt ENV_FILE      "Environment file (.env)"     "$def_env_file"
    prompt PORT          "HTTP port"                   "$DEFAULT_PORT"
    prompt LOG_LEVEL     "Log level (DEBUG/INFO/WARN/ERROR)" "$DEFAULT_LOG_LEVEL"

    # Detect current user (the one who ran sudo)
    local default_user="${SUDO_USER:-$(whoami)}"
    prompt RUN_USER "Run as user" "$default_user"

    # Config deployment mode
    if [[ -d "${SCRIPT_DIR}/configs" ]]; then
        echo ""
        info "How should config files be deployed to ${CONFIG_DIR}?"
        echo "  copy    — copy files (edit in ${CONFIG_DIR}, use '$0 sync' to re-copy from source)"
        echo "  symlink — symlink to source directory (edits in either location take effect)"
        prompt CONFIG_MODE "Config mode (copy/symlink)" "copy"
        if [[ "$CONFIG_MODE" != "copy" && "$CONFIG_MODE" != "symlink" ]]; then
            warn "Invalid config mode '${CONFIG_MODE}', defaulting to 'copy'."
            CONFIG_MODE="copy"
        fi
    fi

    # Warn about home directory paths (systemd ProtectHome=true blocks /home access)
    if [[ "$INIT_SYSTEM" == "systemd" ]]; then
        local home_paths=()
        for p in "$CONFIG_DIR" "$DATA_DIR" "$ENV_FILE"; do
            [[ "$p" == /home/* ]] && home_paths+=("$p")
        done
        if [[ ${#home_paths[@]} -gt 0 ]]; then
            echo ""
            warn "The following paths are under /home/:"
            for p in "${home_paths[@]}"; do
                echo "    ${p}"
            done
            warn "systemd's ProtectHome=true will block access to these paths."
            warn "Move them outside /home/ or the service will fail to read them."
            echo ""
            if ! confirm "Continue anyway?"; then
                warn "Aborted. Re-run the wizard with paths outside /home/."
                exit 0
            fi
        fi
    fi

    echo ""
    info "Summary:"
    echo "  Service:    ${SERVICE_NAME}"
    echo "  Binary:     ${INSTALL_DIR}/${SERVICE_NAME}"
    echo "  Config:     ${CONFIG_DIR}"
    echo "  Data:       ${DATA_DIR}"
    echo "  Env file:   ${ENV_FILE}"
    echo "  Port:       ${PORT}"
    echo "  Log level:  ${LOG_LEVEL}"
    echo "  User:       ${RUN_USER}"
    echo "  Config mode: ${CONFIG_MODE}"
    echo ""

    if ! confirm "Proceed with installation?"; then
        warn "Aborted."
        exit 0
    fi
}

# ── Install ──────────────────────────────────────

do_install() {
    local bin_dest="${INSTALL_DIR}/${SERVICE_NAME}"

    # Create directories
    info "Creating directories..."
    mkdir -p "$INSTALL_DIR" "$CONFIG_DIR" "$DATA_DIR" "$(dirname "$ENV_FILE")"

    # Copy binary
    info "Installing binary to ${bin_dest}..."
    cp -f "$BINARY_SOURCE" "$bin_dest"
    chmod 755 "$bin_dest"

    # Deploy config files
    if [[ -d "${SCRIPT_DIR}/configs" ]]; then
        if [[ "$CONFIG_MODE" == "symlink" ]]; then
            # Remove target dir if empty and create symlink to source
            if [[ -d "$CONFIG_DIR" ]] && [[ ! "$(ls -A "$CONFIG_DIR" 2>/dev/null)" ]]; then
                rmdir "$CONFIG_DIR"
            fi
            if [[ ! -e "$CONFIG_DIR" ]]; then
                info "Symlinking ${CONFIG_DIR} -> ${SCRIPT_DIR}/configs..."
                ln -sfn "${SCRIPT_DIR}/configs" "$CONFIG_DIR"
            else
                warn "Config directory already exists and is not empty, skipping symlink."
            fi
        else
            # Copy mode
            if [[ ! "$(ls -A "$CONFIG_DIR" 2>/dev/null)" ]]; then
                info "Copying config files to ${CONFIG_DIR}..."
                cp -r "${SCRIPT_DIR}/configs/"* "$CONFIG_DIR/" 2>/dev/null || true
            fi
        fi
    fi

    # Create .env if it doesn't exist
    if [[ ! -f "$ENV_FILE" ]]; then
        info "Creating empty env file at ${ENV_FILE}..."
        cat > "$ENV_FILE" <<'ENVEOF'
# Twitch Miner Go — environment variables
# See DEPLOYMENT.md for details on required values.

# TWITCH_CLIENT_ID_TV=
# TWITCH_CLIENT_ID_BROWSER=
# TWITCH_CLIENT_VERSION=
# TWITCH_AUTH_TOKEN_YOUR_USERNAME=
ENVEOF
    fi

    # Set ownership
    info "Setting permissions for user '${RUN_USER}'..."
    chown -R "${RUN_USER}:" "$CONFIG_DIR" "$DATA_DIR" "$(dirname "$ENV_FILE")"
    chmod 600 "$ENV_FILE"

    if [[ "$INIT_SYSTEM" == "systemd" ]]; then
        install_systemd "$bin_dest"
    else
        install_openrc "$bin_dest"
    fi
}

install_systemd() {
    local bin_dest="$1"
    local service_file="/etc/systemd/system/${SERVICE_NAME}.service"

    # Generate systemd unit
    info "Writing systemd unit to ${service_file}..."
    cat > "$service_file" <<EOF
[Unit]
Description=Twitch Channel Points Miner (Go)
Documentation=https://github.com/Guliveer/twitch-miner-go
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${RUN_USER}
WorkingDirectory=${DATA_DIR}
EnvironmentFile=-${ENV_FILE}
ExecStart=${bin_dest} -config ${CONFIG_DIR} -port ${PORT} -log-level ${LOG_LEVEL}
Restart=on-failure
RestartSec=10
TimeoutStopSec=30

# Hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${DATA_DIR}
ReadOnlyPaths=${CONFIG_DIR}
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

    # Reload systemd
    info "Reloading systemd daemon..."
    systemctl daemon-reload

    # Enable & start
    if confirm "Enable service to start on boot?"; then
        systemctl enable "$SERVICE_NAME"
        info "Service enabled."
    fi

    if confirm "Start the service now?"; then
        systemctl start "$SERVICE_NAME"
        info "Service started."
        echo ""
        systemctl --no-pager status "$SERVICE_NAME" || true
    fi

    echo ""
    info "Installation complete!"
    echo ""
    echo "  Useful commands:"
    echo "    systemctl status  ${SERVICE_NAME}"
    echo "    systemctl stop    ${SERVICE_NAME}"
    echo "    systemctl restart ${SERVICE_NAME}"
    echo "    journalctl -u     ${SERVICE_NAME} -f"
    echo ""
    echo "  Config:  ${CONFIG_DIR}"
    echo "  Env:     ${ENV_FILE}"
    echo "  Logs:    journalctl -u ${SERVICE_NAME}"
    echo ""
}

install_openrc() {
    local bin_dest="$1"
    local init_script="/etc/init.d/${SERVICE_NAME}"
    local log_file="${DEFAULT_OPENRC_LOG_DIR}/${SERVICE_NAME}.log"

    # Create log file with correct ownership
    info "Creating log file at ${log_file}..."
    touch "$log_file"
    chown "${RUN_USER}:" "$log_file"

    # Generate OpenRC init script
    info "Writing OpenRC init script to ${init_script}..."
    cat > "$init_script" <<EOF
#!/sbin/openrc-run
# Twitch Channel Points Miner (Go) — OpenRC service

name="Twitch Channel Points Miner (Go)"
description="Twitch Channel Points Miner (Go)"

supervisor=supervise-daemon
command="${bin_dest}"
command_args="-config ${CONFIG_DIR} -port ${PORT} -log-level ${LOG_LEVEL}"
command_user="${RUN_USER}"

directory="${DATA_DIR}"
pidfile="/run/\${RC_SVCNAME}.pid"
output_log="${log_file}"
error_log="${log_file}"
respawn_delay=10

# Load environment variables
start_pre() {
    if [ -f "${ENV_FILE}" ]; then
        set -a
        # shellcheck disable=SC1091
        . "${ENV_FILE}"
        set +a
    fi
}

depend() {
    need net
    after firewall
}
EOF
    chmod 755 "$init_script"

    # Enable & start
    if confirm "Enable service to start on boot?"; then
        rc-update add "$SERVICE_NAME" default
        info "Service added to default runlevel."
    fi

    if confirm "Start the service now?"; then
        rc-service "$SERVICE_NAME" start
        info "Service started."
        echo ""
        rc-service "$SERVICE_NAME" status || true
    fi

    echo ""
    info "Installation complete!"
    echo ""
    echo "  Useful commands:"
    echo "    rc-service ${SERVICE_NAME} status"
    echo "    rc-service ${SERVICE_NAME} stop"
    echo "    rc-service ${SERVICE_NAME} restart"
    echo "    tail -f ${log_file}"
    echo ""
    echo "  Config:  ${CONFIG_DIR}"
    echo "  Env:     ${ENV_FILE}"
    echo "  Logs:    ${log_file}"
    echo ""
}

# ── Uninstall ────────────────────────────────────

do_uninstall() {
    banner
    prompt SERVICE_NAME "Service name to remove" "$DEFAULT_SERVICE_NAME"

    if [[ "$INIT_SYSTEM" == "systemd" ]]; then
        local service_file="/etc/systemd/system/${SERVICE_NAME}.service"
        if [[ ! -f "$service_file" ]]; then
            error "Service file not found: ${service_file}"
            exit 1
        fi
    else
        local service_file="/etc/init.d/${SERVICE_NAME}"
        if [[ ! -f "$service_file" ]]; then
            error "Init script not found: ${service_file}"
            exit 1
        fi
    fi

    echo ""
    warn "This will stop and remove the ${INIT_SYSTEM} service."
    warn "Binary, configs, and data will NOT be deleted."
    echo ""

    if ! confirm "Proceed with uninstall?"; then
        warn "Aborted."
        exit 0
    fi

    if [[ "$INIT_SYSTEM" == "systemd" ]]; then
        info "Stopping service..."
        systemctl stop "$SERVICE_NAME" 2>/dev/null || true

        info "Disabling service..."
        systemctl disable "$SERVICE_NAME" 2>/dev/null || true

        info "Removing service file..."
        rm -f "$service_file"

        info "Reloading systemd daemon..."
        systemctl daemon-reload
    else
        info "Stopping service..."
        rc-service "$SERVICE_NAME" stop 2>/dev/null || true

        info "Removing from default runlevel..."
        rc-update del "$SERVICE_NAME" default 2>/dev/null || true

        info "Removing init script..."
        rm -f "$service_file"
    fi

    info "Service '${SERVICE_NAME}' removed."
    echo ""
    echo "  Remaining files (remove manually if needed):"
    echo "    Binary:  ${DEFAULT_INSTALL_DIR}/${SERVICE_NAME}"
    echo "    Config:  /etc/${SERVICE_NAME}/configs"
    echo "    Data:    /var/lib/${SERVICE_NAME}"
    echo "    Env:     /etc/${SERVICE_NAME}/.env"
    if [[ "$INIT_SYSTEM" == "openrc" ]]; then
        echo "    Logs:    ${DEFAULT_OPENRC_LOG_DIR}/${SERVICE_NAME}.log"
    fi
    echo ""
}

# ── Status ───────────────────────────────────────

do_status() {
    local name="${1:-$DEFAULT_SERVICE_NAME}"

    if [[ "$INIT_SYSTEM" == "systemd" ]]; then
        if systemctl list-unit-files "${name}.service" &>/dev/null; then
            systemctl --no-pager status "$name" || true
            echo ""
            echo "Recent logs:"
            journalctl -u "$name" -n 20 --no-pager 2>/dev/null || true
        else
            error "Service '${name}' not found."
            exit 1
        fi
    else
        if [[ ! -f "/etc/init.d/${name}" ]]; then
            error "Service '${name}' not found (/etc/init.d/${name})."
            exit 1
        fi
        rc-service "$name" status || true
        echo ""
        local log_file="${DEFAULT_OPENRC_LOG_DIR}/${name}.log"
        if [[ -f "$log_file" ]]; then
            echo "Recent logs (${log_file}):"
            tail -n 20 "$log_file" 2>/dev/null || true
        else
            warn "Log file not found: ${log_file}"
        fi
    fi
}

# ── Sync configs ────────────────────────────────

do_sync() {
    local name="${1:-$DEFAULT_SERVICE_NAME}"
    local target_dir="$DEFAULT_CONFIG_DIR"
    local source_dir="${SCRIPT_DIR}/configs"

    if [[ ! -d "$source_dir" ]]; then
        error "Source config directory not found: ${source_dir}"
        exit 1
    fi

    if [[ -L "$target_dir" ]]; then
        warn "Config directory is a symlink (${target_dir} -> $(readlink -f "$target_dir"))."
        warn "Sync is not needed — changes are shared automatically."
        exit 0
    fi

    if [[ ! -d "$target_dir" ]]; then
        error "Target config directory not found: ${target_dir}"
        error "Is the service installed?"
        exit 1
    fi

    info "Syncing configs from ${source_dir} to ${target_dir}..."
    cp -r "${source_dir}/"* "$target_dir/"
    info "Config files updated."

    if confirm "Restart the service to apply changes?"; then
        if [[ "$INIT_SYSTEM" == "systemd" ]]; then
            systemctl restart "$name"
        else
            rc-service "$name" restart
        fi
        info "Service restarted."
    fi
}

# ── Main ─────────────────────────────────────────

usage() {
    echo "Usage: sudo $0 [install|uninstall|status|sync]"
    echo ""
    echo "Commands:"
    echo "  install     Interactive wizard to create a service (default)"
    echo "  uninstall   Stop and remove the service"
    echo "  status      Show service status and recent logs"
    echo "  sync        Re-copy config files from source and optionally restart"
    echo ""
    echo "Supported init systems: systemd, OpenRC (Alpine)"
}

main() {
    local command="${1:-install}"

    case "$command" in
        install)
            preflight
            resolve_binary
            wizard
            do_install
            ;;
        uninstall|remove)
            preflight
            do_uninstall
            ;;
        status)
            preflight
            do_status "${2:-}"
            ;;
        sync)
            preflight
            do_sync "${2:-}"
            ;;
        -h|--help|help)
            usage
            ;;
        *)
            error "Unknown command: $command"
            usage
            exit 1
            ;;
    esac
}

main "$@"
