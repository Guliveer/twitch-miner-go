@echo off
setlocal

REM Build and run twitch-miner-go (local development defaults)
REM Adds flags for local dev: suppress lifecycle notifications, skip unauth accounts, hide banner.
REM Any additional flags are passed through.
REM Usage: _run-localdev.bat [flags]
REM Example: _run-localdev.bat -config configs -port 9090 -log-level debug

cd /d "%~dp0"
set TWITCH_MINER_RUN_LOCALDEV=1
_run.bat -no-lifecycle-notify -skip-unauth -no-banner %*
