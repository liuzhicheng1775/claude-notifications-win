@echo off
setlocal enabledelayedexpansion

set "NOTIFY_EXE=%~dp0notify.exe"
set "SCRIPT_DIR=%~dp0"

rem Check if binary exists
if not exist "%NOTIFY_EXE%" (
    echo [claude-notifications-win] Binary not found, running bootstrap...
    powershell -ExecutionPolicy Bypass -File "%SCRIPT_DIR%bootstrap.ps1"
)

rem Check again after bootstrap
if not exist "%NOTIFY_EXE%" (
    echo [claude-notifications-win] ERROR: notify.exe not found after bootstrap
    exit /b 1
)

rem Execute the hook
"%NOTIFY_EXE%" %*

endlocal
