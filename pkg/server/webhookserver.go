package server

import (
	"crypto/tls"

	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"zerok-injector/internal/config"
	"zerok-injector/pkg/inject"
	"zerok-injector/pkg/storage"
)

type HttpApiHandler struct {
	injector *inject.Injector
}

func (h *HttpApiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	fmt.Printf("Got a request from webhook")

	if err != nil {
		webhookErrorResponse(err, w)
		return
	}

	response, err := h.injector.Inject(body)

	if err != nil {
		fmt.Printf("Error while injecting zk agent %v\n", err)
	}

	// Sending http status as OK, even when injection failed to not disturb the pods in cluster.
	w.WriteHeader(http.StatusOK)
	w.Write(response)

	r.Body.Close()
}

func webhookErrorResponse(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
}

func getMux(cfg config.WebhookConfig, runtimeMap *storage.ImageRuntimeHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle(cfg.Path, getHandler(runtimeMap))
	return mux
}

func getHandler(runtimeMap *storage.ImageRuntimeHandler) http.Handler {
	injectHandler := &HttpApiHandler{
		injector: &inject.Injector{ImageRuntimeHandler: runtimeMap},
	}
	return injectHandler
}

func StartWebHookServer(cfg config.WebhookConfig, serverPair tls.Certificate, runtimeMap *storage.ImageRuntimeHandler) {
	s := &http.Server{
		Addr:           ":8443",
		Handler:        getMux(cfg, runtimeMap),
		TLSConfig:      &tls.Config{Certificates: []tls.Certificate{serverPair}},
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServeTLS("", "")
}

func StartDebugWebHookServer(cfg config.WebhookConfig, runtimeMap *storage.ImageRuntimeHandler) {
	s := &http.Server{
		Addr:           ":8442",
		Handler:        getMux(cfg, runtimeMap),
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServe()
}
