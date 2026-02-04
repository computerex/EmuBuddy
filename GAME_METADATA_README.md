# Game Metadata & Embeddings for Recommendation System

This guide explains how to build a local game database with descriptions for creating a game recommendation system with semantic search.

## Overview

The process has 3 steps:

1. **Scrape metadata** from RAWG API - adds descriptions, genres, ratings to your game lists
2. **Build embedding dataset** - prepares data for vector embeddings
3. **Generate embeddings** - create vector embeddings for semantic search (you'll implement this)

## Step 1: Get RAWG API Key

1. Go to https://rawg.io/apidocs
2. Create a free account
3. Get your API key
4. Copy the key - you'll need it in the next step

**Free tier limits:**
- 20,000 requests/month
- ~20 requests/minute
- Perfect for hobby projects

## Step 2: Install Dependencies

```bash
pip install -r requirements-scraper.txt
```

## Step 3: Run Metadata Scraper

⚠️ **IMPORTANT QUOTA INFO:**
- You have **21,259 total games** but only **20,000 API requests/month**
- Each game uses **2 API requests** (search + details)
- Smart filtering skips betas/pirates/demos to save ~15-20% requests
- The scraper has **resume capability** - you can continue later!

Run the scraper:
```bash
python scrape_game_metadata.py
```

**Recommended Strategy:**
1. Start with 1,000 requests to test (~500 games)
2. Choose top systems first (NES, SNES, PS1, N64, GBA)
3. Run monthly to gradually build your database
4. Enable filtering (default) to skip unwanted games

**Options when running:**
- **Max requests**: How many API calls this session (default: 1,000)
- **Filtering**: Skip betas/pirates/demos? (recommended: Yes)
- **Systems**:
  - Press Enter = NES only (test)
  - `top5` = Most popular systems
  - `1,2,3` = Specific systems by number
  - `all` = All systems (uses full quota)

**Output:** Creates `game_metadata/` folder with enriched JSON files

**Resume capability:** If you hit the limit, just run again later - it will skip already-processed games!

## Step 4: Build Embedding Dataset

After scraping metadata, prepare the data for embeddings:

```bash
python build_embeddings.py
```

**Output:** Creates `game_metadata/embedding_dataset.json` with:
- Clean game titles
- Rich text combining title, genres, tags, descriptions
- Metadata (genres, ratings, release dates)
- ROM download info

## Step 5: Generate Embeddings (You Implement This)

Now you can generate vector embeddings using:

### Option A: Sentence Transformers (Local, Free)

```bash
pip install sentence-transformers chromadb
```

```python
from sentence_transformers import SentenceTransformer
import json

# Load embedding dataset
with open('game_metadata/embedding_dataset.json', 'r') as f:
    games = json.load(f)

# Load embedding model
model = SentenceTransformer('all-MiniLM-L6-v2')  # Fast, good quality

# Generate embeddings
texts = [game['embedding_text'] for game in games]
embeddings = model.encode(texts, show_progress_bar=True)

# Store in ChromaDB for semantic search
import chromadb
client = chromadb.Client()
collection = client.create_collection("games")

collection.add(
    ids=[game['id'] for game in games],
    embeddings=embeddings.tolist(),
    metadatas=[game['metadata'] for game in games],
    documents=[game['embedding_text'] for game in games]
)

# Search for games
query = "action platformer with colorful graphics"
results = collection.query(query_texts=[query], n_results=5)
```

### Option B: OpenAI Embeddings (Cloud, Paid)

```python
import openai
import json

with open('game_metadata/embedding_dataset.json', 'r') as f:
    games = json.load(f)

# Generate embeddings with OpenAI
for game in games:
    response = openai.Embedding.create(
        model="text-embedding-3-small",
        input=game['embedding_text']
    )
    game['embedding'] = response['data'][0]['embedding']

# Save embeddings
with open('game_embeddings.json', 'w') as f:
    json.dump(games, f)
```

## Data Structure

### Enriched Game Data (`game_metadata/*_enriched.json`)

```json
{
  "name": "Super Mario Bros. (USA).zip",
  "url": "https://myrient.erista.me/files/...",
  "size": "40.5 KiB",
  "date": "24-Dec-1996 23:32",
  "cleaned_title": "Super Mario Bros.",
  "metadata": {
    "rawg_id": 1234,
    "title": "Super Mario Bros.",
    "description": "A classic platformer...",
    "genres": ["Platformer", "Action"],
    "tags": ["2D", "Retro", "Difficult"],
    "rating": 4.5,
    "metacritic": 94,
    "released": "1985-09-13",
    "playtime": 2,
    "rawg_url": "https://rawg.io/games/super-mario-bros"
  }
}
```

### Embedding Dataset (`game_metadata/embedding_dataset.json`)

```json
{
  "id": "Super Mario Bros._1234",
  "title": "Super Mario Bros.",
  "cleaned_title": "Super Mario Bros.",
  "embedding_text": "Title: Super Mario Bros.\n\nGenres: Platformer, Action\n\nTags: 2D, Retro...",
  "metadata": {
    "rawg_id": 1234,
    "genres": ["Platformer", "Action"],
    "rating": 4.5,
    ...
  },
  "rom_info": {
    "filename": "Super Mario Bros. (USA).zip",
    "url": "https://myrient.erista.me/files/...",
    "size": "40.5 KiB"
  }
}
```

## Recommendation System Ideas

With embeddings, you can build:

1. **Semantic Search**: "Show me action RPGs with great stories"
2. **Similar Games**: Find games similar to user's favorites
3. **Mood-based**: "I want something relaxing" → suggests puzzle games
4. **Genre Mixing**: "Fighting game meets platformer" → finds unique games
5. **Content-based Filtering**: Recommend based on description similarity

## Performance Tips

### For Scraping:
- Start with one system (NES) to test
- Scraper includes rate limiting (respects RAWG API limits)
- Expect ~5-10 seconds per game with rate limiting
- Failed matches are kept (you can manually fix later)

### For Embeddings:
- `all-MiniLM-L6-v2`: Fast, 384 dimensions, good quality
- `all-mpnet-base-v2`: Slower, 768 dimensions, better quality
- ChromaDB: Easy to use, good for <1M vectors
- FAISS: Faster for large datasets, requires more setup

## Troubleshooting

**"Not found" for many games:**
- ROM filenames may not match RAWG database names
- Regional variants (Japan, Europe) may not be in RAWG
- Improve `clean_game_name()` function for better matching

**Rate limit errors:**
- Increase sleep time in scraper
- RAWG free tier: ~20 requests/minute
- Consider paid RAWG plan for faster scraping

**Missing descriptions:**
- Some older/obscure games lack descriptions on RAWG
- Consider fallback to TheGamesDB or ScreenScraper
- Can manually add descriptions for important titles

## Next Steps

1. Test scraper with NES first
2. Review match quality
3. Tune name cleaning if needed
4. Scale to more systems
5. Build embedding dataset
6. Choose embedding approach
7. Build recommendation UI

## Example Queries for Testing

Once embeddings are built, try these searches:

- "Fast-paced action game with weapons"
- "Relaxing adventure game with exploration"
- "Difficult platformer with precise controls"
- "Story-driven RPG with character development"
- "Puzzle game with unique mechanics"
- "Multiplayer party game"

Good luck building your recommendation system!
