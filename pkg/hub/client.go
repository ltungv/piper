package hub

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const maxMessageSize = 1024
const writeWait = 1 * time.Second
const pongWait = 10 * time.Second
const pingPeriod = (pongWait * 9) / 10

// WSClient stores the queued messages and websocket information
type WSClient struct {
	username string
	h        *Hub
	wsConn   *websocket.Conn
	send     chan *packet
}

func init() {
	log.SetLevel(log.DebugLevel)
}

func (c *WSClient) writePipe() {
	// ticker to manage ping period
	ticker := time.NewTicker(pingPeriod)

	// close connection if error occur
	defer func() {
		ticker.Stop()
		c.h.unsubscribe <- c
	}()

	for {
		select {
		// sending message to client
		case p, ok := <-c.send:
			if err := c.wsConn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Errorf("could not set write deadline; got %v", err)
				return
			}
			if !ok {
				_ = c.wsConn.WriteMessage(websocket.CloseMessage, nil)
				return
			}

			msgBuf, err := json.Marshal(p)
			if err != nil {
				log.Errorf("could not marshal json: got %v", err)
			}
			if err := c.wsConn.WriteMessage(websocket.BinaryMessage, msgBuf); err != nil {
				log.Errorf("could not write packet: got %v", err)
				return
			}
		// periodically ping client and disconnect if cannot ping
		case <-ticker.C:
			if err := c.wsConn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Errorf("could not set write deadline; got %v", err)
				return
			}
			if err := c.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Errorf("could not ping client; got %v", err)
				return
			}
		}
	}
}

func (c *WSClient) readPipe() {
	// unsubsibe client on connection close
	defer func() {
		c.h.unsubscribe <- c
		_ = c.wsConn.Close()
	}()

	c.wsConn.SetReadLimit(maxMessageSize)

	err := c.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log.Errorf("could not set read deadline; got %v", err)
	}

	// extend read deadline when client reponse ping
	c.wsConn.SetPongHandler(func(string) error {
		err := c.wsConn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			log.Errorf("could not set read deadline; got %v", err)
		}
		log.Infof("pong received from %v", c.username)
		return nil
	})

	// create log on client ping
	c.wsConn.SetPingHandler(func(data string) error {
		deadline := time.Now().Add(writeWait)
		err := c.wsConn.WriteControl(websocket.PongMessage, []byte(data), deadline)
		if err != nil {
			return err
		}
		log.Infof("ping received from %v", c.username)
		return nil
	})

	for {
		// receive message from client
		_, _, err := c.wsConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error(err)
			}
			break
		}
	}
}
