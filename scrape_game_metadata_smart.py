#!/usr/bin/env python3
"""
Smart Game Metadata Scraper - Prioritizes newer games, includes betas without final versions
"""

import json
import re
import time
from pathlib import Path
from typing import Dict, List, Optional, Set
import requests
from datetime import datetime

# RAWG API Configuration
RAWG_API_KEY = "bf7311e9b9b746739d4de3da52d4ada0"
RAWG_BASE_URL = "https://api.rawg.io/api"

# API Quota Management
API_MONTHLY_LIMIT = 20000
API_REQUESTS_THIS_RUN = 0

# Platform mapping - CORRECTED IDs from RAWG API!
PLATFORM_MAPPING = {
    "nes": [49], "snes": [79], "n64": [83], "gb": [26], "gbc": [43],
    "gba": [24], "ds": [9], "3ds": [8], "gc": [105], "wii": [11],
    "wiiu": [10], "ps1": [27], "ps2": [15], "psp": [17],
    "dreamcast": [106], "genesis": [167], "sms": [74], "gamegear": [77],
    "saturn": [107], "tg16": [112], "virtualboy": [87], "atari2600": [23],
    "atari7800": [28], "lynx": [46], "ngpc": [119], "ngp": [119],
    "coleco": [45], "intellivision": [115], "wonderswan": [57],
    "wonderswancolor": [57],
}

# System priority (newer systems first) - FIXED WITH CORRECT PLATFORM IDs!
SYSTEM_PRIORITY = [
    'ps2', 'ds', 'wii', 'psp', 'wiiu', '3ds', 'gc', 'gba', 'ps1',
    'dreamcast', 'n64', 'saturn', 'snes', 'nes', 'genesis', 'gbc', 'gb',
    'sms', 'gamegear', 'tg16', 'atari7800', 'atari2600', 'ngpc', 'lynx',
    'virtualboy', 'wonderswancolor', 'wonderswan', 'ngp', 'coleco', 'intellivision'
]

# Always skip these (not in RAWG)
ALWAYS_SKIP_PATTERNS = [
    r'\(Pirate\)',         # Pirate/bootleg games
    r'\(Homebrew\)',       # Homebrew (not in RAWG)
    r'\(Hack\)',           # ROM hacks
    r'\d+-in-1',           # Multicarts
]


def clean_game_name(filename: str) -> str:
    """Extract clean game title from ROM filename."""
    name = re.sub(r'\.(zip|nes|snes|gb|gba|nds|iso|bin|chd|cue|rvz|gcz|wbfs|pbp|cso)$', '', filename, flags=re.IGNORECASE)
    name = re.sub(r'\s*\([^)]*\)', '', name)
    name = re.sub(r'\s*\[[^\]]*\]', '', name)
    name = re.sub(r',\s*The$', '', name)
    name = ' '.join(name.split())
    return name.strip()


def has_final_version(filename: str, all_games: List[Dict]) -> bool:
    """Check if a beta/proto/demo has a final version in the game list."""
    # Extract the base game name
    base_name = clean_game_name(filename)

    # Check if this is a beta/proto/demo
    is_special = any(re.search(pattern, filename, re.IGNORECASE)
                     for pattern in [r'\(Beta\)', r'\(Proto\)', r'\(Demo\)', r'\(Sample\)', r'\(Unl\)'])

    if not is_special:
        return False  # Not a special version, don't skip

    # Look for a final version (same name without Beta/Proto/Demo tags)
    for game in all_games:
        other_name = clean_game_name(game['name'])
        other_is_special = any(re.search(pattern, game['name'], re.IGNORECASE)
                              for pattern in [r'\(Beta\)', r'\(Proto\)', r'\(Demo\)', r'\(Sample\)', r'\(Unl\)'])

        # If we find the same game without special tags, a final version exists
        if other_name == base_name and not other_is_special:
            return True

    return False  # No final version found


def should_skip_game(filename: str, all_games: List[Dict]) -> tuple[bool, str]:
    """Check if a game should be skipped."""
    # Always skip pirates, hacks, multicarts
    for pattern in ALWAYS_SKIP_PATTERNS:
        if re.search(pattern, filename, re.IGNORECASE):
            return True, pattern

    # Only skip beta/proto/demo if a final version exists
    if has_final_version(filename, all_games):
        return True, "(Beta/Proto with final version)"

    return False, ""


