# Yomarr

**Yomarr** is an automated manga collection manager (a Sonarr/Radarr equivalent for Manga). It monitors your favorite series for new chapters, handles the downloads via your preferred client, and organizes them into a clean, tagged library.

## Project Goals
- **Automated Tracking:** Automatically find and fetch new chapters as they release.
- **Metadata Management:** Enrich your collection with high-quality metadata from AniList, MyAnimeList, and MangaDex.
- **Library Organization:** Standardize folder structures and file naming.
- **Portable Format:** Convert downloads to `.cbz` with embedded `ComicInfo.xml` for maximum compatibility with readers like Kavita or Komga.

## High-Level Architecture
- **Indexer Manager:** Connects to MangaDex, generic RSS, and other sources.
- **Metadata Engine:** Maps series to global IDs and manages covers/tags.
- **Download Handler:** Bridges with clients like qBittorrent, Aria2, or custom scrapers.
- **Post-Processor:** Handles image-to-archive conversion and library placement.

## Tech Stack
- **Backend:** Go (Golang)
- **Frontend:** Vue 3 + Vite + Tailwind CSS
- **Database:** SQLite
- **Communication:** REST API + WebSockets

## Status
Pre-alpha. Currently in the design and initialization phase.

ALTER TABLE indexers ADD COLUMN enable_rss INTEGER DEFAULT 1;
ALTER TABLE indexers ADD COLUMN enable_search INTEGER DEFAULT 1;
ALTER TABLE indexers ADD COLUMN additional_parameters TEXT;
ALTER TABLE indexers ADD COLUMN minimum_seeders INTEGER DEFAULT 1;
ALTER TABLE indexers ADD COLUMN seed_time INTEGER DEFAULT 0;
