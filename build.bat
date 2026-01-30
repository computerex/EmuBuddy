@echo off
setlocal enabledelayedexpansion

:: ==========================================
:: EmuBuddy Build Script
:: ==========================================
:: Usage: build.bat [command] [version]
::
:: Commands:
::   all       - Build everything for all platforms (default)
::   installer - Build installer only
::   launcher  - Build launcher only  
::   win       - Build Windows launcher only (fast iteration)
::   win-setup - Build Windows installer only (fast iteration)
::   dist      - Create distribution ZIPs
::   clean     - Remove build artifacts
::
:: Examples:
::   build.bat              (builds all)
::   build.bat dist 1.0.0   (creates distribution ZIPs)
::   build.bat clean        (removes build artifacts)
:: ==========================================

set COMMAND=%1
set VERSION=%2
if "%COMMAND%"=="" set COMMAND=all
if "%VERSION%"=="" set VERSION=1.0.0

echo.
echo ==========================================
echo   EmuBuddy Build System
echo ==========================================
echo   Command: %COMMAND%
echo   Version: %VERSION%
echo ==========================================
echo.

if "%COMMAND%"=="clean" goto :clean
if "%COMMAND%"=="installer" goto :installer
if "%COMMAND%"=="launcher" goto :launcher
if "%COMMAND%"=="win" goto :win
if "%COMMAND%"=="win-setup" goto :win_setup
if "%COMMAND%"=="dist" goto :dist
if "%COMMAND%"=="all" goto :all

echo Unknown command: %COMMAND%
echo Use: build.bat [all^|installer^|launcher^|win^|win-setup^|dist^|clean]
exit /b 1

:: ==========================================
:: WIN (Windows launcher only - fast)
:: ==========================================
:win
echo Building Windows Launcher...
cd launcher\gui
go build -ldflags="-s -w -H windowsgui" -o ..\..\EmuBuddyLauncher.exe .
if %ERRORLEVEL% EQU 0 (
    echo [OK] EmuBuddyLauncher.exe
) else (
    echo [FAILED]
)
cd ..\..
goto :end

:: ==========================================
:: WIN-SETUP (Windows installer only - fast)
:: ==========================================
:win_setup
echo Building Windows Installer...
cd installer
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o ..\EmuBuddySetup.exe main.go
if %ERRORLEVEL% EQU 0 (
    echo [OK] EmuBuddySetup.exe
) else (
    echo [FAILED]
)
cd ..
goto :end

:: ==========================================
:: CLEAN
:: ==========================================
:clean
echo Cleaning build artifacts...
del /q EmuBuddyLauncher.exe 2>nul
del /q EmuBuddyLauncher-linux 2>nul
del /q EmuBuddyLauncher-macos 2>nul
del /q EmuBuddySetup.exe 2>nul
del /q EmuBuddySetup-linux 2>nul
del /q EmuBuddySetup-macos 2>nul
del /q launcher_debug.log 2>nul
del /q launch_debug.txt 2>nul
del /q *.exe~ 2>nul
rmdir /s /q dist 2>nul
rmdir /s /q launcher\gui\fyne-cross 2>nul
echo Done.
goto :end

:: ==========================================
:: INSTALLER (all platforms)
:: ==========================================
:installer
echo Building Installers...
cd installer

set CGO_ENABLED=0

echo   [1/3] Windows...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o ..\EmuBuddySetup.exe main.go
if %ERRORLEVEL% NEQ 0 (echo   [FAILED] Windows & goto :installer_done)
echo   [OK] Windows

echo   [2/3] Linux...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o ..\EmuBuddySetup-linux main.go
if %ERRORLEVEL% NEQ 0 (echo   [FAILED] Linux & goto :installer_done)
echo   [OK] Linux

echo   [3/3] macOS...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o ..\EmuBuddySetup-macos main.go
if %ERRORLEVEL% NEQ 0 (echo   [FAILED] macOS & goto :installer_done)
echo   [OK] macOS

:installer_done
cd ..
goto :end

:: ==========================================
:: LAUNCHER (all platforms via Docker)
:: ==========================================
:launcher
echo Building Launchers...

:: Check Docker
docker info >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Docker is not running. Start Docker Desktop first.
    goto :end
)

cd launcher\gui

:: Windows native build
echo   [1/3] Windows (native)...
set GOOS=
set GOARCH=
set CGO_ENABLED=
go build -ldflags="-s -w -H windowsgui" -o ..\..\EmuBuddyLauncher.exe .
if %ERRORLEVEL% NEQ 0 (echo   [FAILED] Windows & goto :launcher_done)
echo   [OK] Windows

:: Linux via fyne-cross
echo   [2/3] Linux (Docker)...
fyne-cross linux -arch=amd64 -app-id=com.emubuddy.launcher >nul 2>&1
if exist "fyne-cross\bin\linux-amd64\gui" (
    copy /y "fyne-cross\bin\linux-amd64\gui" "..\..\EmuBuddyLauncher-linux" >nul
    echo   [OK] Linux
) else (
    echo   [FAILED] Linux
)

:: macOS via fyne-cross
echo   [3/3] macOS (Docker)...
if exist "..\..\MacOSX11.3.sdk" (
    fyne-cross darwin -arch=amd64 -app-id=com.emubuddy.launcher -macosx-sdk-path="..\..\MacOSX11.3.sdk" >nul 2>&1
) else (
    fyne-cross darwin -arch=amd64 -app-id=com.emubuddy.launcher >nul 2>&1
)
if exist "fyne-cross\bin\darwin-amd64\gui" (
    copy /y "fyne-cross\bin\darwin-amd64\gui" "..\..\EmuBuddyLauncher-macos" >nul
    echo   [OK] macOS
) else (
    echo   [FAILED] macOS
)

