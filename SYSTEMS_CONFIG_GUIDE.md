# Systems Configuration Guide

The EmuBuddy Launcher now uses a `systems.json` configuration file to define all supported systems. This allows you to add or modify systems without recompiling the launcher.

## Configuration File Location

The `systems.json` file must be placed in the root EmuBuddy directory (same folder as `EmuBuddyLauncher.exe`).

## Configuration Structure

```json
{
  "systems": [
    {
      "id": "systemid",
      "name": "Display Name",
      "dir": "roms_subdirectory",
      "romJsonFile": "systemid.json",
      "emulator": {
        "path": "Emulators/Path/To/Emulator.exe",
        "args": ["-L", "cores/core_name.dll"],
        "name": "Emulator Display Name"
      },
      "standaloneEmulator": {
        "path": "Emulators/Path/To/Standalone.exe",
        "args": [],
        "name": "Standalone Emulator Name"
      },
      "fileExtensions": [".ext1", ".ext2"],
      "needsExtract": false
    }
  ]
}
```

## Field Descriptions

### Required Fields

- **id**: Unique identifier for the system (lowercase, no spaces)
- **name**: Display name shown in the launcher
- **dir**: Subdirectory under `roms/` where ROMs are stored
- **romJsonFile**: Name of the JSON file in `1g1rsets/` containing ROM list
- **emulator**: Primary emulator configuration
  - **path**: Relative path from EmuBuddy root to emulator executable
  - **args**: Command-line arguments (use for RetroArch cores: `["-L", "cores/corename.dll"]`)
  - **name**: Display name for this emulator option
- **fileExtensions**: Array of supported file extensions (include the dot: `.zip`, `.iso`)
- **needsExtract**: Boolean - whether to extract ZIP files before launching

### Optional Fields

- **standaloneEmulator**: Alternative emulator configuration (set to `null` if not available)
  - Same structure as `emulator` field
  - When configured, launcher will ask user to choose between primary and standalone

## Examples

### System with RetroArch Only

```json
{
  "id": "nes",
  "name": "Nintendo Entertainment System",
  "dir": "nes",
  "romJsonFile": "nes.json",
  "emulator": {
    "path": "Emulators/RetroArch/RetroArch-Win64/retroarch.exe",
    "args": ["-L", "cores/nestopia_libretro.dll"],
    "name": "RetroArch"
  },
  "standaloneEmulator": null,
  "fileExtensions": [".nes", ".zip"],
  "needsExtract": false
}
```

### System with Both RetroArch and Standalone

```json
{
  "id": "gba",
  "name": "Game Boy Advance",
  "dir": "gba",
  "romJsonFile": "gba.json",
  "emulator": {
    "path": "Emulators/RetroArch/RetroArch-Win64/retroarch.exe",
    "args": ["-L", "cores/mgba_libretro.dll"],
    "name": "RetroArch"
  },
  "standaloneEmulator": {
    "path": "Emulators/mGBA/mGBA-0.10.5-win64/mGBA.exe",
    "args": [],
    "name": "mGBA Standalone"
  },
  "fileExtensions": [".gba", ".zip"],
  "needsExtract": false
}
```

### System with Standalone Only (No RetroArch)

```json
{
  "id": "gc",
  "name": "GameCube",
  "dir": "gc",
  "romJsonFile": "gc.json",
  "emulator": {
    "path": "Emulators/Dolphin/Dolphin-x64/Dolphin.exe",
    "args": ["-e"],
    "name": "Dolphin"
  },
  "standaloneEmulator": null,
  "fileExtensions": [".rvz", ".iso", ".gcm"],
  "needsExtract": true
}
```

## The needsExtract Flag

**Important**: This flag controls ZIP file handling:

- **false**: ZIP files are passed directly to the emulator (RetroArch can handle ZIPs natively)
- **true**: ZIP files are automatically extracted before launching
  - Required for: Dolphin (GameCube/Wii), DeSmuME, Lime3DS, PPSSPP, PCSX2
  - The launcher will extract the ZIP, then launch the extracted file

**Examples:**
- NES with RetroArch: `"needsExtract": false` (RetroArch handles .zip directly)
- GameCube with Dolphin: `"needsExtract": true` (Dolphin needs .rvz extracted)
- Nintendo DS with DeSmuME: `"needsExtract": true` (DeSmuME needs .nds extracted)

## Adding a New System

1. Add your system's emulator to the `Emulators/` directory
2. Create a ROM list JSON file in `1g1rsets/` (e.g., `newsystem.json`)
3. Add a new entry to `systems.json`:

```json
{
  "id": "newsystem",
  "name": "New System Name",
  "dir": "newsystem",
  "romJsonFile": "newsystem.json",
  "emulator": {
    "path": "Emulators/NewEmulator/emulator.exe",
    "args": [],
    "name": "New Emulator"
  },
  "standaloneEmulator": null,
  "fileExtensions": [".ext"],
  "needsExtract": false
}
```

4. Restart the launcher - no recompilation needed!

## Troubleshooting

### "Failed to load systems.json"
- Ensure `systems.json` is in the root EmuBuddy directory
- Check JSON syntax using a JSON validator

### "Emulator not found"
- Verify the `path` in your config matches the actual emulator location
- Use forward slashes (`/`) or escaped backslashes (`\\`) in paths
- Paths are relative to the EmuBuddy root directory

### Games don't launch
- Check `needsExtract` setting - some emulators require extraction
- Verify `fileExtensions` includes all needed formats
- For RetroArch, ensure cores are installed in `Emulators/RetroArch/RetroArch-Win64/cores/`
