package api

import (
	"encoding/json"
	"fmt"
	"github.com/keptn/go-utils/pkg/common/retry"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type StatusBody struct {
	Status string `json:"status"`
}

// RunHealthEndpoint starts a http server that hosts the endpoint for the liveness probe
// Deprecated. User HealthHandler type instead
func RunHealthEndpoint(port string) {
	http.HandleFunc("/health", healthHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Println(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	status := StatusBody{Status: "OK"}

	body, _ := json.Marshal(status)

	w.Header().Set("content-type", "application/json")

	_, err := w.Write(body)
	if err != nil {
		log.Println(err)
	}
}

const defaultHealthCheckPort = 10998

type HealthHandler struct {
	ready *atomic.Value
	port  uint
}

func NewHealthHandler(opts ...HealthHandlerOption) *HealthHandler {
	r := &atomic.Value{}
	r.Store(0)
	h := &HealthHandler{
		ready: r,
		port:  defaultHealthCheckPort,
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

type HealthHandlerOption func(handler *HealthHandler)
type ReadyCheck func(handler *HealthHandler)

func TryGETReadyCheck(url string) func(handler *HealthHandler) {
	return func(handler *HealthHandler) {
		retry.Retry(func() error {
			_, err := http.Get(url)
			if err == nil {
				handler.Ready()
			}
			return err
		}, retry.DelayBetweenRetries(1*time.Second))
	}
}

func (h *HealthHandler) RunWithReadyCheck(readyCheck ReadyCheck) {
	http.HandleFunc("/ready", h.readyHandler)
	http.HandleFunc("/health", h.healthHandler)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", 10998), nil)
		if err != nil {
			log.Println(err)
		}
	}()

	go func() {
		readyCheck(h)
	}()
}

func (h *HealthHandler) healthHandler(w http.ResponseWriter, r *http.Request) {
	status := StatusBody{Status: "OK"}

	body, _ := json.Marshal(status)

	w.Header().Set("content-type", "application/json")

	_, err := w.Write(body)
	if err != nil {
		log.Println(err)
	}
}

func (h *HealthHandler) readyHandler(w http.ResponseWriter, r *http.Request) {
	if val, ok := h.ready.Load().(int); ok && val == 1 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func (h *HealthHandler) Ready() {
	h.ready.Store(1)
}
