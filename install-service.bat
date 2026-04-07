@echo off
setlocal enabledelayedexpansion

REM ================================================================
REM  Twitch Miner Go - Windows Service Installer
REM  Uses NSSM (Non-Sucking Service Manager) to run as a service.
REM  The service rebuilds the binary on every start.
REM
REM  Usage: install-service.bat [install|uninstall|start|stop|restart|status]
REM  Requires: Administrator privileges, Go toolchain
REM ================================================================

set SERVICE_NAME=twitch-miner-go
set SCRIPT_DIR=%~dp0
if "%SCRIPT_DIR:~-1%"=="\" set "SCRIPT_DIR=%SCRIPT_DIR:~0,-1%"
set BINARY_NAME=twitch-miner-go.exe
set WRAPPER_SCRIPT=%SCRIPT_DIR%\_service-start.bat
set NSSM_VERSION=2.24
set NSSM_DIR=%SCRIPT_DIR%\tools\nssm
set LOG_DIR=%SCRIPT_DIR%\logs

REM -- Admin check -------------------------------------------------
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo [x] This script requires administrator privileges.
    echo     Right-click and select "Run as administrator".
    pause
    exit /b 1
)

REM -- Route command -----------------------------------------------
set CMD=%~1
if "%CMD%"=="" set CMD=install

if /i "%CMD%"=="install"   goto :do_install
if /i "%CMD%"=="uninstall" goto :do_uninstall
if /i "%CMD%"=="remove"    goto :do_uninstall
if /i "%CMD%"=="start"     goto :do_start
if /i "%CMD%"=="stop"      goto :do_stop
if /i "%CMD%"=="restart"   goto :do_restart
if /i "%CMD%"=="status"    goto :do_status
if /i "%CMD%"=="-h"        goto :usage
if /i "%CMD%"=="--help"    goto :usage
if /i "%CMD%"=="help"      goto :usage
echo [x] Unknown command: %CMD%
goto :usage

REM ================================================================
REM  INSTALL
REM ================================================================
:do_install
echo.
echo  +------------------------------------------------+
echo  ^|  Twitch Miner Go - Windows Service Installer   ^|
echo  +------------------------------------------------+
echo.

REM -- Preflight ---------------------------------------------------
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [x] Go toolchain not found in PATH.
    echo     Install from https://go.dev/dl/
    pause
    exit /b 1
)
echo [+] Go toolchain found.

call :find_nssm
if %errorlevel% neq 0 (
    pause
    exit /b 1
)
echo [+] NSSM: !NSSM!

REM -- Check existing service --------------------------------------
sc query %SERVICE_NAME% >nul 2>&1
if !errorlevel! equ 0 (
    echo [!] Service '%SERVICE_NAME%' is already installed.
    set /p "REINSTALL=[?] Reinstall? (y/N): "
    if /i not "!REINSTALL!"=="y" (
        echo     Aborted.
        pause
        exit /b 0
    )
    echo [+] Removing existing service...
    "!NSSM!" stop %SERVICE_NAME% >nul 2>&1
    timeout /t 3 /nobreak >nul
    "!NSSM!" remove %SERVICE_NAME% confirm >nul 2>&1
)

REM -- Wizard ------------------------------------------------------
set "SVC_CONFIG=%SCRIPT_DIR%\configs"
set "SVC_PORT=8080"
set "SVC_LOG_LEVEL=INFO"

echo.
echo     Default values shown in brackets. Press Enter to accept.
echo.
set /p "SVC_CONFIG=[?] Config directory [!SVC_CONFIG!]: "
if "!SVC_CONFIG!"=="" set "SVC_CONFIG=%SCRIPT_DIR%\configs"

set /p "SVC_PORT=[?] HTTP port [!SVC_PORT!]: "
if "!SVC_PORT!"=="" set "SVC_PORT=8080"

set /p "SVC_LOG_LEVEL=[?] Log level - DEBUG/INFO/WARN/ERROR [!SVC_LOG_LEVEL!]: "
if "!SVC_LOG_LEVEL!"=="" set "SVC_LOG_LEVEL=INFO"

set "SVC_AUTOSTART=Y"
set /p "SVC_AUTOSTART=[?] Start on boot (autostart)? (Y/n) [!SVC_AUTOSTART!]: "
if "!SVC_AUTOSTART!"=="" set "SVC_AUTOSTART=Y"

echo.
echo [+] Summary:
echo     Service:   %SERVICE_NAME%
echo     Project:   %SCRIPT_DIR%
echo     Config:    !SVC_CONFIG!
echo     Port:      !SVC_PORT!
echo     Log level: !SVC_LOG_LEVEL!
echo     Autostart: !SVC_AUTOSTART!
echo     Logs:      %LOG_DIR%\service.log
echo.

