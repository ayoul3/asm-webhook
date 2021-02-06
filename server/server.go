package server

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func Start() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/mutate", handleMutate)

	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}
	log.Fatal(s.ListenAndServeTLS("./ssl/server.crt", "./ssl/key.pem"))
}
