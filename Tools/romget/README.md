# romget - Simple ROM Downloader

Unix-style single-file ROM downloader with myrient.erista.me support.

## Features

- **Single ROM download** - Does one thing well (Unix philosophy)
- **Myrient-compatible** - Proper Referer and browser headers
- **Auto-retry** - Configurable retry attempts with backoff
- **Progress display** - Real-time download progress
- **Resume detection** - Skips if file already exists
- **Cross-platform** - Pure Go, works on Windows/Linux/macOS

## Installation

```bash
cd tools/romget
go build -o romget
```

Or for direct installation:
```bash
go install github.com/emubuddy/romget@latest
```

## Usage

### Basic Usage
```bash
# Download with auto-detected filename
romget -url "https://myrient.erista.me/files/No-Intro/Nintendo%20-%20Nintendo%20Entertainment%20System%20(Headered)/Super%20Mario%20Bros.%20(World).zip"

# Download with custom output path
romget -url "https://myrient.erista.me/files/.../game.zip" -o roms/nes/mario.zip
```

### Advanced Options
```bash
# More retries and longer timeout
romget -url "https://example.com/rom.zip" -r 5 -t 120

# Quiet mode (no progress output)
romget -url "https://example.com/rom.zip" -q

# Custom referer (auto-detected by default)
romget -url "https://example.com/rom.zip" -referer "https://example.com/roms/"
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-url` | *required* | URL to download |
| `-o` | auto-detect | Output file path |
| `-r` | 3 | Number of retry attempts |
| `-t` | 60 | Timeout in seconds |
| `-referer` | auto-detect | HTTP Referer header (inferred from URL parent dir) |
| `-ua` | Edge/Linux | User-Agent string |
| `-q` | false | Quiet mode (no progress) |

## How Myrient Support Works

The tool mimics a real browser to work with myrient.erista.me:

1. **Referer Header** - Auto-detected from parent directory:
   - URL: `https://myrient.erista.me/files/No-Intro/Nintendo/.../game.zip`
   - Referer: `https://myrient.erista.me/files/No-Intro/Nintendo/.../`

2. **Browser Headers** - Full Edge/Chrome header set:
   - User-Agent, Accept, Sec-Fetch-*, sec-ch-ua headers
   - Connection: keep-alive

3. **Auto-retry** - Handles transient failures with exponential backoff

## Examples

### Download NES ROM
```bash
romget -url "https://myrient.erista.me/files/No-Intro/Nintendo%20-%20Nintendo%20Entertainment%20System%20(Headered)/Super%20Mario%20Bros.%20(World).zip"
```

### Download to specific directory
```bash
mkdir -p roms/snes
romget -url "https://myrient.erista.me/files/.../Zelda.zip" -o roms/snes/zelda.zip
```

### Use in shell scripts
```bash
#!/bin/bash
# Download multiple ROMs
while IFS= read -r url; do
    romget -url "$url" || echo "Failed: $url"
done < urls.txt
```

### Integrate with JSON (using jq)
```bash
# Extract URLs from 1g1r JSON and download
jq -r '.[] | .url' games_1g1r_english_nes.json | while read url; do
    romget -url "$url" -o "roms/nes/$(basename "$url")"
done
```

## Building

### All platforms
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o romget-linux

# Windows
GOOS=windows GOARCH=amd64 go build -o romget.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o romget-macos
```

## Exit Codes

- `0` - Success (or file already exists)
- `1` - Error (download failed, invalid arguments, etc.)

## Design Philosophy

**Unix Philosophy:**
- Do one thing well: Download a single file
- Composable: Use with pipes, loops, xargs
- Standard I/O: Progress to stderr, errors to stderr
- Exit codes: Proper success/failure signaling

**Why not use wget/curl?**
- Myrient requires proper Referer header (auto-detected here)
- Built-in retry logic with backoff
- Clean progress output
- Cross-platform Go binary (no dependencies)
