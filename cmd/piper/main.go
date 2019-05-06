package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/letung3105/piper/pkg/hub"
)

func main() {
	// JWT keys
	jwtSignPath := flag.String("priv", "./keys/jwt/rsa.key", "JWT private signing key")
	jwtVerifyPath := flag.String("pub", "./keys/jwt/rsa.pub", "JWT public verify key")

	// clients credentials
	usersCreds := flag.String("users", "./.users.json", "Users' credentials")

	// SSL certificates
	ca := flag.String("ca", "./keys/certs/pub/cacert.pem", "CA")
	crt := flag.String("crt", "./keys/certs/pub/servercert.pem", "server certificate")
	key := flag.String("key", "./keys/certs/priv/serverkey.pem", "server key")

	// interpreter binary
	binary := flag.String("i", "python", "name of interpreter")
	// script file
	script := flag.String("f", "", "name of script file")
	// network port
	port := flag.String("p", "443", "port to listen")
	flag.Parse()

	signKey, verifyKey, err := getJWTKeys(*jwtSignPath, *jwtVerifyPath)
	if err != nil {
		log.Fatalf("could not parse jwt keys; got %v", err)
	}

	var users map[string]*hub.UserInfo
	creds, err := ioutil.ReadFile(*usersCreds)
	if err != nil {
		log.Fatalf("could not read users file; got %v", err)
	}

	if err := json.Unmarshal(creds, &users); err != nil {
		log.Fatalf("could not parse users creds; got %v", err)
	}

	// Create and start broadcasting hub
	h := hub.New(users, signKey, verifyKey, *binary, *script)
	go h.Run()

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
	}).Methods("GET")
	mux.Handle("/data", h.JWTProtect("contestant")(h.ServeHTTP))
	mux.Handle("/control", h.JWTProtect("admin")(h.Control())).Methods("POST")
	mux.HandleFunc("/subscribe", h.Subscribe).Methods("POST")

	cfg, err := createServerConfig(*ca, *crt, *key)
	if err != nil {
		log.Fatalf("could not create config; got %v", err)
	}

	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"POST", "GET", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "X-Requested-With", "Authorization"}),
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", *port),
		Handler:      cors(mux),
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Printf("serving on port %s", *port)

	log.Fatal(srv.ListenAndServeTLS(*crt, *key))
	// log.Fatal(srv.ListenAndServe())
}
