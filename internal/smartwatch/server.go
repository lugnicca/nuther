package smartwatch

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

type Server struct {
	store *Store
	hub   *EventHub
}

func NewServer(store *Store, hub *EventHub) *Server {
	return &Server{store: store, hub: hub}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", s.hub.ServeSSE)
	mux.HandleFunc("/snapshots", s.handleSnapshots)
	mux.HandleFunc("/snapshots/", s.handleSnapshot)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}` + "\n"))
	})
	return mux
}

func Listen(addr string) (net.Listener, string, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", err
	}
	host, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		_ = ln.Close()
		return nil, "", err
	}
	if host == "::" || host == "0.0.0.0" || host == "" {
		host = "127.0.0.1"
	}
	return ln, "http://" + net.JoinHostPort(host, port), nil
}

func (s *Server) handleSnapshots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	index, err := s.store.LoadIndex()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, index)
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/snapshots/")
	id = strings.Trim(id, "/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	snapshot, err := s.store.ReadSnapshot(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	writeJSONResponse(w, snapshot)
}

func writeJSONResponse(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(value)
}
