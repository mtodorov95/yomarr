package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/mtodorov95/yomarr/internal/api"
	"github.com/mtodorov95/yomarr/internal/db"
	"github.com/mtodorov95/yomarr/internal/metadata"
	"github.com/mtodorov95/yomarr/internal/sync"
)

//go:embed all:web/dist
var webAssets embed.FS

func main() {
	// DB
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "/data/yomarr.db"
	}
	db.Init(path)

	// Server
	mux := http.NewServeMux()
	client := &http.Client{}
	// Metadata
	anilist := &metadata.AnilistProvider{Client: client}
	mangadex := &metadata.MangaDexProvider{Client: client}
	aggregator := metadata.NewAggregatorMetadataProvider(mangadex, anilist)
	// Sync
	syncEngine := sync.NewMangaDexSyncEngine(&db.SQLiteChapterStore{}, mangadex)

	// API routes
	mux.HandleFunc("/api/health", api.HealthHandler)

	seriesHandler := &api.SeriesHandler{
		Store:      &db.SQLiteSeriesStore{},
		Metadata:   aggregator,
		SyncEngine: syncEngine,
	}

	mux.HandleFunc("/api/series", seriesHandler.HandleSeries)

	chapterHandler := &api.ChapterHandler{Store: &db.SQLiteChapterStore{}}
	mux.HandleFunc("/api/chapters", chapterHandler.HandleChapters)

	indexerHandler := &api.IndexerHandler{Store: &db.SQLiteIndexerStore{}}
	mux.HandleFunc("/api/indexers", indexerHandler.HandleIndexers)

	// Static
	dist, err := fs.Sub(webAssets, "web/dist")
	if err != nil {
		log.Fatal(err)
	}

	fileServer := http.FileServer(http.FS(dist))
	mux.Handle("/", fileServer)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
