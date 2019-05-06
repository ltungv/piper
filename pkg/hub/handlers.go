package hub

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// ClientCredentials stores client login information
type ClientCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string
}

// Token stores user jwt
type Token struct {
	Token string `json:"token"`
}

const clientBufSize = 1024

// ServeHTTP handles upgrading and maintaining websocket connection with client
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(usernameKey).(string)

	// update client connection to websocket
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("could not upgrade websocket connection; got %v", err)
		return
	}

	// create and subscribe new client
	wsClient := &WSClient{
		username: username,
		nMsgRead: 0,
		free:     true,
		h:        h,
		wsConn:   wsConn,
		send:     make(chan *packet, clientBufSize),
	}
	h.subscribe <- wsClient

	// start reading messages from client and send to broadcast
	go wsClient.readPipe()

	// start writing messages from broadcast channel to client
	go wsClient.writePipe()
}

// Subscribe validates user credentials and sends back a token
func (h *Hub) Subscribe(w http.ResponseWriter, r *http.Request) {
	// get client username and password from http body
	creds := &ClientCredentials{}
	if err := json.NewDecoder(r.Body).Decode(creds); err != nil {
		log.Errorf("could not parse credentials; got %v", err)
		return
	}

	// validate for username and password
	user, ok := h.users[creds.Username]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Infof("invalid user's credentials")
		return
	}
	if user.Password != creds.Password {
		w.WriteHeader(http.StatusBadRequest)
		log.Infof("invalid user's credentials")
		return
	}

	creds.Role = user.Role

	// sign new jwt for client
	token, err := newJWTToken(h.jwtSign, creds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("could not create token; got %v", err)
		return
	}

	httpWriteJSON(w, &Token{token})
}

// Control starts and stops script from running
func (h *Hub) Control() http.HandlerFunc {
	type request struct {
		Action string `json:"action"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			log.Errorf("could not parse request; got %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch req.Action {
		case "start":
			h.Lock()
			h.isBroadcasting = true
			h.Unlock()
			log.Info("Start broadcasting")
			w.WriteHeader(http.StatusOK)
		case "stop":
			h.Lock()
			if h.isBroadcasting {
				h.isBroadcasting = false
			}
			h.Unlock()
			log.Info("Stop broadcasting")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusBadRequest)
			log.Errorf("invalid action")
		}
	}
}
