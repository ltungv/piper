package hub

import (
	"encoding/json"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

// ClientCredentials stores client login information
type ClientCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Token stores user jwt
type Token struct {
	Token string `json:"token"`
}

const clientBufSize = 1024

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

// ServeHTTP handles upgrading and maintaining websocket connection with client
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get jwt from header
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Infof("no authorization found")
		return
	}

	reqToken := strings.Split(auth, "Bearer")[1]
	reqToken = strings.TrimSpace(reqToken)
	if reqToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Infof("no token found")
		return
	}

	// parse and validate token
	token, err := jwt.Parse(reqToken, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return h.jwtVerify, nil
	})

	// check for token error
	switch err.(type) {
	case nil: // no error
		if !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			log.Infof("invalid token")
			return
		}
	case *jwt.ValidationError:
		vErr := err.(*jwt.ValidationError)
		switch vErr.Errors {
		case jwt.ValidationErrorExpired:
			w.WriteHeader(http.StatusUnauthorized)
			log.Infof("token expired")
			return

		default:
			w.WriteHeader(http.StatusInternalServerError)
			log.Errorf("could not parse token; got %v", vErr)
			return
		}
	default: // something else went wrong
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("could not parse token; got %v", err)
		return
	}

	// parse claims to get client username
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("could not parse claims")
		return
	}

	username, ok := claims["username"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Infof("no username found in jwt claims")
		return
	}
	usernameStr := username.(string)

	// update client connection to websocket
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("could not upgrade websocket connection; got %v", err)
		return
	}

	// create and subscribe new client
	wsClient := &WSClient{
		username: usernameStr,
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
	expectedPassword, ok := users[creds.Username]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Infof("invalid user's credentials")
		return
	}
	if expectedPassword != creds.Password {
		w.WriteHeader(http.StatusBadRequest)
		log.Infof("invalid user's credentials")
		return
	}

	// sign new jwt for client
	token, err := newJWTToken(h.jwtSign, creds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("could not create token; got %v", err)
		return
	}

	httpWriteJSON(w, &Token{token})
}
