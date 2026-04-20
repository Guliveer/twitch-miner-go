@echo off
setlocal

REM Open the Config Editor GUI in browser
REM Usage: edit-config.bat [--port PORT] [--config DIR]

cd /d "%~dp0\tools\config-editor"

where node >nul 2>&1
if %errorlevel% neq 0 (
    echo Node.js is required but not installed.
    echo Download it from https://nodejs.org/
    pause
    exit /b 1
)

if not exist "node_modules" (
    echo Installing dependencies...
    npm install --silent
)

node server.js %*

echo.
echo Config editor exited with code %errorlevel%.
pause
