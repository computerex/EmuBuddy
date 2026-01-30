#!/bin/bash
# Build EmuBuddy Installer for all platforms

echo "========================================="
echo "  Building EmuBuddy Cross-Platform"
echo "========================================="
echo ""

cd installer

# Windows
echo "[1/3] Building for Windows..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ../EmuBuddySetup.exe main.go
if [ $? -eq 0 ]; then
    echo "✓ Windows build complete: EmuBuddySetup.exe"
else
    echo "✗ Windows build failed"
fi

# Linux
echo ""
echo "[2/3] Building for Linux..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../EmuBuddySetup-linux main.go
if [ $? -eq 0 ]; then
    echo "✓ Linux build complete: EmuBuddySetup-linux"
    chmod +x ../EmuBuddySetup-linux
else
    echo "✗ Linux build failed"
fi

# macOS
echo ""
echo "[3/3] Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ../EmuBuddySetup-macos main.go
if [ $? -eq 0 ]; then
    echo "✓ macOS build complete: EmuBuddySetup-macos"
    chmod +x ../EmuBuddySetup-macos
else
    echo "✗ macOS build failed"
fi

cd ..

echo ""
echo "========================================="
echo "  Build Summary"
echo "========================================="
ls -lh EmuBuddySetup* 2>/dev/null || echo "No builds found"
echo ""
echo "Done!"
