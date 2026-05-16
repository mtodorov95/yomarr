package main

import (
	"log"
	"net/http"

	"github.com/mtodorov95/yomarr/internal/api"
)

func main() {
	http.HandleFunc("/health", api.HealthHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	} 
}
