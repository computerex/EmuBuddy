# EmuBuddy - Complete Emulation Station

Cross-platform emulation frontend with automated emulator setup and game launching.

## Features

- **Multi-System Support** - 14 gaming systems (NES to PS2)
- **Automated Setup** - Downloads and installs all emulators automatically
- **Native GUI** - Fast desktop application with controller support
- **Smart Downloads** - Download ROMs from curated 1G1R sets
- **One-Click Launch** - Automatic emulator detection and game launching
- **Search & Filter** - Browse 20,000+ games with instant search
- **Favorites** - Mark and filter your favorite games
- **Controller Navigation** - Full gamepad support for couch gaming
- **Portable** - Fully portable, no system installation required

## Quick Start

### Windows
1. Double-click `EmuBuddySetup.exe` to install emulators
2. Double-click `EmuBuddyLauncher.exe` to start

### macOS
1. Double-click `Run Setup.command` to install emulators
2. Double-click `Start EmuBuddy.command` to launch

### Linux
1. Run `./run-setup.sh` to install emulators
2. Run `./start-emubuddy.sh` to launch

**Note:** The launcher will automatically run setup if emulators aren't installed.

## Supported Systems

| System | Emulator | Notes |
|--------|----------|-------|
| NES | RetroArch (Nestopia) | |
| SNES | RetroArch (Snes9x) | |
| N64 | RetroArch (Mupen64Plus) | |
| Game Boy | RetroArch (Gambatte) | |
| Game Boy Color | RetroArch (Gambatte) | |
| Game Boy Advance | RetroArch (mGBA) | |
| Nintendo DS | melonDS | |
| Nintendo 3DS | Azahar | |
| GameCube | Dolphin | |
| Wii | Dolphin | |
| PlayStation 1 | RetroArch (Beetle PSX) | |
| PlayStation 2 | PCSX2 | |
| PSP | PPSSPP | |
| Dreamcast | RetroArch (Flycast) | |

**Total: 14 systems, 20,000+ curated games**

## Controls

### Keyboard
- **Arrow Keys** - Navigate systems and games
- **Enter** - Launch selected game
- **Tab** - Switch between system and game lists
- **Escape** - Cancel / Close dialogs
- **Type** - Search games

### Controller
- **D-Pad / Left Stick** - Navigate
- **A Button** - Select / Launch
- **B Button** - Back / Cancel
- **Right Stick** - Scroll game list

## Project Structure

```
EmuBuddy/
├── EmuBuddyLauncher.exe    # Main application
├── EmuBuddySetup.exe       # Emulator installer
├── systems.json            # System configuration
├── README.md               # This file
│
├── Emulators/              # Installed emulators
│   ├── RetroArch/          # Multi-system (NES, SNES, N64, GB, GBA, PS1, DC)
│   ├── PCSX2/              # PlayStation 2
│   ├── PPSSPP/             # PSP
│   ├── Dolphin/            # GameCube & Wii
│   ├── melonDS/            # Nintendo DS
│   └── Azahar/             # Nintendo 3DS
│
├── roms/                   # Your game files (organized by system)
│   ├── nes/
│   ├── snes/
│   ├── psx/
│   └── ...
│
└── 1g1rsets/               # ROM databases (curated game lists)
```

## Building from Source

### Requirements
- Go 1.21+
- Docker (for cross-platform builds)

### Build Commands
```batch
build.bat win        # Windows launcher only (fast)
build.bat win-setup  # Windows installer only
build.bat launcher   # All platforms (requires Docker)
build.bat installer  # All platform installers
build.bat dist 1.0   # Create distribution ZIPs
build.bat clean      # Remove build artifacts
```

## License

This software is provided for personal use. Users are responsible for ensuring they have the legal right to use any games with this software.

## Credits

- RetroArch - https://retroarch.com
- Dolphin - https://dolphin-emu.org
- PCSX2 - https://pcsx2.net
- PPSSPP - https://ppsspp.org
- melonDS - https://melonds.kuribo64.net
- Azahar - https://azahar-emu.org
