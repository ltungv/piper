package hub

import (
	"github.com/labstack/gommon/log"
)

func (h *Hub) clientManager(done chan bool) {
	defer func() {
		done <- true
	}()

	for {
		select {
		case wsClient := <-h.subscribe:
			h.sub(wsClient)
		case wsClient := <-h.unsubscribe:
			h.unsub(wsClient)
		}
	}
}

func (h *Hub) unsub(wsClient *WSClient) {
	h.Lock()
	defer h.Unlock()
	if _, ok := h.wsClients[wsClient]; ok {
		delete(h.wsClients, wsClient)
		close(wsClient.send)
		log.Infof("client unsubscribed: %v", wsClient)
	}
}

func (h *Hub) sub(wsClient *WSClient) {
	h.Lock()
	defer h.Unlock()
	h.wsClients[wsClient] = true
	log.Infof("client subscribed: %v", wsClient)
}
