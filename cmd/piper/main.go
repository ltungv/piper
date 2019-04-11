package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/letung3105/piper/pkg/hub"
)

func main() {
	prog := flag.String("prog", "python", "name of interpreter")
	script := flag.String("script", "", "name of script")
	port := flag.String("port", "8000", "port use")
	flag.Parse()

	h := hub.New()
	go h.Run(*prog, *script)

	router := http.NewServeMux()
	router.Handle("/ws", h)

	log.Printf("serving on port %s", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), router))
}
