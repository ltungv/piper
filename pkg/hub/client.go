package hub

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	maxMessageSize = 1024
	writeWait      = 5 * time.Second
	pongWait       = 10 * time.Second
	pingPeriod     = (pongWait * 9) / 10
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

func (c *WSClient) readPipe() {
	var pingCount int64 = 1
	var totalLat time.Duration
	avgPingTick := time.NewTicker(time.Second)
	defer func() {
		c.h.unsubscribe <- c
		_ = c.wsConn.Close()
	}()
	c.wsConn.SetReadLimit(maxMessageSize)
	_ = c.wsConn.SetReadDeadline(time.Now().Add(pongWait))

	c.wsConn.SetPingHandler(func(sentTime string) error {
		t, err := time.Parse(time.RFC3339, sentTime)
		if err != nil {
			return fmt.Errorf("could not parse time: %v", err)
		}
		pingCount++
		totalLat += time.Since(t) / 2
		select {
		case <-avgPingTick.C:
			log.Debugf("n: %v total: %v avg: %v", pingCount, totalLat, totalLat/time.Duration(pingCount))
			totalLat = 0
			pingCount = 1
		default:
		}
		return nil
	})
	c.wsConn.SetPongHandler(func(string) error {
		_ = c.wsConn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, _, err := c.wsConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

func (c *WSClient) writePipe() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.wsConn.Close()
	}()

	// TODO: Change write message type from TextMessage to BinaryMessage to deacrease bandwidth
	for {
		select {
		case p, ok := <-c.send:
			if !ok {
				_ = c.wsConn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.wsConn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Errorf("could not set write deadline; got %v", err)
				return
			}
			if err := c.wsConn.WriteJSON(p); err != nil {
				log.Errorf("could not write packet: %v", err)
				return
			}
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
