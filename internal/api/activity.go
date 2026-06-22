package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mtodorov95/yomarr/internal/sync"
)

type EventHandler struct {
	Hub *sync.EventHub
}

func NewEventHandler(hub *sync.EventHub) *EventHandler {
	return &EventHandler{Hub: hub}
}

func (h *EventHandler) HandleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	clientChan := h.Hub.Register()
	defer h.Hub.Unregister(clientChan)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "[Activity] Streaming unsupported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-r.Context().Done():
			log.Println("[Activity] Frontend client disconnected")
			return

		case event, ok := <-clientChan:
			if !ok {
				return
			}

			payload, err := json.Marshal(event)
			if err != nil {
				continue
			}

			_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()

		case <-ticker.C:
			_, _ = fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