:launcher_done
cd ..\..
goto :end

:: ==========================================
:: ALL (installer + launcher)
:: ==========================================
:all
call :installer
call :launcher
echo.
echo Build Summary:
dir /b EmuBuddy* 2>nul
goto :end

:: ==========================================
:: DIST (create distribution ZIPs)
:: ==========================================
:dist
:: Build everything first
call :installer
call :launcher

echo.
echo Creating distribution packages...

if exist dist rmdir /s /q dist
mkdir dist

:: Windows
echo   [1/3] Windows ZIP...
mkdir dist\EmuBuddy-Windows-%VERSION%
mkdir dist\EmuBuddy-Windows-%VERSION%\1g1rsets
mkdir dist\EmuBuddy-Windows-%VERSION%\roms
mkdir dist\EmuBuddy-Windows-%VERSION%\Emulators
if exist EmuBuddyLauncher.exe copy /y EmuBuddyLauncher.exe dist\EmuBuddy-Windows-%VERSION%\ >nul
if exist EmuBuddySetup.exe copy /y EmuBuddySetup.exe dist\EmuBuddy-Windows-%VERSION%\ >nul
copy /y systems.json dist\EmuBuddy-Windows-%VERSION%\ >nul
copy /y README.md dist\EmuBuddy-Windows-%VERSION%\ >nul
xcopy /s /e /q 1g1rsets\*.json dist\EmuBuddy-Windows-%VERSION%\1g1rsets\ >nul
cd dist
powershell -Command "Compress-Archive -Path 'EmuBuddy-Windows-%VERSION%\*' -DestinationPath 'EmuBuddy-Windows-%VERSION%.zip' -Force"
rmdir /s /q EmuBuddy-Windows-%VERSION%
cd ..

:: Linux
echo   [2/3] Linux ZIP...
mkdir dist\EmuBuddy-Linux-%VERSION%
mkdir dist\EmuBuddy-Linux-%VERSION%\1g1rsets
mkdir dist\EmuBuddy-Linux-%VERSION%\roms
mkdir dist\EmuBuddy-Linux-%VERSION%\Emulators
if exist EmuBuddyLauncher-linux copy /y EmuBuddyLauncher-linux dist\EmuBuddy-Linux-%VERSION%\ >nul
if exist EmuBuddySetup-linux copy /y EmuBuddySetup-linux dist\EmuBuddy-Linux-%VERSION%\ >nul
copy /y systems.json dist\EmuBuddy-Linux-%VERSION%\ >nul
copy /y README.md dist\EmuBuddy-Linux-%VERSION%\ >nul
xcopy /s /e /q 1g1rsets\*.json dist\EmuBuddy-Linux-%VERSION%\1g1rsets\ >nul
echo #!/bin/bash> dist\EmuBuddy-Linux-%VERSION%\run.sh
echo chmod +x EmuBuddyLauncher-linux EmuBuddySetup-linux 2^>/dev/null>> dist\EmuBuddy-Linux-%VERSION%\run.sh
echo ./EmuBuddyLauncher-linux>> dist\EmuBuddy-Linux-%VERSION%\run.sh
cd dist
powershell -Command "Compress-Archive -Path 'EmuBuddy-Linux-%VERSION%\*' -DestinationPath 'EmuBuddy-Linux-%VERSION%.zip' -Force"
rmdir /s /q EmuBuddy-Linux-%VERSION%
cd ..

:: macOS
echo   [3/3] macOS ZIP...
mkdir dist\EmuBuddy-macOS-%VERSION%
mkdir dist\EmuBuddy-macOS-%VERSION%\1g1rsets
mkdir dist\EmuBuddy-macOS-%VERSION%\roms
mkdir dist\EmuBuddy-macOS-%VERSION%\Emulators
if exist EmuBuddyLauncher-macos copy /y EmuBuddyLauncher-macos dist\EmuBuddy-macOS-%VERSION%\ >nul
if exist EmuBuddySetup-macos copy /y EmuBuddySetup-macos dist\EmuBuddy-macOS-%VERSION%\ >nul
copy /y systems.json dist\EmuBuddy-macOS-%VERSION%\ >nul
copy /y README.md dist\EmuBuddy-macOS-%VERSION%\ >nul
xcopy /s /e /q 1g1rsets\*.json dist\EmuBuddy-macOS-%VERSION%\1g1rsets\ >nul
echo #!/bin/bash> dist\EmuBuddy-macOS-%VERSION%\run.sh
echo chmod +x EmuBuddyLauncher-macos EmuBuddySetup-macos 2^>/dev/null>> dist\EmuBuddy-macOS-%VERSION%\run.sh
echo ./EmuBuddyLauncher-macos>> dist\EmuBuddy-macOS-%VERSION%\run.sh
cd dist
powershell -Command "Compress-Archive -Path 'EmuBuddy-macOS-%VERSION%\*' -DestinationPath 'EmuBuddy-macOS-%VERSION%.zip' -Force"
rmdir /s /q EmuBuddy-macOS-%VERSION%
cd ..

echo.
echo Distribution packages created:
for %%f in (dist\*.zip) do echo   %%~nxf

goto :end

:end
echo.
endlocal
