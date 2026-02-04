#!/usr/bin/env python3
"""
Game Metadata Scraper for EmuBuddy
Enriches game ROM lists with descriptions from RAWG API for recommendation system.

IMPORTANT: Free tier has 20,000 requests/month limit!
This script includes resume capability and quota tracking.
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

# Platform mapping from system ID to RAWG platform IDs
PLATFORM_MAPPING = {
    "nes": [49], "snes": [51], "n64": [52], "gb": [43], "gbc": [44],
    "gba": [45], "ds": [48], "3ds": [47], "gc": [46], "wii": [53],
    "wiiu": [54], "ps1": [10], "ps2": [15], "psp": [16],
    "dreamcast": [21], "genesis": [22], "sms": [23], "gamegear": [24],
    "saturn": [25], "tg16": [32], "virtualboy": [39], "atari2600": [33],
    "atari7800": [34], "lynx": [35], "ngpc": [31], "ngp": [31],
    "coleco": [40], "intellivision": [41], "wonderswan": [42],
    "wonderswancolor": [42],
}

# Skip patterns to save API quota
SKIP_PATTERNS = [
    r'\(Beta\)',           # Beta versions
    r'\(Proto\)',          # Prototypes
    r'\(Sample\)',         # Sample versions
    r'\(Demo\)',           # Demo versions
    r'\(Pirate\)',         # Pirate/bootleg games
    r'\(Unl\)',            # Unlicensed
    r'\(Homebrew\)',       # Homebrew (not in RAWG)
    r'\(Hack\)',           # ROM hacks
    r'\d+-in-1',           # Multicarts (23-in-1, etc)
]


def should_skip_game(filename: str, enable_filtering: bool = True) -> tuple[bool, str]:
    """Check if a game should be skipped to save API quota."""
    if not enable_filtering:
        return False, ""

    for pattern in SKIP_PATTERNS:
        if re.search(pattern, filename, re.IGNORECASE):
            return True, pattern
    return False, ""


def clean_game_name(filename: str) -> str:
    """Extract clean game title from ROM filename."""
    name = re.sub(r'\.(zip|nes|snes|gb|gba|nds|iso|bin|chd|cue|rvz|gcz|wbfs|pbp|cso)$', '', filename, flags=re.IGNORECASE)
    name = re.sub(r'\s*\([^)]*\)', '', name)
    name = re.sub(r'\s*\[[^\]]*\]', '', name)
    name = re.sub(r',\s*The$', '', name)
    name = ' '.join(name.split())
    return name.strip()


def search_game_on_rawg(game_name: str, platform_ids: Optional[List[int]] = None) -> Optional[Dict]:
    """Search for a game on RAWG API."""
    global API_REQUESTS_THIS_RUN

    try:
        # Search for the game (1 API request)
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

        # Get detailed game info (1 more API request)
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

        # Get set of already processed game names
        processed_names = {g["name"] for g in existing_games}
        return existing_games, processed_names

    return [], set()


def save_progress(output_file: Path, games: List[Dict]):
    """Save progress immediately to prevent data loss."""
    output_file.parent.mkdir(parents=True, exist_ok=True)
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(games, f, indent=2, ensure_ascii=False)


def enrich_game_list(input_file: Path, output_file: Path, system_id: str,
                     max_requests: Optional[int] = None, enable_filtering: bool = True):
    """Enrich a game list JSON file with metadata from RAWG."""
    global API_REQUESTS_THIS_RUN

    print(f"\n{'='*70}")
    print(f"Processing: {input_file.name}")
    print(f"System: {system_id}")
    print(f"{'='*70}")

    # Load existing game list
    with open(input_file, 'r', encoding='utf-8') as f:
        games = json.load(f)

    print(f"Total games in list: {len(games)}")

    # Load progress (resume capability)
    enriched_games, processed_names = load_progress(output_file)

    if processed_names:
        print(f"OK Found existing progress: {len(processed_names)} games already processed")
        print(f"  Will resume from where we left off")

    # Get platform IDs
    platform_ids = PLATFORM_MAPPING.get(system_id)

    # Track statistics
    stats = {
        "success": len([g for g in enriched_games if g.get('metadata')]),
        "failed": len([g for g in enriched_games if not g.get('metadata')]),
        "skipped": 0,
        "already_done": len(processed_names),
    }

    # Process games
    for idx, game in enumerate(games, 1):
        original_name = game["name"]

        # Skip if already processed
        if original_name in processed_names:
            continue

        # Check if we should skip to save quota
        should_skip, skip_reason = should_skip_game(original_name, enable_filtering)
        if should_skip:
            print(f"[{idx}/{len(games)}] SKIP Skipping: {original_name[:60]}... ({skip_reason})")
            enriched_games.append({
                **game,
                "cleaned_title": clean_game_name(original_name),
                "metadata": None,
                "skip_reason": skip_reason
            })
            stats["skipped"] += 1
            continue

        # Check API quota
        if max_requests and API_REQUESTS_THIS_RUN >= max_requests:
            print(f"\nWARNING  Reached API request limit ({max_requests})")
            print(f"   Processed {stats['success']} games successfully")
            print(f"   Saving progress... You can resume later!")
            save_progress(output_file, enriched_games)
            return stats

        clean_name = clean_game_name(original_name)
        print(f"[{idx}/{len(games)}] {clean_name[:55]}...", end=" ", flush=True)

        # Search for metadata (uses 2 API requests: search + details)
        metadata = search_game_on_rawg(clean_name, platform_ids)

        if metadata:
            enriched_games.append({
                **game,
                "cleaned_title": clean_name,
                "metadata": metadata
            })
            print(f"OK ({API_REQUESTS_THIS_RUN} API calls)")
            stats["success"] += 1
        else:
            enriched_games.append({
                **game,
                "cleaned_title": clean_name,
                "metadata": None
            })
            print(f"X Not found ({API_REQUESTS_THIS_RUN} API calls)")
            stats["failed"] += 1

        # Save progress every 10 games
        if (stats["success"] + stats["failed"]) % 10 == 0:
            save_progress(output_file, enriched_games)

        # Rate limiting: ~20 requests/minute = 3 seconds between game lookups
        time.sleep(3.5)

    # Final save
    save_progress(output_file, enriched_games)

    print(f"\n{'='*70}")
    print(f"OK Completed: {output_file.name}")
    print(f"  Success:  {stats['success']:5} games")
    print(f"  Failed:   {stats['failed']:5} games")
    print(f"  Skipped:  {stats['skipped']:5} games (betas, pirates, etc.)")
    print(f"  API calls: {API_REQUESTS_THIS_RUN}")
    print(f"{'='*70}")

    return stats


def main():
    """Main entry point."""
    global API_REQUESTS_THIS_RUN

    print("="*70)
    print("EmuBuddy Game Metadata Scraper")
    print("="*70)
    print(f"WARNING: API LIMIT: {API_MONTHLY_LIMIT:,} requests/month")
    print(f"   Total games: 21,259 (you're 1,259 over quota)")
    print(f"   Smart filtering enabled to save requests")
    print("="*70)

    project_root = Path(__file__).parent
    input_dir = project_root / "1g1rsets"
    output_dir = project_root / "game_metadata"

    # Load systems
    systems_file = project_root / "systems.json"
    with open(systems_file, 'r', encoding='utf-8') as f:
        systems_config = json.load(f)
    systems = systems_config["systems"]

    # Show system options with game counts
    print("\nAvailable systems:")
    for idx, system in enumerate(systems, 1):
        rom_json = system.get('romJsonFile')
        if rom_json:
            json_path = input_dir / rom_json
            if json_path.exists():
                with open(json_path, 'r', encoding='utf-8') as f:
                    count = len(json.load(f))
                print(f"  {idx:2}. {system['name']:40} ({count:5,} games)")

    print("\nOptions:")
    print("  - Enter system numbers (comma-separated, e.g., '1,2,3')")
    print("  - Enter 'top5' for top 5 most popular systems")
    print("  - Enter 'all' to process all systems (will hit quota limit)")
    print("  - Press Enter for NES only (testing)")

    # Ask for max requests limit
    print("\nAPI Quota Management:")
    choice = input(f"  Max API requests for this run (default: 1000, max: {API_MONTHLY_LIMIT}): ").strip()
    max_requests = 1000
    if choice.isdigit():
        max_requests = min(int(choice), API_MONTHLY_LIMIT)

    print(f"\n  Using max {max_requests:,} API requests for this session")

    # Ask for filtering
    filter_choice = input("  Skip betas/pirates/demos to save quota? (Y/n): ").strip().lower()
    enable_filtering = filter_choice != 'n'

    # System selection
    system_choice = input("\nSelect systems: ").strip()

    if system_choice.lower() == 'all':
        selected_systems = systems
    elif system_choice.lower() == 'top5':
        # Top 5 by popularity/quality
        top_ids = ['nes', 'snes', 'n64', 'ps1', 'gba']
        selected_systems = [s for s in systems if s['id'] in top_ids]
    elif system_choice == '':
        selected_systems = [s for s in systems if s['id'] == 'nes']
    else:
        try:
            indices = [int(x.strip()) - 1 for x in system_choice.split(',')]
            selected_systems = [systems[i] for i in indices]
        except (ValueError, IndexError):
            print("Invalid selection")
            return

    print(f"\nWill process {len(selected_systems)} system(s)")
    print(f"Filtering enabled: {enable_filtering}")
    print(f"Max API requests: {max_requests:,}")

    input("\nPress Enter to start scraping...")

    # Process systems
    total_stats = {"success": 0, "failed": 0, "skipped": 0}
    start_time = datetime.now()

    for system in selected_systems:
        if API_REQUESTS_THIS_RUN >= max_requests:
            print(f"\nWARNING  Reached max API requests ({max_requests})")
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
            max_requests=max_requests - API_REQUESTS_THIS_RUN,
            enable_filtering=enable_filtering
        )

        total_stats["success"] += stats["success"]
        total_stats["failed"] += stats["failed"]
        total_stats["skipped"] += stats["skipped"]

    # Final summary
    elapsed = (datetime.now() - start_time).total_seconds()

    print("\n" + "="*70)
    print("FINAL SUMMARY")
    print("="*70)
    print(f"API Requests Used: {API_REQUESTS_THIS_RUN:,} / {max_requests:,}")
    print(f"Time Elapsed: {elapsed/60:.1f} minutes")
    print(f"Success:  {total_stats['success']:5} games")
    print(f"Failed:   {total_stats['failed']:5} games")
    print(f"Skipped:  {total_stats['skipped']:5} games")
    print(f"Remaining quota: ~{API_MONTHLY_LIMIT - API_REQUESTS_THIS_RUN:,} requests")
    print("="*70)
    print("OK Progress saved! You can run this script again to continue.")


if __name__ == "__main__":
    main()
