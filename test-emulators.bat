@echo off
REM EmuBuddy Emulator Test Script - Windows Batch Wrapper
REM This script launches the PowerShell test script

setlocal

REM Check if PowerShell is available
where powershell >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo ERROR: PowerShell not found!
    echo This test script requires PowerShell.
    pause
    exit /b 1
)

REM Run the PowerShell script
powershell -ExecutionPolicy Bypass -File "%~dp0test-emulators.ps1" %*

pause
