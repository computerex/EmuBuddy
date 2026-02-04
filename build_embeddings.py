#!/usr/bin/env python3
"""
Build embeddings database for game recommendations.
Processes enriched game metadata and creates vector embeddings for semantic search.
"""

import json
from pathlib import Path
from typing import List, Dict

def load_enriched_games(metadata_dir: Path) -> List[Dict]:
    """Load all enriched game data from metadata directory."""
    all_games = []

    for json_file in metadata_dir.glob("*_enriched.json"):
        print(f"Loading {json_file.name}...")
        with open(json_file, 'r', encoding='utf-8') as f:
            games = json.load(f)

        # Filter games that have metadata
        games_with_metadata = [
            g for g in games
            if g.get('metadata') is not None
            and g['metadata'].get('description')
        ]

        print(f"  Found {len(games_with_metadata)} games with descriptions")
        all_games.extend(games_with_metadata)

    return all_games


def create_embedding_text(game: Dict) -> str:
    """
    Create a rich text representation for embedding.
    Combines title, description, genres, and tags for better semantic search.
    """
    metadata = game['metadata']

    parts = []

    # Title
    parts.append(f"Title: {metadata['title']}")

    # Genres
    if metadata.get('genres'):
        parts.append(f"Genres: {', '.join(metadata['genres'])}")

    # Tags (gameplay elements, themes)
    if metadata.get('tags'):
        parts.append(f"Tags: {', '.join(metadata['tags'])}")

    # Description (main content for embeddings)
    if metadata.get('description'):
        parts.append(f"Description: {metadata['description']}")

    return "\n\n".join(parts)


def build_embedding_dataset(output_file: Path):
    """
    Build a dataset ready for embedding generation.
    Creates a JSON file with game ID, text for embedding, and metadata.
    """
    project_root = Path(__file__).parent
    metadata_dir = project_root / "game_metadata"

    if not metadata_dir.exists():
        print(f"ERROR: {metadata_dir} not found")
        print("Please run scrape_game_metadata.py first")
        return

    print("Loading enriched game data...")
    games = load_enriched_games(metadata_dir)

    print(f"\nTotal games with descriptions: {len(games)}")

    if not games:
        print("No games found. Please run scrape_game_metadata.py first.")
        return

    # Build embedding dataset
    print("\nBuilding embedding dataset...")
    embedding_dataset = []

    for game in games:
        metadata = game['metadata']

        entry = {
            "id": f"{game.get('cleaned_title', '')}_{metadata['rawg_id']}",
            "title": metadata['title'],
            "cleaned_title": game.get('cleaned_title', ''),
            "embedding_text": create_embedding_text(game),
            "metadata": {
                "rawg_id": metadata['rawg_id'],
                "genres": metadata.get('genres', []),
                "tags": metadata.get('tags', []),
                "rating": metadata.get('rating', 0),
                "released": metadata.get('released', ''),
                "playtime": metadata.get('playtime', 0),
                "platforms": metadata.get('platforms', []),
                "rawg_url": metadata.get('rawg_url', ''),
            },
            "rom_info": {
                "filename": game['name'],
                "url": game['url'],
                "size": game['size'],
            }
        }

        embedding_dataset.append(entry)

    # Save dataset
    output_file.parent.mkdir(parents=True, exist_ok=True)
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(embedding_dataset, f, indent=2, ensure_ascii=False)

    print(f"\n✓ Saved embedding dataset to: {output_file}")
    print(f"  Total entries: {len(embedding_dataset)}")

    # Print statistics
    print("\n" + "="*60)
    print("Dataset Statistics:")
    print("="*60)

    # Genre distribution
    all_genres = {}
    for game in embedding_dataset:
        for genre in game['metadata']['genres']:
            all_genres[genre] = all_genres.get(genre, 0) + 1

    print("\nTop Genres:")
    for genre, count in sorted(all_genres.items(), key=lambda x: x[1], reverse=True)[:10]:
        print(f"  {genre}: {count}")

    # Average description length
    avg_length = sum(len(g['embedding_text']) for g in embedding_dataset) / len(embedding_dataset)
    print(f"\nAverage embedding text length: {avg_length:.0f} characters")

    print("\nNext steps:")
    print("  1. Use an embedding model (e.g., OpenAI, Sentence-Transformers)")
    print("  2. Generate embeddings for each game's embedding_text")
    print("  3. Store embeddings in a vector database (ChromaDB, FAISS, Pinecone)")
    print("  4. Build semantic search for game recommendations")


def show_example_games(n=5):
    """Show some example games from the dataset."""
    project_root = Path(__file__).parent
    dataset_file = project_root / "game_metadata" / "embedding_dataset.json"

    if not dataset_file.exists():
        print("Dataset not built yet. Run build_embedding_dataset() first.")
        return

    with open(dataset_file, 'r', encoding='utf-8') as f:
        dataset = json.load(f)

    print(f"\n{'='*60}")
    print(f"Example Games (showing {min(n, len(dataset))} of {len(dataset)})")
    print(f"{'='*60}\n")

    for game in dataset[:n]:
        print(f"Title: {game['title']}")
        print(f"Genres: {', '.join(game['metadata']['genres'])}")
        print(f"Rating: {game['metadata']['rating']}/5")
        print(f"Embedding text preview:")
        print(f"  {game['embedding_text'][:200]}...")
        print(f"\n{'-'*60}\n")


if __name__ == "__main__":
    project_root = Path(__file__).parent
    output_file = project_root / "game_metadata" / "embedding_dataset.json"

    build_embedding_dataset(output_file)

    # Show some examples
    show_example_games(3)
