package hub

import (
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	// maxMessageSize = 1024
	writeWait  = 1 * time.Second
	readWait   = 10 * time.Second
	pingPeriod = (readWait * 9) / 10
)

// WSClient stores the queued messages and websocket information
type WSClient struct {
	h      *Hub
	wsConn *websocket.Conn
	send   chan *packet
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
		_ = c.wsConn.Close()
		c.h.unsubscribe <- c
	}()

	for {
		if err := c.wsConn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
			log.Errorf("could not set write deadline; got %v", err)
			return
		}

		select {
		// sending message to client
		case p, ok := <-c.send:
			if !ok {
				_ = c.wsConn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.wsConn.WriteJSON(p); err != nil {
				log.Errorf("could not write packet: %v", err)
				return
			}
		// periodically ping client and disconnect if cannot ping
		case <-ticker.C:
			if err := c.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Errorf("could not ping client; got %v", err)
				return
			}
		}
	}
}

// func (c *WSClient) readPipe() {
// 	// unsubsibe client on connection close
// 	defer func() {
// 		c.h.unsubscribe <- c
// 		_ = c.wsConn.Close()
// 	}()
//
// 	c.wsConn.SetReadLimit(maxMessageSize)
// 	_ = c.wsConn.SetReadDeadline(time.Now().Add(readWait))
//
// 	for {
// 		// receive message from client
// 		_, message, err := c.wsConn.ReadMessage()
// 		if err != nil {
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 				log.Printf("error: %v", err)
// 			}
// 			break
// 		}
// 		c.h.broadcast <- &packet{time.Now(), message}
// 	}
// }
