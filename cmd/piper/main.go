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
	// binary command arguments
	jwtSignPath := flag.String("priv", "./keys/jwt/rsa.key", "JWT private signing key")
	jwtVerifyPath := flag.String("pub", "./keys/jwt/rsa.pub", "JWT public verify key")
	ca := flag.String("ca", "./keys/ca/cacert.pem", "CA")
	crt := flag.String("crt", "./keys/server/servercert.pem", "server certificate")
	key := flag.String("key", "./keys/server/serverkey.pem", "server key")
	usersCreds := flag.String("users", "./.creds.json", "Users' credentials")
	binary := flag.String("i", "python3", "name of interpreter")
	script := flag.String("f", "./scripts/main.py", "name of script file")
	backPort := flag.String("back-port", "4433", "backend port to listen")
	frontPort := flag.String("front-port", "3000", "frontend port to listen")
	frontPath := flag.String("front-path", "./frontend/dist", "frontend static files")
	flag.Parse()

	// jwt verification keys pair
	signKey, verifyKey, err := getJWTKeys(*jwtSignPath, *jwtVerifyPath)
	if err != nil {
		log.Fatalf("could not parse jwt keys; got %v", err)
	}
	// users login infomation
	creds, err := ioutil.ReadFile(*usersCreds)
	if err != nil {
		log.Fatalf("could not read users file; got %v", err)
	}
	var users map[string]*hub.UserInfo
	if err := json.Unmarshal(creds, &users); err != nil {
		log.Fatalf("could not parse users creds; got %v", err)
	}
	// server ssl configuration
	cfg, err := createServerConfig(*ca, *crt, *key)
	if err != nil {
		log.Fatalf("could not create config; got %v", err)
	}

	// Create and start broadcasting hub
	h := hub.New(users, signKey, verifyKey)
	go h.Run()
	go h.BroadcastScript(*binary, *script)

	// cross origin configuration
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"POST", "GET", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "X-Requested-With", "Authorization"}),
	)

	go func() {
		r := mux.NewRouter()
		r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(*frontPath))))

		srv := &http.Server{
			Addr:         fmt.Sprintf(":%s", *frontPort),
			Handler:      cors(r),
			TLSConfig:    cfg,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		}

		log.Printf("serving frontend on port %s", *frontPort)
		log.Fatal(srv.ListenAndServeTLS(*crt, *key))
	}()

	// Routing for HTTP connection
	r := mux.NewRouter()
	// Serve index page on all unhandled routes
	r.Handle("/data", h.JWTProtect("contestant")(h.ServeHTTP))
	r.Handle("/control", h.JWTProtect("admin")(h.Control())).Methods("POST")
	r.HandleFunc("/subscribe", h.Subscribe).Methods("POST")

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", *backPort),
		Handler:      cors(r),
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Printf("serving backend on port %s", *backPort)
	log.Fatal(srv.ListenAndServeTLS(*crt, *key))
}
