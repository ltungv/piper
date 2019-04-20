package hub

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const maxMessageSize = 1024
const writeWait = 1 * time.Second
const pongWait = 60 * time.Second
const pingPeriod = (pongWait * 8) / 10
const maxMsgPerSec = 35

// WSClient stores the queued messages and websocket information
type WSClient struct {
	username string
	nMsgRead uint8
	free     bool
	h        *Hub
	wsConn   *websocket.Conn
	send     chan *packet
	sync.RWMutex
}

type confirmPacket struct {
	Fini bool `json:"finished"`
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
			deadline := time.Now().Add(writeWait)
			if !ok {
				_ = c.wsConn.WriteControl(websocket.CloseMessage, nil, deadline)
				log.Infof("send channel closed")
				return
			}

			if err := c.wsConn.SetWriteDeadline(deadline); err != nil {
				log.Errorf("could not set write deadline; got %v", err)
				return
			}

			msgBuf, err := json.Marshal(p)
			if err != nil {
				log.Errorf("could not marshal json: got %v", err)
				continue
			}
			if err := c.wsConn.WriteMessage(websocket.BinaryMessage, msgBuf); err != nil {
				log.Errorf("could not write packet: got %v", err)
				return
			}
			c.Lock()
			c.free = false
			c.Unlock()
		// periodically ping client and disconnect if cannot ping
		case <-ticker.C:
			deadline := time.Now().Add(writeWait)
			if err := c.wsConn.WriteControl(websocket.PingMessage, nil, deadline); err != nil {
				log.Errorf("could not ping client; got %v", err)
				return
			}
		}
	}
}

func (c *WSClient) readPipe() {
	ticker := time.NewTicker(time.Second)

	// unsubsibe client on connection close
	defer func() {
		c.h.unsubscribe <- c
		_ = c.wsConn.Close()
	}()

	c.wsConn.SetReadLimit(maxMessageSize)

	err := c.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log.Errorf("could not set read deadline; got %v", err)
		return
	}

	// extend read deadline when client reponse ping
	c.wsConn.SetPongHandler(func(string) error {
		err := c.wsConn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			log.Errorf("could not set read deadline; got %v", err)
			return err
		}
		log.Infof("pong received from %v", c.username)
		return nil
	})

	// create log on client ping
	c.wsConn.SetPingHandler(func(data string) error {
		deadline := time.Now().Add(writeWait)
		err := c.wsConn.WriteControl(websocket.PongMessage, []byte(data), deadline)
		if err != nil {
			log.Errorf("could not pong; got %v", err)
			return err
		}

		log.Infof("ping received from %v", c.username)
		return nil
	})

	for {
		select {
		case <-ticker.C:
			c.nMsgRead = 0
		default:
			if c.nMsgRead > maxMsgPerSec {
				log.Error("read rate limit exceeded")
				return
			}

			// receive message from client
			_, msg, err := c.wsConn.ReadMessage()
			if err != nil {
				log.Errorf("could not read; got %v", err)
				return
			}
			c.nMsgRead++

			packet := &confirmPacket{}
			if err := json.Unmarshal(msg, packet); err != nil {
				log.Errorf("could not parse confirm message; got %v", err)
				continue
			}
			if packet.Fini {
				c.Lock()
				c.free = true
				c.Unlock()
			}
		}
	}
}
