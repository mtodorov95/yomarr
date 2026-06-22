package sync

import (
	"sync"
	"github.com/mtodorov95/yomarr/internal/models"
)

type EventHub struct {
	mu          sync.RWMutex
	clients     map[chan models.QueueEvent]bool
}

func NewEventHub() *EventHub {
	return &EventHub{
		clients: make(map[chan models.QueueEvent]bool),
	}
}

func (h *EventHub) Register() chan models.QueueEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	ch := make(chan models.QueueEvent, 10)
	h.clients[ch] = true
	return ch
}

func (h *EventHub) Unregister(ch chan models.QueueEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if _, exists := h.clients[ch]; exists {
		delete(h.clients, ch)
		close(ch)
	}
}

func (h *EventHub) Broadcast(event models.QueueEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for ch := range h.clients {
		select {
		case ch <- event:
		default:
			// 
		}
	}
}
