@echo off
REM Build EmuBuddy Installer for all platforms

echo =========================================
echo   Building EmuBuddy Cross-Platform
echo =========================================
echo.

cd installer

REM Windows
echo [1/3] Building for Windows...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o ..\EmuBuddySetup.exe main.go

if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] Windows build complete: EmuBuddySetup.exe
) else (
    echo [ERROR] Windows build failed
    cd ..
    pause
    exit /b 1
)
