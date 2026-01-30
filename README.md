# EmuBuddy - Complete Emulation Station

Cross-platform emulation frontend with automated ROM downloading and launching.

## Features

- **Multi-System Support** - 14 gaming systems (NES to PS2)
- **Automated Setup** - Download and extract emulators automatically
- **Web-based Frontend** - Modern browser-based UI (no build required)
- **Native GUI Option** - Desktop application (requires 64-bit GCC)
- **Smart Downloads** - Auto-download ROMs from 1g1r curated sets
- **One-Click Launch** - Automatic emulator detection and launching
- **Search & Browse** - Filter 20,000+ games instantly
- **ROM Management** - Track downloaded vs available ROMs
- **Portable** - Fully portable, no system changes required

## Quick Start

### 1. Install Emulators (one-time setup)

**Windows:**
```batch
EmuBuddySetup.exe
```

Simply double-click `EmuBuddySetup.exe` and it will automatically:
- Download all emulators (~350 MB)
- Download RetroArch cores (~468 MB)
- Extract everything to the correct locations
- Show progress for each step

**Total download: ~820 MB** (one-time)

### 2. Choose Your Frontend

**Option A: Web Frontend (Recommended)**
```batch
start_launcher.bat          # Windows
./start_launcher.sh         # Linux/macOS
```
- Opens browser at http://localhost:8080/static/
- No compilation required
- Works immediately

**Option B: Native GUI**
```batch
start_gui.bat               # Windows
./start_gui.sh              # Linux/macOS
```
- Native desktop application
- Lower resource usage
- See `launcher/gui/BUILD.md` for build requirements (64-bit GCC)

## Supported Systems

| System | Emulator | Cores/Support |
|--------|----------|---------------|
| NES | RetroArch (Nestopia) | . |
| SNES | RetroArch (Snes9x) | . |
| N64 | RetroArch (Mupen64Plus) | . |
| Game Boy | RetroArch (Gambatte) | . |
| Game Boy Color | RetroArch (Gambatte) | . |
| Game Boy Advance | RetroArch (mGBA) | . |
| Nintendo DS | DeSmuME | . |
| Nintendo 3DS | Lime3DS | . |
| GameCube | Dolphin | . |
| Wii | Dolphin | . |
| PlayStation 1 | RetroArch (Beetle PSX) | . |
| PlayStation 2 | PCSX2 | . |
| PSP | PPSSPP | . |
| Dreamcast | RetroArch (Flycast) | . |

**Total: 14 systems, 20,000+ curated games**

## Project Structure

```
emubuddy/
├── Emulators/                  # Downloaded emulators
│   ├── RetroArch/             # Multi-system emulator
│   ├── PCSX2/                 # PS2
│   ├── PPSSPP/                # PSP
│   ├── Dolphin/               # GameCube/Wii
│   ├── DeSmuME/               # Nintendo DS
│   └── Lime3DS/               # Nintendo 3DS
│
├── roms/                       # Downloaded ROMs (auto-created)
│   ├── nes/
│   ├── snes/
│   └── ...
│
├── 1g1rsets/                   # ROM databases (60 JSON files)
│   ├── games_1g1r_english_nes.json
│   └── ...
│
├── Tools/
│   ├── romget/                # ROM downloader utility
│   └── 7zip/                  # Portable 7-Zip (auto-installed)
│
├── launcher/
│   ├── frontend/              # Web-based launcher (Go)
│   └── gui/                   # Native GUI (Go + Fyne)
│
├── installer/                  # Installer source code
│   └── main.go                # Single-exe installer
│
├── EmuBuddySetup.exe          # One-click installer
├── start_launcher.bat          # Quick start (Windows - Web)
├── start_gui.bat               # Quick start (Windows - GUI)
└── README.md                   # This file
```

## Frontend Options

### Web Frontend (Recommended)

**Best for:** Most users, zero build friction

```batch
start_launcher.bat          # Windows
./start_launcher.sh         # Linux/macOS
```

