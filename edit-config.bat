@echo off
setlocal

REM Open the Config Editor
REM Usage: edit-config.bat [--config DIR] [--port PORT] [--tui] [--no-browser]

set "SCRIPT_DIR=%~dp0"
set "BINARY=%SCRIPT_DIR%config-editor.exe"

echo Building config-editor...
cd /d "%SCRIPT_DIR%"
go build -o config-editor.exe ./cmd/config-editor
if %errorlevel% neq 0 (
    echo Build failed.
    pause
    exit /b 1
)

"%BINARY%" %*

echo.
echo Config editor exited with code %errorlevel%.
pause
