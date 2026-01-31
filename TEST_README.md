# EmuBuddy Emulator Test Suite

## Overview

The `test-emulators.sh` script is an automated testing tool that verifies all emulators work correctly on your system.

## What It Does

1. **Downloads Test ROMs** - Downloads small (< 1MB) homebrew/open-source test ROMs for each system
2. **Launches Emulators** - Automatically launches each emulator with the test ROM
3. **Interactive Testing** - Prompts you to confirm if each emulator worked
4. **Generates Report** - Creates a detailed test report showing pass/fail status

## Test ROMs Used

All test ROMs are **legal, open-source homebrew** projects:

- **240p Test Suite** (NES, SNES, Genesis, N64) - Display calibration and test patterns
- **Adjustris** (Game Boy) - Simple Tetris-like puzzle game
- **GoodBoyAdvance** (GBA) - Demo ROM for testing

## Usage

### Windows

Run the test suite using either method:

**Double-click:**
```
test-emulators.bat
```

**PowerShell:**
```powershell
.\test-emulators.ps1
```

**With custom installation directory:**
```powershell
.\test-emulators.ps1 -InstallDir "C:\Path\To\EmuBuddy"
```

### Linux/macOS

```bash
cd /path/to/emubuddy
./test-emulators.sh
```

### What to Expect

For each system:
1. Script downloads a test ROM (if not already downloaded)
2. Launches the emulator with the ROM
3. Emulator runs for 5 seconds, then auto-closes
4. You are prompted: "Did the emulator launch and display the game? (y/n)"
5. Answer `y` if it worked, `n` if it didn't
6. If you answered `n`, you can optionally describe the issue

### After Testing

The script generates:
- **Console output** - Colored summary of results
- **emulator-test-report.txt** - Detailed text report

## Report Format

The script generates a detailed test report:

```
EmuBuddy Emulator Test Report
Generated: 2026-01-31 03:00:00
Installation: C:\Users\User\EmuBuddy
Platform: Windows
PowerShell Version: 5.1.x.x

=========================================

nes                  : PASS   : Emulator launched successfully
snes                 : PASS   : Emulator launched successfully
n64                  : FAIL   : Black screen, no display
gb                   : PASS   : Emulator launched successfully
gba                  : SKIP   : Core not installed

=========================================
Summary:
  Passed:  3
  Failed:  1
  Skipped: 1
  Total:   5
=========================================
```

## Tested Systems

Tests **ALL 27 systems**:

**Nintendo:**
- NES, SNES, N64
- Game Boy, Game Boy Color, Game Boy Advance
- Nintendo DS, 3DS
- GameCube, Wii, Wii U

**Sega:**
- Genesis/Mega Drive, Master System, Game Gear
- Dreamcast

**Sony:**
- PlayStation 1, PlayStation 2
- PSP

**Atari:**
- Atari 2600, Atari 7800
- Lynx

**Other:**
- TurboGrafx-16/PC Engine
- Virtual Boy
- Neo Geo Pocket Color
- WonderSwan, WonderSwan Color
- ColecoVision
- Intellivision

## Adding More Systems

To add more test ROMs, edit the `TEST_ROMS` array in the script:

```bash
TEST_ROMS=(
    "system_id|rom_name|download_url|file_extension"
)
```

## Notes

- Test ROMs are downloaded to `test-roms/` directory (gitignored)
- Each ROM is only downloaded once (cached for subsequent runs)
- The script auto-detects your EmuBuddy installation directory
- Works on Linux, macOS, and Windows (with Git Bash)

## Troubleshooting

**"Could not find EmuBuddy installation directory"**
- Make sure EmuBuddy is installed in your home directory
- Or edit the `install_dir` detection in the script

**"Emulator not found"**
- Run the EmuBuddySetup first to install emulators
- Check that the emulator exists in the Emulators/ folder

**"Core not found"**
- Some RetroArch cores may not be installed
- Run EmuBuddySetup and reinstall RetroArch cores

## Windows-Specific Notes

### Requirements
- Windows 10/11 with PowerShell 5.1 or later
- EmuBuddyLauncher.exe must be in the installation directory

### Usage

Run the test suite using either method:

**Double-click:**
```
test-emulators.bat
```

**PowerShell:**
```powershell
.\test-emulators.ps1
```

**With custom installation directory:**
```powershell
.\test-emulators.ps1 -InstallDir "C:\Path\To\EmuBuddy"
```

### Permission Issues
If you get "execution policy" errors:
1. Right-click on `test-emulators.bat` and select "Run as Administrator", OR
2. Run PowerShell as Administrator and execute:
   ```powershell
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   ```

### Test Results Location
- Report is saved to: `emulator-test-report.txt` in the EmuBuddy directory
- Test ROMs are cached in: `test-roms/` subdirectory

### Troubleshooting Launcher Issues
If `EmuBuddyLauncher.exe` is not found, specify the path:
```powershell
.\test-emulators.ps1 -InstallDir "C:\Path\To\EmuBuddy"
```

### Tested Windows Emulators
All 27 systems are tested on Windows:

| Emulator | Windows Executable |
|----------|-------------------|
| RetroArch | `RetroArch\RetroArch-Win64\retroarch.exe` |
| PCSX2 | `PCSX2\pcsx2-qt.exe` |
| PPSSPP | `PPSSPP\PPSSPPWindows64.exe` |
| Dolphin | `Dolphin\Dolphin-x64\Dolphin.exe` |
| melonDS | `melonDS\melonDS.exe` |
| Azahar | `Azahar\azahar.exe` |
| Cemu | `Cemu\Cemu.exe` |
