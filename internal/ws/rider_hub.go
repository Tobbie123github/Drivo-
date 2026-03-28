package ws

import (
	"sync"

	"github.com/google/uuid"
)

type RiderHub struct {
	clients    map[uuid.UUID]*RiderClient
	Register   chan *RiderClient
	Unregister chan *RiderClient
	mu         sync.RWMutex
}

type RiderClient struct {
	Hub     *RiderHub
	RiderID uuid.UUID
	Send    chan []byte
}

func NewRiderHub() *RiderHub {
	return &RiderHub{
		clients:    make(map[uuid.UUID]*RiderClient),
		Register:   make(chan *RiderClient),
		Unregister: make(chan *RiderClient),
	}
}

func (h *RiderHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.RiderID] = client
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.RiderID]; ok {
				delete(h.clients, client.RiderID)
				close(client.Send)
			}
			h.mu.Unlock()
		}
	}
}

func (h *RiderHub) SendToRider(riderID uuid.UUID, message []byte) bool {
	h.mu.RLock()
	client, ok := h.clients[riderID]
	h.mu.RUnlock()

	if !ok {
		return false
	}

	select {
	case client.Send <- message:
		return true
	default:
		h.mu.Lock()
		delete(h.clients, riderID)
		close(client.Send)
		h.mu.Unlock()
		return false
	}
}

func (h *RiderHub) IsOnline(riderID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[riderID]
	return ok
}

func (h *RiderHub) GetOnlineRiderIDs() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]uuid.UUID, 0, len(h.clients))
	for id := range h.clients {
		ids = append(ids, id)
	}
	return ids
}
