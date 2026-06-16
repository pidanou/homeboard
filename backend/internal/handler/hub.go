package handler

import "sync"

// Hub broadcasts refresh signals to SSE clients grouped by family.
type Hub struct {
	mu      sync.Mutex
	clients map[string][]chan struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string][]chan struct{})}
}

func (h *Hub) Subscribe(familyID string) chan struct{} {
	ch := make(chan struct{}, 1)
	h.mu.Lock()
	h.clients[familyID] = append(h.clients[familyID], ch)
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(familyID string, ch chan struct{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, c := range h.clients[familyID] {
		if c == ch {
			h.clients[familyID] = append(h.clients[familyID][:i], h.clients[familyID][i+1:]...)
			return
		}
	}
}

func (h *Hub) Broadcast(familyID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range h.clients[familyID] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
