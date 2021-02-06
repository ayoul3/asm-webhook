package server

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"

	m "github.com/ayoul3/ssm-webhook/mutate"
	log "github.com/sirupsen/logrus"
)

func sendError(err error, w http.ResponseWriter) {
	log.Warnf("Error mutating request %s", err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s", err)
}

func handleMutate(w http.ResponseWriter, r *http.Request) {
	var mutated []byte

	log.Println("Received mutate request ...")
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		sendError(err, w)
		return
	}

	// mutate the request
	log.Info("Calling mutate method ...")
	if mutated, err = m.Mutate(body); err != nil {
		sendError(err, w)
		return
	}

	// and write it back
	w.WriteHeader(http.StatusOK)
	w.Write(mutated)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Debug("Received request at root ...")
	fmt.Fprintf(w, "hello %q", html.EscapeString(r.URL.Path))
}
