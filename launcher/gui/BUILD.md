# Building EmuBuddy GUI

## Prerequisites

### Windows

The GUI requires a 64-bit C compiler for CGO (OpenGL bindings).

**Option 1: TDM-GCC (Recommended)**
1. Download: https://jmeubank.github.io/tdm-gcc/download/
2. Install TDM64 (64-bit version)
3. Add to PATH: `C:\TDM-GCC-64\bin`

**Option 2: MSYS2**
```bash
# Install MSYS2 from https://www.msys2.org/
# Then in MSYS2 terminal:
pacman -S mingw-w64-x86_64-gcc
```

**Option 3: Use Pre-Built Fyne Binaries**
```bash
# Install fyne command
go install fyne.io/fyne/v2/cmd/fyne@latest

# Build using fyne command (handles CGO automatically)
fyne build -o emubuddy-gui.exe
```

### Linux

```bash
# Ubuntu/Debian
sudo apt-get install gcc libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install gcc libXcursor-devel libXrandr-devel mesa-libGL-devel
```

### macOS

```bash
xcode-select --install
```

## Build Instructions

### Method 1: Using Go Build (Requires CGO)

```bash
cd launcher/gui
go build -o emubuddy-gui.exe
```

### Method 2: Using Fyne Command (Easier)

```bash
# Install fyne command once
go install fyne.io/fyne/v2/cmd/fyne@latest

# Build
cd launcher/gui
fyne build -o emubuddy-gui.exe
```

### Method 3: Quick Start Scripts

**Windows:**
```batch
start_gui.bat
```

**Linux/macOS:**
```bash
./start_gui.sh
```

## If Build Fails

### "64-bit mode not compiled in"

Your MinGW/GCC is 32-bit only. Options:
1. Install TDM-GCC 64-bit (see Prerequisites above)
2. Use WSL2 with Linux build tools
3. Use the web frontend instead: `start_launcher.bat`

### "cannot find package"

```bash
go mod tidy
go get fyne.io/fyne/v2@latest
```

### "undefined: fyne"

```bash
go get fyne.io/fyne/v2@latest
go mod tidy
```

## Alternative: Use Web Frontend

If you can't build the GUI, use the web-based frontend:

```batch
# Windows
start_launcher.bat

# Linux/macOS
./start_launcher.sh
```

Then open: http://localhost:8080/static/

The web frontend has the same features and requires no compilation.

## Cross-Compilation

### Build for Windows from Linux

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o emubuddy-gui.exe
```

### Build for Linux from Windows

```bash
# Requires WSL2 or Docker
docker run --rm -v "%cd%":/app -w /app golang:1.19 go build -o emubuddy-gui
```

### Build for macOS

```bash
# Must be done on macOS
GOOS=darwin GOARCH=amd64 go build -o emubuddy-gui
```

## Distribution

### Windows Installer

```bash
# Using fyne command
fyne package -os windows -name "EmuBuddy" -icon icon.png
```

### Portable ZIP

```bash
# Build
go build -o emubuddy-gui.exe

# Create package
7z a emubuddy-gui-portable.zip emubuddy-gui.exe README.md
```

## Troubleshooting

### GUI doesn't start

Check console output:
```bash
# Windows
emubuddy-gui.exe

# Linux/macOS
./emubuddy-gui
```

### "Cannot find Emulators directory"

Run from emubuddy root directory:
```bash
cd C:\projects\emubuddy
launcher\gui\emubuddy-gui.exe
```

### OpenGL errors on Linux

```bash
# Install Mesa drivers
sudo apt-get install mesa-utils libgl1-mesa-dri
```