def search_game_on_rawg(game_name: str, platform_ids: Optional[List[int]] = None) -> Optional[Dict]:
    """Search for a game on RAWG API."""
    global API_REQUESTS_THIS_RUN

    try:
        params = {
            "key": RAWG_API_KEY,
            "search": game_name,
            "page_size": 5
        }

        if platform_ids:
            params["platforms"] = ",".join(map(str, platform_ids))

        response = requests.get(f"{RAWG_BASE_URL}/games", params=params, timeout=10)
        response.raise_for_status()
        API_REQUESTS_THIS_RUN += 1

        data = response.json()
        results = data.get("results", [])

        if not results:
            return None

        game = results[0]
        game_id = game["id"]

        detail_response = requests.get(
            f"{RAWG_BASE_URL}/games/{game_id}",
            params={"key": RAWG_API_KEY},
            timeout=10
        )
        detail_response.raise_for_status()
        API_REQUESTS_THIS_RUN += 1

        detail_data = detail_response.json()

        metadata = {
            "rawg_id": game_id,
            "title": detail_data.get("name", game_name),
            "description": detail_data.get("description_raw", ""),
            "description_html": detail_data.get("description", ""),
            "released": detail_data.get("released", ""),
            "rating": detail_data.get("rating", 0),
            "metacritic": detail_data.get("metacritic"),
            "genres": [g["name"] for g in detail_data.get("genres", [])],
            "tags": [t["name"] for t in detail_data.get("tags", [])[:10]],
            "platforms": [p["platform"]["name"] for p in detail_data.get("platforms", [])],
            "developers": [d["name"] for d in detail_data.get("developers", [])],
            "publishers": [p["name"] for p in detail_data.get("publishers", [])],
            "playtime": detail_data.get("playtime", 0),
            "rawg_url": f"https://rawg.io/games/{detail_data.get('slug', '')}",
        }

        return metadata

    except requests.RequestException as e:
        print(f"Error: {e}")
        return None


def load_progress(output_file: Path) -> tuple[List[Dict], Set[str]]:
    """Load existing progress to enable resume."""
    if output_file.exists():
        with open(output_file, 'r', encoding='utf-8') as f:
            existing_games = json.load(f)
        processed_names = {g["name"] for g in existing_games}
        return existing_games, processed_names
    return [], set()


def save_progress(output_file: Path, games: List[Dict]):
    """Save progress immediately."""
    output_file.parent.mkdir(parents=True, exist_ok=True)
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(games, f, indent=2, ensure_ascii=False)


def enrich_game_list(input_file: Path, output_file: Path, system_id: str, max_requests: Optional[int] = None):
    """Enrich a game list with smart filtering."""
    global API_REQUESTS_THIS_RUN

    print(f"\n{'='*70}")
    print(f"Processing: {system_id.upper()}")
    print(f"{'='*70}")

    # Load game list
    with open(input_file, 'r', encoding='utf-8') as f:
        games = json.load(f)

    print(f"Total games in list: {len(games)}")

    # Load progress
    enriched_games, processed_names = load_progress(output_file)

    if processed_names:
        print(f"Resuming: {len(processed_names)} games already processed")

    platform_ids = PLATFORM_MAPPING.get(system_id)

    stats = {
        "success": len([g for g in enriched_games if g.get('metadata')]),
        "failed": len([g for g in enriched_games if not g.get('metadata')]),
        "skipped": 0,
    }

    for idx, game in enumerate(games, 1):
        original_name = game["name"]

        # Skip if already processed
        if original_name in processed_names:
            continue

        # Check API quota
        if max_requests and API_REQUESTS_THIS_RUN >= max_requests:
            print(f"\nWARNING: Reached API limit ({max_requests})")
            print(f"Processed {stats['success']} games. Progress saved!")
            save_progress(output_file, enriched_games)
            return stats

        # Smart skip check (includes beta check)
        should_skip, skip_reason = should_skip_game(original_name, games)
        if should_skip:
            print(f"[{idx}/{len(games)}] SKIP {original_name[:50]}... ({skip_reason})")
            enriched_games.append({
                **game,
                "cleaned_title": clean_game_name(original_name),
                "metadata": None,
                "skip_reason": skip_reason
            })
            stats["skipped"] += 1
            continue

        clean_name = clean_game_name(original_name)
        # Handle Unicode errors on Windows console
        try:
            print(f"[{idx}/{len(games)}] {clean_name[:50]}...", end=" ", flush=True)
        except UnicodeEncodeError:
            print(f"[{idx}/{len(games)}] [Unicode name]...", end=" ", flush=True)

        metadata = search_game_on_rawg(clean_name, platform_ids)

        if metadata:
            enriched_games.append({
                **game,
                "cleaned_title": clean_name,
                "metadata": metadata
            })
            print(f"OK ({API_REQUESTS_THIS_RUN})")
            stats["success"] += 1
        else:
            enriched_games.append({
                **game,
                "cleaned_title": clean_name,
                "metadata": None
            })
            print(f"X ({API_REQUESTS_THIS_RUN})")
            stats["failed"] += 1

        # Save every 10 games
        if (stats["success"] + stats["failed"]) % 10 == 0:
            save_progress(output_file, enriched_games)

        # Rate limiting: 3.5 seconds between requests
        time.sleep(3.5)

    save_progress(output_file, enriched_games)

    print(f"\n{'='*70}")
    print(f"Completed: {system_id}")
    print(f"  Success: {stats['success']:5}")
    print(f"  Failed:  {stats['failed']:5}")
    print(f"  Skipped: {stats['skipped']:5}")
    print(f"{'='*70}")

    return stats


