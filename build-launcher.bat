@echo off
:: Quick Windows launcher build for development/testing
cd /d "%~dp0launcher\gui"

:: Build GUI version (no console window)
go build -ldflags="-s -w -H windowsgui" -o ..\..\EmuBuddyLauncher.exe .
if %ERRORLEVEL% EQU 0 (
    echo Built EmuBuddyLauncher.exe [GUI]
) else (
    echo Build failed
    exit /b 1
)

:: Build console version (for testing/headless mode)
go build -ldflags="-s -w" -o ..\..\EmuBuddyLauncher-console.exe .
if %ERRORLEVEL% EQU 0 (
    echo Built EmuBuddyLauncher-console.exe [Console/Testing]
) else (
    echo Console build failed
)
