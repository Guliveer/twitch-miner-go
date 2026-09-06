@echo off
setlocal

REM Build and run twitch-miner-go
REM Usage: run.bat [flags]
REM Example: run.bat -config configs -port 8080 -log-level debug

cd /d "%~dp0"

if exist "%~dp0VERSION" (
    set /p VERSION=<"%~dp0VERSION"
) else (
    set VERSION=dev
)

for /f "tokens=*" %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i
if not defined GIT_COMMIT set GIT_COMMIT=unknown

set LDFLAGS=-X github.com/Guliveer/twitch-miner-go/internal/version.Number=%VERSION% -X github.com/Guliveer/twitch-miner-go/internal/version.GitCommit=%GIT_COMMIT%

echo Building twitch-miner-go v%VERSION%...
go build -ldflags "%LDFLAGS%" -o twitch-miner-go.exe ./cmd/twitch-miner-go
if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b %errorlevel%
)

echo Starting twitch-miner-go...
set EXTRA_ARGS=-no-console
if defined TWITCH_MINER_RUN_LOCALDEV set EXTRA_ARGS=
twitch-miner-go.exe %EXTRA_ARGS% %*

if defined TWITCH_MINER_RUN_LOCALDEV goto :show_exit
exit /b %errorlevel%

:show_exit
echo.
echo twitch-miner-go exited with code %errorlevel%.
pause
