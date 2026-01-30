#!/usr/bin/env python3
"""
Download box art from libretro thumbnails for games in 1g1rsets JSON files.
"""

import json
import os
import sys
import urllib.request
import urllib.parse
from pathlib import Path

# Mapping of JSON filenames to libretro thumbnail folder names
SYSTEM_MAPPING = {
    "3ds.json": "Nintendo - Nintendo 3DS",
    "ds.json": "Nintendo - Nintendo DS",
    "gba.json": "Nintendo - Game Boy Advance",
    "gbc.json": "Nintendo - Game Boy Color",
    "gb.json": "Nintendo - Game Boy",
    "nes.json": "Nintendo - Nintendo Entertainment System",
    "snes.json": "Nintendo - Super Nintendo Entertainment System",
    "n64.json": "Nintendo - Nintendo 64",
    "wii.json": "Nintendo - Wii",
    "psp.json": "Sony - PlayStation Portable",
    "dreamcast.json": "Sega - Dreamcast",
    "gc.json": "Nintendo - GameCube",
    "ps1.json": "Sony - PlayStation",
    "ps2.json": "Sony - PlayStation 2",
}

BASE_URL = "https://thumbnails.libretro.com"
OUTPUT_DIR = Path("boxart")


def sanitize_filename(name):
    """Remove problematic characters from filename."""
    # Remove .zip extension
    name = name.replace(".zip", "")
    # Replace common problematic characters
    name = name.replace(":", "")
    name = name.replace("*", "")
    name = name.replace("?", "")
    name = name.replace('"', "")
    name = name.replace("<", "")
    name = name.replace(">", "")
    name = name.replace("|", "")
    return name


def get_boxart_name(game_name):
    """
    Convert game name to match libretro thumbnail naming convention.
    """
    # Remove .zip extension
    name = game_name.replace(".zip", "")
    # URL encode the name
    return urllib.parse.quote(name)


def download_boxart(system_name, game_name, output_path):
    """
    Download box art for a game from libretro thumbnails.
    """
    # Remove .zip extension
    base_name = game_name.replace(".zip", "")

    # Try different naming patterns (these are the actual filenames to try)
    patterns = [
        # Exact match
        base_name,
        # Try with different regions
        f"{base_name.split('(')[0].strip()} (USA)",
        f"{base_name.split('(')[0].strip()} (World)",
        f"{base_name.split('(')[0].strip()} (Europe)",
        f"{base_name.split('(')[0].strip()} (Japan)",
        # Try just base name
        base_name.split('(')[0].strip(),
    ]

    for pattern in patterns:
        # URL encode the entire pattern
        encoded_pattern = urllib.parse.quote(pattern)
        url = f"{BASE_URL}/{urllib.parse.quote(system_name)}/Named_Boxarts/{encoded_pattern}.png"
        try:
            # Try to download with HEAD request first to check if file exists
            req = urllib.request.Request(url, method="HEAD")
            with urllib.request.urlopen(req, timeout=10) as response:
                if response.status == 200:
                    # Download the actual image
                    urllib.request.urlretrieve(url, output_path)
                    print(f"  Downloaded: {game_name}")
                    return True
        except (urllib.error.HTTPError, urllib.error.URLError):
            continue
        except Exception as e:
            print(f"  Error downloading {url}: {e}")
            continue

    return False


def main():
    # Create output directory
    OUTPUT_DIR.mkdir(exist_ok=True)

    json_dir = Path("1g1rsets")
    if not json_dir.exists():
        print(f"Error: {json_dir} directory not found!")
        sys.exit(1)

    total_games = 0
    downloaded = 0
    failed = []

    for json_file, system_name in SYSTEM_MAPPING.items():
        json_path = json_dir / json_file
        if not json_path.exists():
            print(f"Warning: {json_file} not found, skipping...")
            continue

        print(f"\nProcessing {json_file} -> {system_name}")

        # Create system-specific output directory
        system_output_dir = OUTPUT_DIR / system_name
        system_output_dir.mkdir(exist_ok=True)

        with open(json_path, "r", encoding="utf-8") as f:
            games = json.load(f)

        for game in games:
            game_name = game.get("name", "")
            if not game_name:
                continue

            total_games += 1
            sanitized_name = sanitize_filename(game_name)
            output_path = system_output_dir / f"{sanitized_name}.png"

            # Skip if already downloaded
            if output_path.exists():
                print(f"  Already exists: {game_name}")
                downloaded += 1
                continue

            if download_boxart(system_name, game_name, output_path):
                downloaded += 1
            else:
                print(f"  Not found: {game_name}")
                failed.append(game_name)

    # Print summary
    print("\n" + "=" * 60)
    print("DOWNLOAD SUMMARY")
    print("=" * 60)
    print(f"Total games: {total_games}")
    print(f"Downloaded: {downloaded}")
    print(f"Failed: {len(failed)}")

    if failed:
        print("\nFailed downloads:")
        for game in failed[:20]:  # Show first 20
            print(f"  - {game}")
        if len(failed) > 20:
            print(f"  ... and {len(failed) - 20} more")


if __name__ == "__main__":
    main()
