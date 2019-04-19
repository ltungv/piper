package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/letung3105/piper/pkg/hub"
)

func main() {
	// Receiving command line args
	jwtSignPath := flag.String("priv", "./keys/jwt/rsa.key", "JWT private signing key")
	jwtVerifyPath := flag.String("pub", "./keys/jwt/rsa.pub", "JWT public verify key")
	ca := flag.String("ca", "./keys/certs/pub/cacert.pem", "CA")
	crt := flag.String("crt", "./keys/certs/pub/servercert.pem", "server certificate")
	key := flag.String("key", "./keys/certs/priv/serverkey.pem", "server key")
	binary := flag.String("i", "python", "name of interpreter")
	script := flag.String("f", "", "name of script file")
	port := flag.String("p", "8000", "port to listen")
	flag.Parse()

	signKey, verifyKey, err := getJWTKeys(*jwtSignPath, *jwtVerifyPath)
	if err != nil {
		log.Fatalf("could not parse jwt keys; got %v", err)
	}

	// Create and start broadcasting hub
	h := hub.New(signKey, verifyKey)
	go h.Run()
	// Run script and broadcast it output
	go h.BroadcastScript(*binary, *script)

	// Routing for HTTP connection
	mux := mux.NewRouter()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		n, err := w.Write([]byte("VGU Robocon 2019 Broadcasting Server"))
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("write %d", n)
	})
	mux.Handle("/data", h)
	mux.HandleFunc("/subscribe", h.Subscribe).Methods("POST")

	cfg, err := createServerConfig(*ca, *crt, *key)
	if err != nil {
		log.Fatalf("could not create config; got %v", err)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", *port),
		Handler:      mux,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Printf("serving on port %s", *port)

	log.Fatal(srv.ListenAndServeTLS(*crt, *key))
	// log.Fatal(srv.ListenAndServe())
}
