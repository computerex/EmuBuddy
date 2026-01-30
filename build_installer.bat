@echo off
REM Build EmuBuddy Installer and Launcher for all platforms

echo =========================================
echo   Building EmuBuddy Cross-Platform
echo =========================================
echo.

REM =========================================
REM   Build Installer
REM =========================================
cd installer

REM Windows Installer
echo [1/6] Building Installer for Windows...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o ..\EmuBuddySetup.exe main.go

if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] Windows installer: EmuBuddySetup.exe
) else (
    echo [ERROR] Windows installer build failed
    cd ..
    pause
    exit /b 1
)

echo.

REM Linux Installer
echo [2/6] Building Installer for Linux...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o ..\EmuBuddySetup-linux main.go

if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] Linux installer: EmuBuddySetup-linux
) else (
    echo [ERROR] Linux installer build failed
)

echo.

REM macOS Installer
echo [3/6] Building Installer for macOS...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o ..\EmuBuddySetup-macos main.go

if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] macOS installer: EmuBuddySetup-macos
) else (
    echo [ERROR] macOS installer build failed
)

cd ..

REM =========================================
REM   Build Launcher
REM =========================================
echo.
echo =========================================
echo   Building Launcher (Windows only)
echo =========================================
echo NOTE: Fyne GUI requires native builds for Linux/macOS
echo.

cd launcher\gui

REM Windows Launcher (GUI mode - no console window)
echo Building Launcher for Windows...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w -H windowsgui" -o ..\..\EmuBuddyLauncher.exe .

if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] Windows launcher: EmuBuddyLauncher.exe
) else (
    echo [ERROR] Windows launcher build failed
)

cd ..\..

echo.
echo =========================================
echo   Build Summary
echo =========================================
echo.
echo Installers:
dir EmuBuddySetup* 2>nul
echo.
echo Launcher:
dir EmuBuddyLauncher.exe 2>nul
echo.
echo Done!
echo.
pause
