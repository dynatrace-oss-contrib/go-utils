package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	port       uint
	readyCheck ReadyCheck
}

func NewHealthHandler(opts ...HealthHandlerOption) *HealthHandler {
	h := &HealthHandler{
		port:       defaultHealthCheckPort,
		readyCheck: func() bool { return true },
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

type HealthHandlerOption func(handler *HealthHandler)

func WithReadyCheck(fn func() bool) HealthHandlerOption {
	return func(handler *HealthHandler) {
		handler.readyCheck = fn
	}
}

func WithPort(port uint) HealthHandlerOption {
	return func(handler *HealthHandler) {
		handler.port = port
	}
}

type ReadyCheck func() bool

func TryGETReadyCheck(urls ...string) func() bool {
	return func() bool {
		for _, url := range urls {
			_, err := http.Get(url)
			if err != nil {
				return false
			}
		}
		return true
	}
}

func (h *HealthHandler) Run(ctx context.Context) {
	http.HandleFunc("/ready", h.readyHandler)
	http.HandleFunc("/health", h.healthHandler)

	go func() {
		server := &http.Server{Addr: fmt.Sprintf(":%d", h.port), Handler: nil}
		err := server.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
		go func() {
			<-ctx.Done()
			server.Shutdown(ctx)
		}()
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
	if h.readyCheck() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
