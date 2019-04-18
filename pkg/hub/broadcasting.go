package hub

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"time"

	"github.com/labstack/gommon/log"
)

// BroadcastScript starts a script and broadcast its output
func (h *Hub) BroadcastScript(prog string, script string) {
	// define script and binary used
	cmd := exec.Command(prog, script)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("could not start script: %v", err)
	}

	// start script
	if err := cmd.Start(); err != nil {
		log.Fatalf("could not start script: %v", err)
	}

	// create new goroutine for reading script's output
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			message := scanner.Bytes()
			var data []map[string]interface{}

			err := json.Unmarshal(message, &data)
			if err != nil {
				log.Errorf("failed to parse json; got %v", err)
				continue
			}

			h.broadcast <- &packet{time.Now().UnixNano(), data}
		}
	}()

	// wait for script to finish
	if err := cmd.Wait(); err != nil {
		log.Fatalf("could not wait script: %v", err)
	}
}
