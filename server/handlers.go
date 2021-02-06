package server

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"

	m "github.com/ayoul3/ssm-webhook/mutate"
	log "github.com/sirupsen/logrus"
	kwhhttp "github.com/slok/kubewebhook/v2/pkg/http"
	kwhlog "github.com/slok/kubewebhook/v2/pkg/log"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
)

func sendError(err error, w http.ResponseWriter) {
	log.Warnf("Error mutating request %s", err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s", err)
}

func handleMutate(w http.ResponseWriter, r *http.Request) {
	var mutated, body []byte
	var err error

	log.Println("Received mutate request ...")
	if body, err = ioutil.ReadAll(r.Body); err != nil {
		sendError(err, w)
		return
	}
	defer r.Body.Close()

	if mutated, err = m.Mutate(body); err != nil {
		sendError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(mutated)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Debug("Received request at root ...")
	fmt.Fprintf(w, "hello %q", html.EscapeString(r.URL.Path))
}

func handlerFor(mutator kwhmutating.MutatorFunc, logger kwhlog.Logger) http.Handler {
	webhook, err := kwhmutating.NewWebhook(kwhmutating.WebhookConfig{
		ID:      "ssm-webhook",
		Mutator: mutator,
		Logger:  logger,
	})
	if err != nil {
		logger.Errorf("error creating webhook: %s", err)
	}

	return kwhhttp.MustHandlerFor(kwhhttp.HandlerConfig{Webhook: webhook, Logger: logger})
}
