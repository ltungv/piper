package hub

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("could not upgrade websocket connection; got %v", err)
		return
	}
	wsClient := &WSClient{h: h, wsConn: wsConn, send: make(chan *packet, 4096)}
	h.subscribe <- wsClient
	go wsClient.readPipe()
	go wsClient.writePipe()
}
