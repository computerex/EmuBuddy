@echo off
REM Build EmuBuddy Launcher for all platforms using Docker (fyne-cross)
REM Prerequisites: Docker Desktop must be installed and running

echo =========================================
echo   Building EmuBuddy Launcher (Docker)
echo =========================================
echo.

REM Check if Docker is running
docker info >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Docker is not running or not installed
    echo Please install Docker Desktop and make sure it's running
    echo Download from: https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)

echo Docker is running...
echo.

REM Add GOPATH\bin to PATH
for /f "delims=" %%i in ('go env GOPATH') do set GOPATH=%%i
set PATH=%GOPATH%\bin;%PATH%

REM Check if fyne-cross is installed
where fyne-cross >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo fyne-cross not found. Installing...
    go install github.com/fyne-io/fyne-cross@latest
    if %ERRORLEVEL% NEQ 0 (
        echo [ERROR] Failed to install fyne-cross
        pause
        exit /b 1
    )
    echo [SUCCESS] fyne-cross installed
    echo.
)

echo Using fyne-cross version:
fyne-cross version
echo.

cd launcher\gui

echo Building Windows natively (faster)...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w -H windowsgui" -o ..\..\EmuBuddyLauncher.exe .

if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] Windows build complete
) else (
    echo [ERROR] Windows build failed
)

echo.

echo [1/2] Building for Linux (amd64) using Docker...
fyne-cross linux -arch=amd64 -app-id com.emubuddy.launcher -name EmuBuddyLauncher

if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] Linux build complete
    copy /Y fyne-cross\bin\linux-amd64\gui ..\..\EmuBuddyLauncher-linux >nul 2>&1
) else (
    echo [ERROR] Linux build failed
)

echo.

echo [2/2] Building for macOS (darwin/amd64) using Docker...

REM Check for macOS SDK in emubuddy root (try multiple versions)
set "MACOS_SDK_PATH="
if exist "%~dp0MacOSX11.3.sdk" set "MACOS_SDK_PATH=%~dp0MacOSX11.3.sdk"
if exist "%~dp0MacOSX12.3.sdk" set "MACOS_SDK_PATH=%~dp0MacOSX12.3.sdk"
if exist "%~dp0MacOSX13.3.sdk" set "MACOS_SDK_PATH=%~dp0MacOSX13.3.sdk"
if exist "%~dp0MacOSX14.0.sdk" set "MACOS_SDK_PATH=%~dp0MacOSX14.0.sdk"
if exist "%~dp0MacOSX14.5.sdk" set "MACOS_SDK_PATH=%~dp0MacOSX14.5.sdk"

if not defined MACOS_SDK_PATH goto :skip_macos

echo Found macOS SDK at %MACOS_SDK_PATH%
fyne-cross darwin -arch=amd64 -app-id com.emubuddy.launcher -name EmuBuddyLauncher -macosx-sdk-path "%MACOS_SDK_PATH%"
if errorlevel 1 (
    echo [ERROR] macOS build failed
    goto :after_macos
)
echo [SUCCESS] macOS build complete
copy /Y fyne-cross\bin\darwin-amd64\gui ..\..\EmuBuddyLauncher-macos >nul 2>&1
goto :after_macos

:skip_macos
echo [SKIPPED] macOS SDK not found
echo.:
echo   RECOMMENDED: MacOSX 11.3 SDK (officially supported by fyne-cross)
echo   1. Download Command Line Tools for Xcode 12.5.1 from Apple Developer
echo   2. Extract SDK using: fyne-cross darwin-sdk-extract --xcode-path /path/to/file
echo   3. Place the extracted MacOSX11.3.sdk in: %~dp0
echo.
echo   ALTERNATIVE: Try newer SDKs (may have compatibility issues)
echo   - Download from: https://github.com/joseluisq/macosx-sdks/releases
echo   - Versions 12.x or 13.x are more likely to work than 14.x+mpatibility issues.
echo       Use SDK version 12.x - 14.x instead.

:after_macos

cd ..\..

echo.
echo =========================================
echo   Build Summary
echo =========================================
dir EmuBuddyLauncher.exe 2>nul
dir EmuBuddyLauncher-linux 2>nul
dir EmuBuddyLauncher-macos 2>nul
echo.
echo Done!
echo.
pause
