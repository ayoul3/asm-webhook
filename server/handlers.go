package server

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	kwhhttp "github.com/slok/kubewebhook/v2/pkg/http"
	kwhlog "github.com/slok/kubewebhook/v2/pkg/log"
	kwhprometheus "github.com/slok/kubewebhook/v2/pkg/metrics/prometheus"
	kwhwebhook "github.com/slok/kubewebhook/v2/pkg/webhook"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Debug("Received request at root ...")
	fmt.Fprint(w, "Welcome to asm-webhook mutator")
}

func handlerFor(mutator kwhmutating.MutatorFunc, logger kwhlog.Logger, metricsRec *kwhprometheus.Recorder) http.Handler {
	webhook, err := kwhmutating.NewWebhook(kwhmutating.WebhookConfig{
		ID:      "asm-webhook",
		Mutator: mutator,
		Logger:  logger,
	})
	if err != nil {
		logger.Errorf("error creating webhook: %s", err)
	}
	webhook = kwhwebhook.NewMeasuredWebhook(metricsRec, webhook)

	return kwhhttp.MustHandlerFor(kwhhttp.HandlerConfig{Webhook: webhook, Logger: logger})
}
