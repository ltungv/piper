package hub

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

// BroadcastScript starts a script and broadcast its output
func (h *Hub) BroadcastScript(interpreter, script string) {
	for {
		cmd := exec.Command(interpreter, script)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Errorf("could not get stdout script: %v", err)
		}

		// start script
		if err := cmd.Start(); err != nil {
			log.Errorf("could not start script: %v", err)
		}

		log.Info("Start broadcasting script output")
		// create new goroutine for reading script's output
		go func() {
			scanner := bufio.NewScanner(stdout)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				message := scanner.Bytes()
				if h.isBroadcasting {
					var data []map[string]interface{}

					err := json.Unmarshal(message, &data)
					if err != nil {
						log.Errorf("failed to parse json; got %v", err)
						continue
					}

					h.broadcast <- &packet{time.Now().UnixNano(), data}
				}
			}
		}()

		// wait for script to finish
		if err := cmd.Wait(); err != nil {
			log.Errorf("could not wait script: %v", err)
		}
	}
}
