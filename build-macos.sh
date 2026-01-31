#!/bin/bash
# ==========================================
# EmuBuddy macOS Native Build Script
# ==========================================
# Run this script ON A MAC to build working macOS binaries.
# Cross-compiled macOS binaries from Windows/Linux often crash.
#
# Usage: ./build-macos.sh
# ==========================================

set -e

echo ""
echo "=========================================="
echo "  EmuBuddy macOS Native Build"
echo "=========================================="
echo ""

# Build Launcher
echo "[1/2] Building macOS Launcher..."
cd launcher/gui
go build -ldflags="-s -w" -o ../../EmuBuddyLauncher-macos .
echo "  [OK] EmuBuddyLauncher-macos"
cd ../..

# Build Installer
echo "[2/2] Building macOS Installer..."
cd installer
CGO_ENABLED=0 go build -ldflags="-s -w" -o ../EmuBuddySetup-macos main.go
echo "  [OK] EmuBuddySetup-macos"
cd ..

echo ""
echo "=========================================="
echo "  Build Complete!"
echo "=========================================="
echo ""
ls -la EmuBuddy*-macos 2>/dev/null || echo "No macOS binaries found"
echo ""
