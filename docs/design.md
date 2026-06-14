## 1. Overview
Yomarr follows the "Arr" philosophy: Index -> Search -> Grab -> Process -> Organize.

## 2. Core Components

### Indexer Manager
- Responsible for polling/searching sources.
- Supports prioritizing specific scanlation groups.
- Handles rate-limiting and bypasses (e.g., via FlareSolverr).

### Metadata Engine
- Primary ID source: AniList / MangaDex.
- Stores local cache of covers to avoid hotlinking issues.
- Handles title aliases (e.g., "Attack on Titan" vs "Shingeki no Kyojin").

### Download Handler
- Supports multiple download protocols.
- Monitors progress and reports status back to the UI.

### Post-Processor
- Validates downloaded images (removes ads/credits if possible).
- Packs images into `.cbz` (ZIP format).
- Injects `ComicInfo.xml` metadata for reader compatibility.

## 3. Database Schema Concept
- **Series Table:** Title, IDs, Path, Status (Monitored/Ended).
- **Chapters Table:** Number, Volume, Release Date, File Path, Status (Missing/Downloaded).
- **Indexers Table:** URL, API Key, Priority.
- **History Table:** Event logs for auditing.

## 4. Filesystem Layout
- Base Path: `/Manga`
- Structure: `{Series Title} ({Year}) [{ProviderID}]/{Series Title} - c{ChapterNum} [{Group}].cbz`
