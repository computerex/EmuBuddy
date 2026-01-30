# Verified Download URLs

All URLs have been tested and verified as of January 29, 2026.

## Emulators

### PCSX2 (PlayStation 2)
- **Version:** 2.2.0
- **URL:** https://github.com/PCSX2/pcsx2/releases/download/v2.2.0/pcsx2-v2.2.0-windows-x64-Qt.7z
- **Status:** ✓ Verified
- **Source:** [PCSX2 GitHub Releases](https://github.com/PCSX2/pcsx2/releases)

### PPSSPP (PlayStation Portable)
- **Version:** 1.19.3
- **URL:** https://www.ppsspp.org/files/1_19_3/ppsspp_win.zip
- **Status:** ✓ Verified
- **Source:** [PPSSPP Official Downloads](https://www.ppsspp.org/download/)

### Dolphin (GameCube / Wii)
- **Version:** 2512
- **URL:** https://dl.dolphin-emu.org/releases/2512/dolphin-2512-x64.7z
- **Status:** ✓ Verified
- **Source:** [Dolphin Stable Releases](https://dolphin-emu.org/download/)

### DeSmuME (Nintendo DS)
- **Version:** 0.9.13
- **URL:** https://github.com/TASEmulators/desmume/releases/download/release_0_9_13/desmume-0.9.13-win64.zip
- **Status:** ✓ Verified
- **Source:** [DeSmuME GitHub Releases](https://github.com/TASEmulators/desmume/releases)

### Azahar (Nintendo 3DS) - formerly Lime3DS
- **Version:** 2124.3
- **URL:** https://github.com/azahar-emu/azahar/releases/download/2124.3/azahar-2124.3-windows-msvc.zip
- **Status:** ✓ Verified
- **Note:** Lime3DS was renamed to Azahar in January 2026
- **Source:** [Azahar GitHub Releases](https://github.com/azahar-emu/azahar/releases)

### mGBA (Game Boy / Game Boy Advance)
- **Version:** 0.10.5
- **URL:** https://github.com/mgba-emu/mgba/releases/download/0.10.5/mGBA-0.10.5-win64.7z
- **Status:** ✓ Verified
- **Source:** [mGBA GitHub Releases](https://github.com/mgba-emu/mgba/releases)

### RetroArch (Multi-System)
- **Version:** 1.19.1
- **URL:** https://buildbot.libretro.com/stable/1.19.1/windows/x86_64/RetroArch.7z
- **Status:** ✓ Verified
- **Source:** [LibRetro Buildbot](https://buildbot.libretro.com/stable/)

### RetroArch Cores
- **Version:** 1.19.1
- **URL:** https://buildbot.libretro.com/stable/1.19.1/windows/x86_64/RetroArch_cores.7z
- **Status:** ✓ Verified
- **Cores included:** Nestopia (NES), Snes9x (SNES), Mupen64Plus (N64), Gambatte (GB/GBC), mGBA (GBA), Beetle PSX (PS1), Flycast (Dreamcast)

## Tools

### 7-Zip Standalone
- **URL:** https://www.7-zip.org/a/7zr.exe
- **Status:** ✓ Verified
- **Note:** Downloaded automatically by installer

## System Coverage

The installer provides emulators for **14 gaming systems**:
- NES, SNES, N64 (via RetroArch)
- Game Boy, Game Boy Color, Game Boy Advance (via RetroArch & mGBA)
- Nintendo DS (DeSmuME)
- Nintendo 3DS (Azahar)
- PlayStation 1 (via RetroArch)
- PlayStation 2 (PCSX2)
- PlayStation Portable (PPSSPP)
- GameCube, Wii (Dolphin)
- Dreamcast (via RetroArch)

## Total Download Size

- **Emulators:** ~350 MB
- **RetroArch Cores:** ~468 MB
- **7-Zip:** ~1 MB
- **Total:** ~820 MB

## Notes

1. All URLs point to stable, official releases
2. URLs are hardcoded in installer for reliability
3. Installer includes retry logic and error handling
4. Downloads resume if interrupted
5. Progress tracking for each emulator

## Updating URLs

To update URLs in the future:
1. Edit `installer/main.go`
2. Update the `emulators` array
3. Rebuild: `go build -ldflags="-s -w" -o ../EmuBuddySetup.exe main.go`

Last verified: January 29, 2026
