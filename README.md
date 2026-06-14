# Yomarr 📚

**Yomarr** is an automated manga collection manager and tracker (an "Arr" equivalent for Manga). It monitors your favorite series for new chapters, handles tracking and downloads via your preferred client, and organizes them into a clean, standardized local library.

---

## ⚠️ Disclaimer & Project Status

Yomarr is now in **Beta**. 

While the core dashboard, database schema, metadata sync loops, and basic automation pipelines are fully functional and stable enough for daily deployment, this is still an active passion project. 

* **Development Pace:** This project is worked on entirely in my free time. Updates will happen whenever I have availability.
* **The 1.0 Milestone:** Version `1.0` will be declared once **every core goal listed below is fully implemented**. Features developed after the 1.0 milestone will be considered "nice-to-haves" and are not an immediate priority.
* **Contributions:** Public contributions, bug reports, and pull requests are highly welcome! Please check the limitations below before opening major feature requests.

---

## 🚀 Current Supported Features vs. Roadmap

To keep expectations transparent for early users, here is exactly what Yomarr handles right now versus what is currently planned:

### 🟢 What Works Right Now
* **Supported Indexers:** Nyaa.si.
* **Supported Download Clients:** qBittorrent.
* **Tracked Languages:** English (`EN`) and `RAW` only.

### 🟡 What is Planned (Roadmap to 1.0)
* **Post-Processor Engine:** Automated raw image validation, custom directory packing into `.cbz` zip formats, and inline `ComicInfo.xml` metadata injection for instant compatibility with external readers like Kavita or Komga.
* **Expanded Metadata Aggregation:** Deep secondary metadata mapping profiles sourcing tags, structural descriptions, and tracking links from AniList, MyAnimeList, and MangaDex.
* **Flexibility:** Support for additional download clients (e.g., Aria2) and generic RSS indexing frames.

---

## 🛠️ Tech Stack

* **Backend:** Go (Golang)
* **Frontend:** Vue 3 (Vite + Vanilla CSS)
* **Database:** SQLite
* **Communication:** REST API

---

## 📦 Getting Started

The easiest way to host Yomarr is using Docker Compose. 

### 1. Prerequisite Setup
Ensure the directories you intend to mount exist on your host system and have correct permissions. Yomarr runs as a non-root user (`1000:1000`) by default to maintain strict host security guidelines.

### 2. Docker Compose Configuration
Create a `docker-compose.yml` file in your desired deployment directory and use the official image structure below:

```yaml
services:
  yomarr:
    image: mariotodorov95/yomarr:latest
    container_name: yomarr
    user: "1000:1000"
    ports:
      - "9191:9191"
    volumes:
      - ./data:/data
      - /path/to/your/manga/library:/Manga
      - /path/to/your/torrent/downloads:/downloads
    restart: unless-stopped
