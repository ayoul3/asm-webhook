package server

import (
	"net/http"
	"time"

	"github.com/ayoul3/asm-webhook/mutate"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	kwhlog "github.com/slok/kubewebhook/v2/pkg/log/logrus"
	kwhprometheus "github.com/slok/kubewebhook/v2/pkg/metrics/prometheus"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	"github.com/spf13/afero"
)

func Start(tlsCrt, tlsKey string) {
	mux := http.NewServeMux()
	logger := kwhlog.NewLogrus(log.WithField("app", "asm-webhook"))
	mutatorClient, err := mutate.CreateClient(afero.NewOsFs())
	if err != nil {
		log.Fatalf("Cannot create mutator client: %s", err)
	}
	mutator := kwhmutating.MutatorFunc(mutatorClient.SecretsMutator)
	metricsRec, err := kwhprometheus.NewRecorder(kwhprometheus.RecorderConfig{Registry: prometheus.DefaultRegisterer})

	podHandler := handlerFor(mutator, logger, metricsRec)

	mux.HandleFunc("/", handleRoot)
	mux.Handle("/mutate", podHandler)
	mux.Handle("/metrics", promhttp.Handler())

	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}
	log.Fatal(s.ListenAndServeTLS(tlsCrt, tlsKey))
}
