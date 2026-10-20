package collections

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/fnune/kyaraben/internal/model"
)

const (
	DefaultPort        = 47410
	DefaultCommandPort = 47420
)

const statusMaxAge = time.Minute

type Server struct {
	service  *Service
	sessions *Sessions
	origins  map[string]bool
	port     int
	version  string
}

func NewServer(service *Service, sessions *Sessions, collections map[string]model.CollectionConfig, port int, version string) *Server {
	return &Server{
		service:  service,
		sessions: sessions,
		origins:  AllowedOrigins(collections),
		port:     port,
		version:  version,
	}
}

func AllowedOrigins(collections map[string]model.CollectionConfig) map[string]bool {
	origins := map[string]bool{}
	for _, c := range collections {
		u, err := url.Parse(c.URL)
		if err != nil || u.Host == "" {
			continue
		}
		origins[u.Scheme+"://"+u.Host] = true
	}
	return origins
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", s.health)
	mux.HandleFunc("GET /v1/collections", s.listCollections)
	mux.HandleFunc("GET /v1/collections/{collection}", s.collectionStatus)
	mux.HandleFunc("POST /v1/collections/{collection}/games/{game}/launch", s.launch)
	mux.HandleFunc("GET /v1/sessions/{session}", s.session)
	return s.guard(mux)
}

func (s *Server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.isLoopbackHost(r.Host) {
			writeError(w, http.StatusForbidden, "unexpected host, possibly DNS rebinding")
			return
		}
		origin := r.Header.Get("Origin")
		if origin != "" {
			if !s.origins[origin] {
				writeError(w, http.StatusForbidden, "origin not allowed: add its collection to kyaraben's config")
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "600")
			if r.Header.Get("Access-Control-Request-Private-Network") == "true" {
				w.Header().Set("Access-Control-Allow-Private-Network", "true")
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) isLoopbackHost(host string) bool {
	h, p, err := net.SplitHostPort(host)
	if err != nil || p != strconv.Itoa(s.port) {
		return false
	}
	return h == "127.0.0.1" || h == "localhost" || h == "::1"
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": s.version})
}

func (s *Server) listCollections(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string][]string{"collections": s.service.IDs()})
}

func (s *Server) collectionStatus(w http.ResponseWriter, r *http.Request) {
	st, err := s.service.Status(r.Context(), r.PathValue("collection"), statusMaxAge)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *Server) launch(w http.ResponseWriter, r *http.Request) {
	if err := s.sessions.CheckIdle(); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	opts := LaunchOptions{OpenMenu: r.URL.Query().Get("menu") == "true"}
	plan, err := s.service.Prepare(ctx, r.PathValue("collection"), r.PathValue("game"), opts)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	session, err := s.sessions.Launch(plan)
	if errors.Is(err, ErrGameRunning) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, session)
}

func (s *Server) session(w http.ResponseWriter, r *http.Request) {
	session, ok := s.sessions.Get(r.PathValue("session"))
	if !ok {
		writeError(w, http.StatusNotFound, "unknown session")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