**Features:**
- Opens in browser (http://localhost:8080/static/)
- No compilation required
- Works immediately
- Modern web interface
- Easy to update

### Native GUI

**Best for:** Desktop app lovers, minimal resource usage

```batch
start_gui.bat               # Windows
./start_gui.sh              # Linux/macOS
```

**Features:**
- Native OS look and feel
- 50% less memory than web
- Truly offline operation
- Faster performance
- **Requires:** 64-bit GCC for building
- See `launcher/gui/BUILD.md` for instructions

### Comparison

| Feature | Web Frontend | Native GUI |
|---------|-------------|-----------|
| Setup | No build needed | Requires build tools |
| Memory | ~100 MB | ~30-50 MB |
| Performance | Fast | Faster |
| Look | Custom web theme | Native OS theme |
| Updates | Refresh browser | Rebuild required |

Both frontends offer identical functionality: browse 20,000+ games, search/filter, auto-download ROMs, one-click launching.

## Component Documentation

- **Installer:** `installer/main.go` - Single-exe installer source
- **Web Frontend:** `launcher/frontend/README.md`
- **Native GUI:** `launcher/gui/README.md`
- **GUI Build Instructions:** `launcher/gui/BUILD.md`
- **ROM Downloader:** `Tools/romget/README.md`

## Configuration

### Add New System

1. Place 1g1r JSON in `1g1rsets/games_1g1r_english_{system}.json`
2. Update system config in `launcher/frontend/main.go` or `launcher/gui/main.go`
3. Rebuild the frontend

### Change Emulator

Edit system config in the frontend's main.go file to use a different emulator path or arguments.

## Troubleshooting

### Emulators Not Downloading
- Check internet connection
- Run `EmuBuddySetup.exe` again (it will resume where it left off)
- Ensure you have ~1 GB of free disk space

### ROM Download Fails
- Verify `tools/romget/romget.exe` exists
- Check myrient.erista.me is accessible

### Game Won't Launch
- Verify ROM exists in `roms/{system}/` directory
- Check emulator path in system config
- Try launching emulator manually first

### Port 8080 In Use
Change port in `launcher/frontend/main.go`

## Building from Source

### Installer
```bash
cd installer
go build -ldflags="-s -w" -o ../EmuBuddySetup.exe main.go
```

### Web Frontend
```bash
cd launcher/frontend
go build -o emubuddy-frontend.exe
```

### Native GUI
```bash
cd launcher/gui
go build -o emubuddy-gui.exe
```

### ROM Downloader
```bash
cd Tools/romget
go build -o romget.exe
```

## System Requirements

- **OS:** Windows 10/11, Linux, macOS
- **RAM:** 4 GB minimum, 8 GB recommended
- **Storage:** 2 GB for emulators + space for ROMs
- **Internet:** Required for initial setup and ROM downloads

## Total Download Size

**Automatic installer downloads:**
- **Emulators:** ~350 MB (7 standalone + RetroArch)
- **RetroArch Cores:** ~468 MB (13 cores)
- **7-Zip:** ~1 MB (downloaded automatically)
- **Total Download:** ~820 MB
- **Total Disk Space:** ~1.8 GB (includes extracted files)

## Gaming Era Coverage

| Era | Years | Systems |
|-----|-------|---------|
| 8-bit | 1977-1990 | NES, Master System, Atari 2600 |
| 16-bit | 1988-1996 | SNES, Genesis, Game Gear |
| 32/64-bit | 1993-2002 | PS1, N64, GBA |
| 128-bit | 1998-2013 | PS2, GameCube, Wii, PSP, DS |
| Modern | 2012+ | 3DS, Wii U |
| Arcade | 1978-1995 | MAME, FBNeo |

**Total Coverage: 40+ years of gaming history (1977-2017)**

## License

Educational/Personal use. Respect ROM copyright laws in your jurisdiction.

## Credits

- **RetroArch** - https://www.retroarch.com/
- **PCSX2** - https://pcsx2.net/
- **PPSSPP** - https://www.ppsspp.org/
- **Dolphin** - https://dolphin-emu.org/
- **mGBA** - https://mgba.io/
- **DeSmuME** - https://desmume.org/
- **Lime3DS** - https://github.com/Lime3DS/Lime3DS
- **Myrient** - https://myrient.erista.me/ (ROM hosting)
- **No-Intro** - https://no-intro.org/ (ROM preservation)

## Disclaimer

EmuBuddy is a frontend/launcher tool. Users are responsible for obtaining ROMs legally. Only download ROMs for games you own physically.
