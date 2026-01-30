@echo off
setlocal enabledelayedexpansion

echo ==========================================
echo EmuBuddy Distribution Package Builder
echo ==========================================
echo.

:: Set version (can be passed as argument or defaulted)
set VERSION=%1
if "%VERSION%"=="" set VERSION=1.0.0

:: First, run the build scripts to create all binaries
echo Step 1: Building all binaries...
echo.

:: Build Installer (pure Go - works on all platforms)
call build_installer.bat
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Installer build failed
    exit /b 1
)

echo.
echo Step 2: Building Launcher using Docker (fyne-cross)...
echo.

:: Build Launcher (requires fyne-cross with Docker)
call build_launcher.bat
if %ERRORLEVEL% NEQ 0 (
    echo WARNING: Launcher build may have partially failed
    echo Continuing with available binaries...
)

echo.
echo Step 3: Creating distribution packages...
echo.

:: Create dist folder
set DIST_DIR=dist
if exist "%DIST_DIR%" rmdir /s /q "%DIST_DIR%"
mkdir "%DIST_DIR%"

:: Temp folders for each platform
set WIN_DIR=%DIST_DIR%\EmuBuddy-Windows-%VERSION%
set LINUX_DIR=%DIST_DIR%\EmuBuddy-Linux-%VERSION%
set MACOS_DIR=%DIST_DIR%\EmuBuddy-macOS-%VERSION%

:: ==========================================
:: Windows Distribution
:: ==========================================
echo [1/3] Creating Windows distribution...
mkdir "%WIN_DIR%"
mkdir "%WIN_DIR%\1g1rsets"
mkdir "%WIN_DIR%\roms"
mkdir "%WIN_DIR%\Emulators"

:: Copy Windows files
if exist "EmuBuddyLauncher.exe" (
    copy EmuBuddyLauncher.exe "%WIN_DIR%\"
    echo   + EmuBuddyLauncher.exe
) else (
    echo   ! WARNING: EmuBuddyLauncher.exe not found
)
if exist "EmuBuddySetup.exe" (
    copy EmuBuddySetup.exe "%WIN_DIR%\"
    echo   + EmuBuddySetup.exe
)
copy systems.json "%WIN_DIR%\"
copy README.md "%WIN_DIR%\"
xcopy /s /e /q 1g1rsets\*.json "%WIN_DIR%\1g1rsets\" >nul

:: Create Windows ZIP
echo   Creating ZIP...
cd "%DIST_DIR%"
powershell -Command "Compress-Archive -Path 'EmuBuddy-Windows-%VERSION%\*' -DestinationPath 'EmuBuddy-Windows-%VERSION%.zip' -Force"
cd ..

:: ==========================================
:: Linux Distribution
:: ==========================================
echo [2/3] Creating Linux distribution...
mkdir "%LINUX_DIR%"
mkdir "%LINUX_DIR%\1g1rsets"
mkdir "%LINUX_DIR%\roms"
mkdir "%LINUX_DIR%\Emulators"

:: Copy Linux files
if exist "EmuBuddyLauncher-linux" (
    copy EmuBuddyLauncher-linux "%LINUX_DIR%\"
    echo   + EmuBuddyLauncher-linux
) else (
    echo   ! WARNING: EmuBuddyLauncher-linux not found
)
if exist "EmuBuddySetup-linux" (
    copy EmuBuddySetup-linux "%LINUX_DIR%\"
    echo   + EmuBuddySetup-linux
)
copy systems.json "%LINUX_DIR%\"
copy README.md "%LINUX_DIR%\"
xcopy /s /e /q 1g1rsets\*.json "%LINUX_DIR%\1g1rsets\" >nul

:: Create run script for Linux
echo #!/bin/bash> "%LINUX_DIR%\run.sh"
echo chmod +x EmuBuddyLauncher-linux EmuBuddySetup-linux 2^>/dev/null>> "%LINUX_DIR%\run.sh"
echo ./EmuBuddyLauncher-linux>> "%LINUX_DIR%\run.sh"

:: Create Linux ZIP
echo   Creating ZIP...
cd "%DIST_DIR%"
powershell -Command "Compress-Archive -Path 'EmuBuddy-Linux-%VERSION%\*' -DestinationPath 'EmuBuddy-Linux-%VERSION%.zip' -Force"
cd ..

:: ==========================================
:: macOS Distribution
:: ==========================================
echo [3/3] Creating macOS distribution...
mkdir "%MACOS_DIR%"
mkdir "%MACOS_DIR%\1g1rsets"
mkdir "%MACOS_DIR%\roms"
mkdir "%MACOS_DIR%\Emulators"

:: Copy macOS files
if exist "EmuBuddyLauncher-macos" (
    copy EmuBuddyLauncher-macos "%MACOS_DIR%\"
    echo   + EmuBuddyLauncher-macos
) else (
    echo   ! WARNING: EmuBuddyLauncher-macos not found
)
if exist "EmuBuddySetup-macos" (
    copy EmuBuddySetup-macos "%MACOS_DIR%\"
    echo   + EmuBuddySetup-macos
)
copy systems.json "%MACOS_DIR%\"
copy README.md "%MACOS_DIR%\"
xcopy /s /e /q 1g1rsets\*.json "%MACOS_DIR%\1g1rsets\" >nul

:: Create run script for macOS
echo #!/bin/bash> "%MACOS_DIR%\run.sh"
echo chmod +x EmuBuddyLauncher-macos EmuBuddySetup-macos 2^>/dev/null>> "%MACOS_DIR%\run.sh"
echo ./EmuBuddyLauncher-macos>> "%MACOS_DIR%\run.sh"

:: Create macOS ZIP
echo   Creating ZIP...
cd "%DIST_DIR%"
powershell -Command "Compress-Archive -Path 'EmuBuddy-macOS-%VERSION%\*' -DestinationPath 'EmuBuddy-macOS-%VERSION%.zip' -Force"
cd ..

:: Cleanup temp folders
echo.
echo Cleaning up temporary folders...
rmdir /s /q "%WIN_DIR%" 2>nul
rmdir /s /q "%LINUX_DIR%" 2>nul
rmdir /s /q "%MACOS_DIR%" 2>nul

echo.
echo ==========================================
echo Distribution packages created in dist\
echo ==========================================
echo.
echo Files created:
for %%f in ("%DIST_DIR%\*.zip") do (
    echo   %%~nxf (%%~zf bytes)
)
echo.
echo Done!

endlocal
