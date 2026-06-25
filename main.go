package main

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mtodorov95/yomarr/internal/api"
	"github.com/mtodorov95/yomarr/internal/config"
	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/metadata"
	"github.com/mtodorov95/yomarr/internal/sync"
)

//go:embed all:web/dist
var webAssets embed.FS

var AppVersion = "v0.0.0-dev"

func main() {
	// Env
	config.LoadEnv()
	port := os.Getenv("PORT")
	if port == "" {
		port = "9191"
	}

	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "/data/yomarr.db"
	}

	// DB
	database := db.Init(path)
	defer database.Close()

	chapterStore := &db.SQLiteChapterStore{DB: database}
	seriesStore := &db.SQLiteSeriesStore{DB: database}
	indexerStore := &db.SQLiteIndexerStore{DB: database}
	clientStore := &db.SQLiteDownloadClientStore{DB: database}
	queueStore := &db.SQLiteQueueStore{DB: database}
	// Server
	mux := http.NewServeMux()
	client := &http.Client{}
	// Metadata
	anilist := &metadata.AnilistProvider{Client: client}
	mangadex := &metadata.MangaDexProvider{Client: client}
	aggregator := metadata.NewAggregatorMetadataProvider(mangadex, anilist)
	eventHub := sync.NewEventHub()
	// Sync
	manager := sync.NewDynamicManager(indexerStore, clientStore, queueStore)
	syncEngine := sync.NewMangaDexSyncEngine(chapterStore, mangadex)
	searchEngine := sync.NewSearchEngine(chapterStore, seriesStore, manager, manager)
	rssEngine := sync.NewRssEngine(chapterStore, seriesStore, manager)
	monitor := sync.NewDownloadMonitor(chapterStore, seriesStore, manager, queueStore, eventHub)

	scanner := sync.NewLibraryScanner(chapterStore, seriesStore, aggregator, syncEngine)
	if err := scanner.ScanLibrary(); err != nil {
		log.Printf("[Scanner] Initial library boot scan failed: %v", err)
	}

	// Scheduled tasks
	// Download client monitor
	monitor.Start()
	// Local scan
	scanner.StartBackgroundScan(6 * time.Hour)
	// Metadata refresh
	scanner.StartBackgroundMetadataRefresher(12 * time.Hour)
	// RSS feed
	rssEngine.StartBackgroundRssCheck(15 * time.Minute)
	// Missing chapter search
	searchEngine.StartBackgroundSearcher(72 * time.Hour)

	// API routes
	healthHandler := api.NewHealthHandler(AppVersion)
	mux.HandleFunc("/api/health", healthHandler.HandleHealth)

	seriesHandler := &api.SeriesHandler{
		Store:      seriesStore,
		Metadata:   aggregator,
		SyncEngine: syncEngine,
		Scanner:    scanner,
	}

	mux.HandleFunc("/api/series", seriesHandler.HandleSeries)

	chapterHandler := &api.ChapterHandler{Store: chapterStore}
	mux.HandleFunc("/api/chapters", chapterHandler.HandleChapters)

	indexerHandler := &api.IndexerHandler{Store: indexerStore}
	mux.HandleFunc("/api/indexers", indexerHandler.HandleIndexers)

	downloadClientHandler := &api.DownloadClientHandler{Store: clientStore}
	mux.HandleFunc("/api/download-clients", downloadClientHandler.HandleDownloadClients)

	searchHandler := api.NewSearchHandler(searchEngine)
	mux.HandleFunc("/api/series/search-missing", searchHandler.SearchMissing)

	libraryHandler := api.NewLibraryHandler(scanner)
	mux.HandleFunc("/api/library/scan", libraryHandler.ScanLibrary)

	systemHandler := api.NewSystemHandler(seriesStore, chapterStore)
	mux.HandleFunc("/api/system/stats", systemHandler.HandleSystemStats)

	activityHandler := api.NewEventHandler(eventHub)
	mux.HandleFunc("/api/activity", activityHandler.HandleStream)

	// Assets
	mux.HandleFunc("/api/proxy-cover", api.ProxyCoverHandler)
	mux.HandleFunc("/api/assets", api.FileAssetHandler)

	// Static
	dist, err := fs.Sub(webAssets, "web/dist")
	if err != nil {
		log.Fatal(err)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		filePath := r.URL.Path
		if filePath == "" || filePath == "/" {
			filePath = "index.html"
		} else if filePath[0] == '/' {
			filePath = filePath[1:]
		}

		file, err := dist.Open(filePath)
		if err == nil {
			file.Close()
			http.FileServer(http.FS(dist)).ServeHTTP(w, r)
			return
		}

		indexFile, err := dist.Open("index.html")
		if err != nil {
			http.Error(w, "[Server] index.html not found in embedded assets", http.StatusInternalServerError)
			return
		}
		defer indexFile.Close()

		stat, err := indexFile.Stat()
		if err != nil {
			http.Error(w, "[Server] Failed to read index.html info", http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile.(io.ReadSeeker))
	})

	log.Printf("[Server] Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("[Server] Server failed to start: %v", err)
	}
}
