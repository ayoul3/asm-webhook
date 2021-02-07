package server

import (
	"net/http"
	"time"

	"github.com/ayoul3/asm-webhook/mutate"
	log "github.com/sirupsen/logrus"
	kwhlog "github.com/slok/kubewebhook/v2/pkg/log/logrus"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
)

func Start() {
	mux := http.NewServeMux()
	logger := kwhlog.NewLogrus(log.WithField("app", "asm-webhook"))
	mutatorClient := mutate.CreateClient()

	mutator := kwhmutating.MutatorFunc(mutatorClient.SecretsMutator)
	podHandler := handlerFor(mutator, logger)

	mux.HandleFunc("/", handleRoot)
	mux.Handle("/mutate", podHandler)

	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}
	log.Fatal(s.ListenAndServeTLS("./ssl/server.crt", "./ssl/key.pem"))
}
