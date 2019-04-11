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

// fieldObject defines attributes of object on the map
type fieldObject struct {
	Name      string `json:"name"`
	Position  [2]int `json:"pos"`
	Dimension [2]int `json:"dim"`
}

// packet define the format of a message sent by the server
type packet struct {
	Time     time.Time     `json:"time"`
	FieldObj []fieldObject `json:"objects"`
}

var (
	// upgrader upgrades normal HTTP connection to a WebSocket
	upgrader = websocket.Upgrader{
		WriteBufferSize: 1200,
		ReadBufferSize:  1200,
		CheckOrigin:     func(r *http.Request) bool { return true }, // accepts connections from anyone
	}
)

// New returns a broadcasting hub
func New() *Hub {
	return &Hub{
		wsClients:   make(map[*WSClient]bool),
		subscribe:   make(chan *WSClient),
		unsubscribe: make(chan *WSClient),
	}
}

// Run starts hub client manager
func (h *Hub) Run() {
	done := make(chan bool)
	go clientManager(done)
	<-done
}

// RunScript starts a script and broadcast its output
func (h *Hub) RunScript(prog string, scriptPath string) {
	cmd := exec.Command(prog, scriptPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("could not start script: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("could not start script: %v", err)
	}

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
