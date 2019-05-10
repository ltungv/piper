// TODO: Add authentication to server
// TODO: remove client buffer

package hub

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"sync"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// UserInfo stores role and password
type UserInfo struct {
	Password string `json:"password"`
	Role     string `json:"role"`
}

// Hub manages subscribed clients and message broadcast
type Hub struct {
	jwtSign        *rsa.PrivateKey        // private rsa key for signing jwt
	jwtVerify      *rsa.PublicKey         // public rsa key for verifying jwt
	isBroadcasting bool                   // check if server if broadcasting messages
	users          map[string]*UserInfo   // users credentials for subscribing
	instances      map[string]uint8       // limit number for instances per client
	wsClients      map[*WSClient]struct{} // manage subscribed client websocket connection
	subscribe      chan *WSClient         // clients queue for subscription
	unsubscribe    chan *WSClient         // clients queue for unsubscription
	broadcast      chan *packet           // messages queue for broadcasting
	sync.RWMutex
}

// packet define the format of a message sent by the server
type packet struct {
	Time int64       `json:"time"` // unix nano timestamp
	Data interface{} `json:"data"` // data to be sent
}

const broadcastBufSize = 4096 // messages queue size
const maxInstances = 3        // maximum instances per client

// upgrader upgrades normal HTTP connection to a WebSocket
var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // accepts connections from anyone
}

// New returns a broadcasting hub
func New(users map[string]*UserInfo, signKey, verifyKey []byte) *Hub {
	fmt.Println(users)
	// load private rsa key
	jwtSign, err := jwt.ParseRSAPrivateKeyFromPEM(signKey)
	if err != nil {
		log.Fatalf("could not parse private key; got %v", err)
	}

	// load public rsa key
	jwtVerify, err := jwt.ParseRSAPublicKeyFromPEM(verifyKey)
	if err != nil {
		log.Fatalf("could not parse public key; got %v", err)
	}

	return &Hub{
		jwtSign:        jwtSign,
		jwtVerify:      jwtVerify,
		isBroadcasting: false,
		users:          users,
		wsClients:      make(map[*WSClient]struct{}),
		instances:      make(map[string]uint8),
		subscribe:      make(chan *WSClient),
		unsubscribe:    make(chan *WSClient),
		broadcast:      make(chan *packet, broadcastBufSize),
	}
}

// Run starts hub client manager and messages broadcasting
func (h *Hub) Run() {
	go func() {
		for {
			select {
			// subscribe client websocket connection
			case wsClient := <-h.subscribe:
				h.sub(wsClient)
			// unsubscribe client websocket connection
			case wsClient := <-h.unsubscribe:
				h.unsub(wsClient)
			}
		}
	}()

	go func() {
		// send message from queue to all subscribed client
		for {
			p := <-h.broadcast
			h.RLock()
			for wsClient := range h.wsClients {
				wsClient.RLock()
				if wsClient.free {
					select {
					case wsClient.send <- p:
					// disconnect client immediately after buffer is full
					default:
						log.Errorf("send channel buffer overload; client %v", wsClient)
						h.unsubscribe <- wsClient
					}
				}
				wsClient.RUnlock()
			}
			h.RUnlock()
		}
	}()
}

// unsub deletes client from map
func (h *Hub) unsub(wsClient *WSClient) {
	h.Lock()
	defer h.Unlock()
	// check if client was subscribed
	if _, ok := h.wsClients[wsClient]; ok {
		// decrease number of connected instances
		// delete key if instances is decreased to 0
		instances := h.instances[wsClient.username]
		if instances == 1 {
			delete(h.instances, wsClient.username)
		} else {
			h.instances[wsClient.username]--
		}

		// unsubscribe and close send channel
		delete(h.wsClients, wsClient)
		close(wsClient.send)
		log.Infof("client unsubscribed: %v with %v instances", wsClient.username, instances-1)
	}
}

// sub add new client to map
func (h *Hub) sub(wsClient *WSClient) {
	h.Lock()
	defer h.Unlock()
	// check for current connected instances
	instances, ok := h.instances[wsClient.username]

	// create new key if not existed
	// or increase number of instances if existed
	// client websocket connection is not registered if max instances reached
	if !ok {
		h.instances[wsClient.username] = 1
		h.wsClients[wsClient] = struct{}{}
		log.Infof("client subscribed: %v with %v instances", wsClient.username, 1)
	} else {
		if instances < maxInstances {
			h.instances[wsClient.username]++
			h.wsClients[wsClient] = struct{}{}
			log.Infof("client subscribed: %v with %v instances", wsClient.username, instances+1)
		} else {
			log.Infof("user %v max instances reached %v", wsClient.username, instances)
			_ = wsClient.wsConn.Close()
		}
	}
}
