package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/mtodorov95/yomarr/internal/api"
	"github.com/mtodorov95/yomarr/internal/config"
	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/download"
	"github.com/mtodorov95/yomarr/internal/indexer"
	"github.com/mtodorov95/yomarr/internal/metadata"
	"github.com/mtodorov95/yomarr/internal/sync"
)

//go:embed all:web/dist
var webAssets embed.FS

func main() {
	config.LoadEnv()
	qbitURL := os.Getenv("QBIT_URL")
	if qbitURL == "" { qbitURL = "http://127.0.0.1:8080" }
	
	qbitUser := os.Getenv("QBIT_USER")
	if qbitUser == "" { qbitUser = "admin" }
	
	qbitPass := os.Getenv("QBIT_PASS")
	if qbitPass == "" { qbitPass = "adminadmin" }

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	// DB
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "/data/yomarr.db"
	}
	db.Init(path)

	chapterStore := &db.SQLiteChapterStore{} 
	seriesStore := &db.SQLiteSeriesStore{} 
	// Server
	mux := http.NewServeMux()
	client := &http.Client{}
	// Metadata
	anilist := &metadata.AnilistProvider{Client: client}
	mangadex := &metadata.MangaDexProvider{Client: client}
	aggregator := metadata.NewAggregatorMetadataProvider(mangadex, anilist)
	// Indexer
	nyaaIndexer := indexer.NewNyaaIndexer()
	// Download
	qbClient, err := download.NewQBittorrentClient(qbitURL, qbitUser, qbitPass)
	if err != nil {
		log.Printf("Warning: Could not connect to qbittorrent client: %v", err)
	}
	// Sync
	syncEngine := sync.NewMangaDexSyncEngine(chapterStore, mangadex)
	nyaaEngine := sync.NewNyaaSyncEngine(chapterStore, seriesStore, nyaaIndexer, qbClient)

	// API routes
	mux.HandleFunc("/api/health", api.HealthHandler)

	seriesHandler := &api.SeriesHandler{
		Store:      seriesStore,
		Metadata:   aggregator,
		SyncEngine: syncEngine,
	}

	mux.HandleFunc("/api/series", seriesHandler.HandleSeries)

	chapterHandler := &api.ChapterHandler{Store: &db.SQLiteChapterStore{}}
	mux.HandleFunc("/api/chapters", chapterHandler.HandleChapters)

	indexerHandler := &api.IndexerHandler{Store: &db.SQLiteIndexerStore{}}
	mux.HandleFunc("/api/indexers", indexerHandler.HandleIndexers)

	searchHandler := api.NewSearchHandler(nyaaEngine)
	mux.HandleFunc("/api/series/search-missing", searchHandler.SearchMissing)

	// Static
	dist, err := fs.Sub(webAssets, "web/dist")
	if err != nil {
		log.Fatal(err)
	}

	fileServer := http.FileServer(http.FS(dist))
	mux.Handle("/", fileServer)

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
