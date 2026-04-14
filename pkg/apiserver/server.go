// API server implementation. One route per resource family and /status endpoints
// Exposes endpoints for Pods, nodes and deployments
// Translates HTTP+JSON into types Go objects
// Reads/writes to store

package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	v1 "github.com/therandombyte/mini-k8s/pkg/api/v1"
	"github.com/therandombyte/mini-k8s/pkg/config"
	"github.com/therandombyte/mini-k8s/pkg/store"
	"github.com/therandombyte/mini-k8s/pkg/util"
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
	mux.HandleFunc("/api/v1/pods/", s.handlePodsByName)

	mux.HandleFunc("/api/v1/nodes", s.handleNodes)
	mux.HandleFunc("/api/v1/nodes/", s.handleNodesByName)


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
	switch r.Method {
	case http.MethodGet:
		items, rv, err := s.st.List(r.Context(), "pods", "default")
		
		if err != nil {
			util.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		// api response being formed
		list := v1.PodList{}
		list.APIVersion = "mini-k8s/v1"
		list.Kind = "PodList"
		list.Metadata.ResourceVersion = rv

		for _, item := range items {
			// type assertion, treat item as *v1.Pod if it really is one
			// k8s objects are stored and returned as pointers to structs
			if pod, ok := item.(*v1.Pod); ok {
				list.Items = append(list.Items, *pod)
			}
		}
		util.WriteJSON(w, http.StatusOK, list)

	case http.MethodPost:
		var pod v1.Pod
		if err := decode(r, &pod); err != nil {
			util.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if pod.Metadata.Namespace == "" {
			pod.Metadata.Namespace = "default"
		}
		if pod.Metadata.CreationTimestamp.IsZero() {
			pod.Metadata.CreationTimestamp = time.Now()
		}
		pod.Metadata.ResourceVersion = s.nextRV()
		if pod.Status.Phase == "" {
			pod.Status.Phase = "Pending"
		}

		if err := s.st.Create(r.Context(), "pods", pod.Metadata.Namespace, pod.Metadata.Name, &pod); err != nil {
			util.WriteJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}

		util.WriteJSON(w, http.StatusCreated, pod)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	
	}
}

func decode[T any](r *http.Request, out *T) error {
	return json.NewDecoder(r.Body).Decode(out)
}

func (s *Server) handlePodsByName(w http.ResponseWriter, r *http.Request) {
	
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, rv, err := s.st.List(r.Context(), "nodes", "")
		if err != nil {
			util.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		list := v1.NodeList{}
		list.APIVersion = "mini-k8s/v1"
		list.Kind= "NodeList"
		list.Metadata.ResourceVersion = rv

		for _, item := range items {
			if node, ok := item.(*v1.Node); ok {
				list.Items = append(list.Items, *node)
			}
		}
		util.WriteJSON(w, http.StatusOK, list)
	case http.MethodPost:
		var node v1.Node
		if err := decode(r, &node); err != nil {
			util.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if node.Metadata.CreationTimestamp.IsZero() {
			node.Metadata.CreationTimestamp = time.Now()
		}
		node.Metadata.ResourceVersion = s.nextRV()

		if err := s.st.Create(r.Context(), "nodes", "", node.Metadata.Name, &node); err != nil {
			util.WriteJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		util.WriteJSON(w, http.StatusCreated, node)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleNodesByName(w http.ResponseWriter, r *http.Request) {
	
}
