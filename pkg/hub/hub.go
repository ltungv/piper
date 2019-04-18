// TODO: Add authentication to server

package hub

import (
	"crypto/rsa"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Hub manages subscribed clients and message broadcast
type Hub struct {
	jwtSign     *rsa.PrivateKey
	jwtVerify   *rsa.PublicKey
	wsClients   map[*WSClient]bool
	instances   map[string]uint8
	subscribe   chan *WSClient
	unsubscribe chan *WSClient
	broadcast   chan *packet
}

// packet define the format of a message sent by the server
type packet struct {
	Time int64       `json:"time"`
	Data interface{} `json:"data"`
}

const broadcastBufSize = 4096
const maxInstances = 2

// upgrader upgrades normal HTTP connection to a WebSocket
var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // accepts connections from anyone
}

// New returns a broadcasting hub
func New(signKey, verifyKey []byte) *Hub {
	jwtSign, err := jwt.ParseRSAPrivateKeyFromPEM(signKey)
	if err != nil {
		log.Fatalf("could not parse private key; got %v", err)
	}

	jwtVerify, err := jwt.ParseRSAPublicKeyFromPEM(verifyKey)
	if err != nil {
		log.Fatalf("could not parse public key; got %v", err)
	}

	return &Hub{
		jwtSign:     jwtSign,
		jwtVerify:   jwtVerify,
		wsClients:   make(map[*WSClient]bool),
		instances:   make(map[string]uint8),
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
			for wsClient := range h.wsClients {
				select {
				case wsClient.send <- p:
				// disconnect client immediately after buffer is full
				default:
					log.Errorf("send channel buffer overload; client %v", wsClient)
					h.unsubscribe <- wsClient
				}
			}
		}
	}
}

// unsub deletes client from map
func (h *Hub) unsub(wsClient *WSClient) {
	if _, ok := h.wsClients[wsClient]; ok {
		instances := h.instances[wsClient.username]
		if instances == 1 {
			delete(h.instances, wsClient.username)
		} else {
			h.instances[wsClient.username]--
		}
		delete(h.wsClients, wsClient)
		close(wsClient.send)
		log.Infof("client unsubscribed: %v with %v instances", wsClient.username, instances-1)
	}
}

// sub add new client to map
func (h *Hub) sub(wsClient *WSClient) {
	instances, ok := h.instances[wsClient.username]
	if !ok {
		h.instances[wsClient.username] = 1
		h.wsClients[wsClient] = true
		log.Infof("client subscribed: %v with %v instances", wsClient.username, 1)
	} else {
		if instances < maxInstances {
			h.instances[wsClient.username]++
			h.wsClients[wsClient] = true
			log.Infof("client subscribed: %v with %v instances", wsClient.username, instances+1)
		} else {
			log.Infof("user %v max instances reached %v", wsClient.username, instances)
			_ = wsClient.wsConn.Close()
		}
	}
}
