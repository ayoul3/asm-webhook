package server

import (
	"fmt"
	"html"
	"net/http"

	log "github.com/sirupsen/logrus"
	kwhhttp "github.com/slok/kubewebhook/v2/pkg/http"
	kwhlog "github.com/slok/kubewebhook/v2/pkg/log"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Debug("Received request at root ...")
	fmt.Fprintf(w, "hello %q", html.EscapeString(r.URL.Path))
}

func handlerFor(mutator kwhmutating.MutatorFunc, logger kwhlog.Logger) http.Handler {
	webhook, err := kwhmutating.NewWebhook(kwhmutating.WebhookConfig{
		ID:      "asm-webhook",
		Mutator: mutator,
		Logger:  logger,
	})
	if err != nil {
		logger.Errorf("error creating webhook: %s", err)
	}

	return kwhhttp.MustHandlerFor(kwhhttp.HandlerConfig{Webhook: webhook, Logger: logger})
}
