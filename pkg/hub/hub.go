package hub

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Hub manages subscribed inputs and outputs
type Hub struct {
	wsClients   map[*WSClient]bool
	subscribe   chan *WSClient
	unsubscribe chan *WSClient
	sync.RWMutex
}

type fieldObject struct {
	Name      string `json:"name"`
	Position  [2]int `json:"pos"`
	Dimension [2]int `json:"dim"`
}

type packet struct {
	Time     time.Time     `json:"time"`
	FieldObj []fieldObject `json:"objects"`
}

var (
	upgrader = websocket.Upgrader{
		WriteBufferSize: 1200,
		ReadBufferSize:  1200,
		CheckOrigin:     func(r *http.Request) bool { return true }, // accepts connections from anyone
	}
)

// New returns a new pub-sub hub
func New() *Hub {
	return &Hub{
		wsClients:   make(map[*WSClient]bool),
		subscribe:   make(chan *WSClient),
		unsubscribe: make(chan *WSClient),
	}
}

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

// Run starts hub public subscriber service
func (h *Hub) Run(prog string, scriptPath string) {
	cmd := exec.Command(prog, scriptPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("could not start script: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("could not start script: %v", err)
	}

	go h.clientManager()

	go h.broadcast(stdout)

	if err := cmd.Wait(); err != nil {
		log.Fatalf("could not wait script: %v", err)
	}
}

// Reading output of python script and broadcast it to all connected client
func (h *Hub) broadcast(cmdOutput io.ReadCloser) {
	scanner := bufio.NewScanner(cmdOutput)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var objs []fieldObject
		if err := json.Unmarshal([]byte(scanner.Text()), &objs); err != nil {
			log.Errorf("could not parse JSON: %v", err)
			continue
		}

		h.RLock()
		for wsClient := range h.wsClients {
			select {
			case wsClient.send <- &packet{time.Now(), objs}:
			default:
				log.Errorf("send channel buffer overload; client %v", wsClient)
				h.unsubscribe <- wsClient
			}
		}
		h.RUnlock()
	}
}

func (h *Hub) clientManager() {
	for {
		select {
		case wsClient := <-h.subscribe:
			h.sub(wsClient)
		case wsClient := <-h.unsubscribe:
			h.unsub(wsClient)
		}
	}
}

func (h *Hub) unsub(wsClient *WSClient) {
	h.Lock()
	defer h.Unlock()
	if _, ok := h.wsClients[wsClient]; ok {
		delete(h.wsClients, wsClient)
		close(wsClient.send)
		log.Infof("client unsubscribed: %v", wsClient)
	}
}

func (h *Hub) sub(wsClient *WSClient) {
	h.Lock()
	defer h.Unlock()
	h.wsClients[wsClient] = true
	log.Infof("client subscribed: %v", wsClient)
}
