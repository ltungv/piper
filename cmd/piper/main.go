// TODO: Use TLS to encrypt connection

package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/letung3105/piper/pkg/hub"
)

func main() {
	// Receiving command line args
	binary := flag.String("b", "python", "name of interpreter")
	script := flag.String("s", "", "name of script")
	port := flag.String("p", "8000", "port use")
	flag.Parse()

	// Create and start broadcasting hub
	h := hub.New()
	go h.Run()

	// Run script and broadcast it output
	go h.BroadcastScript(*binary, *script)

	// Routing for HTTP connection
	router := http.NewServeMux()
	router.Handle("/ws", h)

	log.Printf("serving on port %s", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), router))
}
