// API server implementation. One route per resource family and /status endpoints
// Exposes endpoints for Pods, nodes and deployments
// Translates HTTP+JSON into types Go objects
// Reads/writes to store

package apiserver

import (
	"context"
	"net/http"
	"sync/atomic"

	"github.com/therandombyte/mini-k8s/pkg/config"
	"github.com/therandombyte/mini-k8s/pkg/store"
)

type Server struct {
	cfg config.APIConfig
	st store.Store
	rv atomic.Int64
}

func New(cfg config.APIConfig, st store.Store) *Server {
	return &Server{cfg: cfg, st: st}
}

func (s *Server) nextRV() int64 {
	return s.rv.Add(1)
}

// Struct as a method receiver
// Run() is bound to Server type via the receiver (s *Server)
// so call it as s.Run()
// Run() is a function with special first paramater (the receiver)
// Multiplexer is a HTTP request router (implements http.Handler) that maps URL paths to handler functions
// you register routes to it, then pass it to http.ListenAndServe
func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_,_ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/api/v1/pods", s.handlePods)


	httpServer := &http.Server{Addr: s.cfg.Address, Handler: mux}

	// background go-routine that waits for the context to be cancelled and gracefully shuts down the server
	// launches an anonymous func in a new goroutine, so it runs concurrently
	go func() {
		<-ctx.Done()
		_ = httpServer.Shutdown(context.Background())
	}()

	return httpServer.ListenAndServe()
}

func (s *Server) handlePods(w http.ResponseWriter, r *http.Request) {
	
}