set /p "CONFIRM=[?] Proceed with installation? (Y/n): "
if /i "!CONFIRM!"=="n" (
    echo     Aborted.
    pause
    exit /b 0
)

REM -- Initial build (verify it works) -----------------------------
echo.
echo [+] Building %BINARY_NAME%...
call :build_binary
if !errorlevel! neq 0 (
    echo [x] Build failed! Fix build errors before installing.
    pause
    exit /b 1
)
echo [+] Build successful.

REM -- Create wrapper script ---------------------------------------
echo [+] Creating service wrapper...
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"

> "%WRAPPER_SCRIPT%" echo @echo off
>> "%WRAPPER_SCRIPT%" echo setlocal
>> "%WRAPPER_SCRIPT%" echo cd /d "%SCRIPT_DIR%"
>> "%WRAPPER_SCRIPT%" echo call run.bat -config "!SVC_CONFIG!" -port !SVC_PORT! -log-level !SVC_LOG_LEVEL!

REM -- Install via NSSM --------------------------------------------
echo [+] Installing service via NSSM...
"!NSSM!" install %SERVICE_NAME% "%WRAPPER_SCRIPT%"
if !errorlevel! neq 0 (
    echo [x] NSSM install failed!
    pause
    exit /b 1
)

REM Configure service behavior
"!NSSM!" set %SERVICE_NAME% AppDirectory "%SCRIPT_DIR%" >nul
"!NSSM!" set %SERVICE_NAME% Description "Twitch Channel Points Miner (Go)" >nul
"!NSSM!" set %SERVICE_NAME% AppStdout "%LOG_DIR%\service.log" >nul
"!NSSM!" set %SERVICE_NAME% AppStderr "%LOG_DIR%\service.log" >nul
"!NSSM!" set %SERVICE_NAME% AppStdoutCreationDisposition 4 >nul
"!NSSM!" set %SERVICE_NAME% AppStderrCreationDisposition 4 >nul
"!NSSM!" set %SERVICE_NAME% AppRotateFiles 1 >nul
"!NSSM!" set %SERVICE_NAME% AppRotateBytes 10485760 >nul
"!NSSM!" set %SERVICE_NAME% AppExit Default Restart >nul
"!NSSM!" set %SERVICE_NAME% AppRestartDelay 10000 >nul

REM Configure startup type
if /i "!SVC_AUTOSTART!"=="y" (
    "!NSSM!" set %SERVICE_NAME% Start SERVICE_AUTO_START >nul
    echo [+] Service set to start automatically on boot.
) else (
    "!NSSM!" set %SERVICE_NAME% Start SERVICE_DEMAND_START >nul
    echo [+] Service set to manual start.
)

echo [+] Service installed successfully.

echo.
set /p "START_NOW=[?] Start the service now? (Y/n): "
if /i not "!START_NOW!"=="n" (
    "!NSSM!" start %SERVICE_NAME%
    timeout /t 2 /nobreak >nul
    echo.
    "!NSSM!" status %SERVICE_NAME%
)

echo.
echo [+] Installation complete!
echo.
echo   Useful commands:
echo     install-service.bat start     Start the service
echo     install-service.bat stop      Stop the service
echo     install-service.bat restart   Restart (triggers rebuild)
echo     install-service.bat status    Show service status
echo     install-service.bat uninstall Remove the service
echo.
echo   Logs: %LOG_DIR%\service.log
echo.
echo   NOTE: The service runs under the SYSTEM account.
echo   Ensure 'go' and 'git' are in the system PATH (not just user PATH).
echo.
pause
goto :eof

REM ================================================================
REM  UNINSTALL
REM ================================================================
:do_uninstall
call :find_nssm
if %errorlevel% neq 0 (
    pause
    exit /b 1
)

sc query %SERVICE_NAME% >nul 2>&1
if !errorlevel! neq 0 (
    echo [x] Service '%SERVICE_NAME%' is not installed.
    pause
    exit /b 1
)

echo.
echo [!] This will stop and remove the '%SERVICE_NAME%' service.
set /p "CONFIRM=[?] Continue? (y/N): "
if /i not "!CONFIRM!"=="y" (
    echo     Aborted.
    goto :eof
)

echo [+] Stopping service...
"!NSSM!" stop %SERVICE_NAME% >nul 2>&1
timeout /t 3 /nobreak >nul

