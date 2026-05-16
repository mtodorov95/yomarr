package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/mtodorov95/yomarr/internal/api"
)

//go:embed all:web/dist
var webAssets embed.FS

func main() {
	mux := http.NewServeMux()
	// API routes
	mux.HandleFunc("/api/health", api.HealthHandler)

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
