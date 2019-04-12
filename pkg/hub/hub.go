// TODO: Add authentication to server

package hub

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type (
	// Hub manages subscribed clients and message broadcast
	Hub struct {
		wsClients   map[*WSClient]bool
		subscribe   chan *WSClient
		unsubscribe chan *WSClient
		broadcast   chan *packet
		sync.RWMutex
	}

	// packet define the format of a message sent by the server
	packet struct {
		Time time.Time `json:"time"`
		Data []byte    `json:"data"`
	}
)

const (
	broadcastBufSize = 4096
)

var (
	// upgrader upgrades normal HTTP connection to a WebSocket
	upgrader = websocket.Upgrader{
		WriteBufferSize: 1280,
		ReadBufferSize:  1280,
		CheckOrigin:     func(r *http.Request) bool { return true }, // accepts connections from anyone
	}
)

// New returns a broadcasting hub
func New() *Hub {
	return &Hub{
		wsClients:   make(map[*WSClient]bool),
		subscribe:   make(chan *WSClient),
		unsubscribe: make(chan *WSClient),
		broadcast:   make(chan *packet, broadcastBufSize),
	}
}

// Run starts hub client manager
func (h *Hub) Run() {
	for {
		select {
		case wsClient := <-h.subscribe:
			h.sub(wsClient)
		case wsClient := <-h.unsubscribe:
			h.unsub(wsClient)
		case p := <-h.broadcast:
			h.RLock()
			for wsClient := range h.wsClients {
				select {
				case wsClient.send <- p:
				// disconnect client immediately after buffer is full
				default:
					log.Errorf("send channel buffer overload; client %v", wsClient)
					h.unsubscribe <- wsClient
				}
			}
			h.RUnlock()
		}
	}
}

// unsub deletes client from map
func (h *Hub) unsub(wsClient *WSClient) {
	h.Lock()
	defer h.Unlock()
	if _, ok := h.wsClients[wsClient]; ok {
		delete(h.wsClients, wsClient)
		close(wsClient.send)
		log.Infof("client unsubscribed: %v", wsClient)
	}
}

// sub add new client to map
func (h *Hub) sub(wsClient *WSClient) {
	h.Lock()
	defer h.Unlock()
	h.wsClients[wsClient] = true
	log.Infof("client subscribed: %v", wsClient)
}