def main():
    """Main entry point."""
    global API_REQUESTS_THIS_RUN

    print("="*70)
    print("EmuBuddy Smart Metadata Scraper")
    print("="*70)
    print("Strategy: Newer systems first, include betas without final versions")
    print(f"Available quota: ~19,900 requests")
    print("="*70)

    project_root = Path(__file__).parent
    input_dir = project_root / "1g1rsets"
    output_dir = project_root / "game_metadata"

    # Load systems
    systems_file = project_root / "systems.json"
    with open(systems_file, 'r', encoding='utf-8') as f:
        systems_config = json.load(f)
    systems = systems_config["systems"]

    # Sort systems by priority (newer first)
    systems_by_priority = []
    for system_id in SYSTEM_PRIORITY:
        system = next((s for s in systems if s['id'] == system_id), None)
        if system:
            systems_by_priority.append(system)

    # Show plan
    print("\nProcessing order (newer systems first):")
    for idx, system in enumerate(systems_by_priority[:10], 1):
        rom_json = system.get('romJsonFile')
        if rom_json:
            json_path = input_dir / rom_json
            if json_path.exists():
                with open(json_path, 'r', encoding='utf-8') as f:
                    count = len(json.load(f))
                print(f"  {idx:2}. {system['name']:35} ({count:5,} games)")

    print(f"  ... and {len(systems_by_priority) - 10} more systems")

    max_requests = 2480  # ALL remaining quota - exhaust everything!
    print(f"\nUsing {max_requests:,} API requests (ALL remaining quota)")

    confirm = input("\nStart scraping? (Y/n): ").strip().lower()
    if confirm == 'n':
        return

    # Process systems
    total_stats = {"success": 0, "failed": 0, "skipped": 0}
    start_time = datetime.now()

    for system in systems_by_priority:
        if API_REQUESTS_THIS_RUN >= max_requests:
            print(f"\nReached max API requests ({max_requests:,})")
            break

        system_id = system['id']
        rom_json_file = system.get('romJsonFile')

        if not rom_json_file:
            continue

        input_file = input_dir / rom_json_file
        if not input_file.exists():
            continue

        output_file = output_dir / f"{system_id}_enriched.json"

        stats = enrich_game_list(
            input_file, output_file, system_id,
            max_requests=max_requests - API_REQUESTS_THIS_RUN
        )

        total_stats["success"] += stats["success"]
        total_stats["failed"] += stats["failed"]
        total_stats["skipped"] += stats["skipped"]

        # Show running totals
        print(f"\nRunning totals: {total_stats['success']} success, "
              f"{API_REQUESTS_THIS_RUN:,} API calls used, "
              f"{max_requests - API_REQUESTS_THIS_RUN:,} remaining\n")

    # Final summary
    elapsed = (datetime.now() - start_time).total_seconds()

    print("\n" + "="*70)
    print("FINAL SUMMARY")
    print("="*70)
    print(f"API Requests: {API_REQUESTS_THIS_RUN:,} / {max_requests:,}")
    print(f"Time: {elapsed/60:.1f} minutes ({elapsed/3600:.1f} hours)")
    print(f"Success:  {total_stats['success']:6,} games")
    print(f"Failed:   {total_stats['failed']:6,} games")
    print(f"Skipped:  {total_stats['skipped']:6,} games")
    print(f"Remaining quota: ~{API_MONTHLY_LIMIT - API_REQUESTS_THIS_RUN:,}")
    print("="*70)
    print("Progress saved! Run again to continue if stopped early.")


if __name__ == "__main__":
    main()
