package hub

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	clientBufSize = 2048
)

// ServeHTTP handles upgrading and maintaining connection with client
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("could not upgrade websocket connection; got %v", err)
		return
	}

	// create and subscribe new client
	wsClient := &WSClient{h: h, wsConn: wsConn, send: make(chan *packet, clientBufSize)}
	h.subscribe <- wsClient

	// go wsClient.readPipe()

	// start writing messages from broadcast channel to client
	go wsClient.writePipe()
}
