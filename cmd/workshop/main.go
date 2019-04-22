package main

import (
	"flag"
	"log"
	"net/http"
)

var mux = http.NewServeMux()

func main() {
	dir := flag.String("dir", "./", "Served directory")
	flag.Parse()

	mux.HandleFunc("/pass", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		n, err := w.Write([]byte("<h1>rec@12345</h1>"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Print(err)
			return
		}
		log.Printf("write %d bytes", n)
	})
	mux.Handle("/", http.FileServer(http.Dir(*dir)))

	log.Print("Server running on port 80")
	log.Fatal(http.ListenAndServe(":80", mux))
}