echo [+] Removing service...
"!NSSM!" remove %SERVICE_NAME% confirm
if !errorlevel! neq 0 (
    echo [x] Failed to remove service. It may still be stopping.
    echo     Wait a moment and try again.
    pause
    exit /b 1
)

if exist "%WRAPPER_SCRIPT%" (
    del "%WRAPPER_SCRIPT%"
    echo [+] Wrapper script removed.
)

echo [+] Service '%SERVICE_NAME%' removed.
pause
goto :eof

REM ================================================================
REM  START / STOP / RESTART / STATUS
REM ================================================================
:do_start
call :find_nssm
if %errorlevel% neq 0 exit /b 1
echo [+] Starting %SERVICE_NAME% (will rebuild first)...
"!NSSM!" start %SERVICE_NAME%
goto :eof

:do_stop
call :find_nssm
if %errorlevel% neq 0 exit /b 1
echo [+] Stopping %SERVICE_NAME%...
"!NSSM!" stop %SERVICE_NAME%
goto :eof

:do_restart
call :find_nssm
if %errorlevel% neq 0 exit /b 1
echo [+] Restarting %SERVICE_NAME% (will rebuild)...
"!NSSM!" restart %SERVICE_NAME%
goto :eof

:do_status
call :find_nssm
if %errorlevel% neq 0 exit /b 1
"!NSSM!" status %SERVICE_NAME%
echo.
sc query %SERVICE_NAME% 2>nul
goto :eof

REM ================================================================
REM  HELPERS
REM ================================================================

:find_nssm
where nssm >nul 2>&1
if %errorlevel% equ 0 (
    for /f "tokens=*" %%i in ('where nssm') do set "NSSM=%%i"
    exit /b 0
)
if exist "%NSSM_DIR%\nssm.exe" (
    set "NSSM=%NSSM_DIR%\nssm.exe"
    exit /b 0
)

echo [!] NSSM not found. Attempting to download...
call :download_nssm
exit /b %errorlevel%

:download_nssm
if not exist "%NSSM_DIR%" mkdir "%NSSM_DIR%"

echo [+] Downloading NSSM %NSSM_VERSION%...
powershell -NoProfile -Command ^
    "try { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://nssm.cc/release/nssm-%NSSM_VERSION%.zip' -OutFile '%TEMP%\nssm.zip' -UseBasicParsing } catch { Write-Host $_.Exception.Message; exit 1 }"
if %errorlevel% neq 0 (
    echo [x] Download failed. Install NSSM manually:
    echo       winget install NSSM.NSSM
    echo       choco install nssm
    echo       https://nssm.cc/download
    exit /b 1
)

echo [+] Extracting...
powershell -NoProfile -Command "Expand-Archive -Path '%TEMP%\nssm.zip' -DestinationPath '%TEMP%\nssm-extract' -Force"

if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    copy /y "%TEMP%\nssm-extract\nssm-%NSSM_VERSION%\win64\nssm.exe" "%NSSM_DIR%\nssm.exe" >nul
) else (
    copy /y "%TEMP%\nssm-extract\nssm-%NSSM_VERSION%\win32\nssm.exe" "%NSSM_DIR%\nssm.exe" >nul
)

del /q "%TEMP%\nssm.zip" 2>nul
rmdir /s /q "%TEMP%\nssm-extract" 2>nul

if exist "%NSSM_DIR%\nssm.exe" (
    set "NSSM=%NSSM_DIR%\nssm.exe"
    echo [+] NSSM installed to %NSSM_DIR%\nssm.exe
    exit /b 0
)
exit /b 1

:build_binary
cd /d "%SCRIPT_DIR%"

if exist "%SCRIPT_DIR%\VERSION" (
    set /p VERSION=<"%SCRIPT_DIR%\VERSION"
) else (
    set VERSION=dev
)
for /f "tokens=*" %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i
if not defined GIT_COMMIT set GIT_COMMIT=unknown

set LDFLAGS=-X github.com/Guliveer/twitch-miner-go/internal/version.Number=!VERSION! -X github.com/Guliveer/twitch-miner-go/internal/version.GitCommit=!GIT_COMMIT!
go build -ldflags "!LDFLAGS!" -o "%SCRIPT_DIR%\%BINARY_NAME%" ./cmd/twitch-miner-go
exit /b %errorlevel%

:usage
echo Usage: %~nx0 [command]
echo.
echo Commands:
echo   install     Install as a Windows service (default)
echo   uninstall   Stop and remove the service
echo   start       Start the service
echo   stop        Stop the service
echo   restart     Restart the service (triggers a rebuild)
echo   status      Show service status
echo.
echo Requires NSSM (auto-downloaded if not found).
exit /b 0
