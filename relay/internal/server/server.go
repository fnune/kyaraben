package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fnune/kyaraben/relay/internal/pairing"
)

type Config struct {
	Addr string
	TTL  time.Duration
}

func DefaultConfig() Config {
	return Config{
		Addr: ":8080",
		TTL:  pairing.DefaultTTL,
	}
}

type Server struct {
	cfg          Config
	store        *pairing.Store
	handlers     *Handlers
	server       *http.Server
	rateLimiters []*RateLimiter
}

func New(cfg Config) *Server {
	store := pairing.NewStore(cfg.TTL)
	handlers := NewHandlers(store)

	mux := http.NewServeMux()

	createLimiter := NewRateLimiter(10, time.Minute)
	getLimiter := NewRateLimiter(30, time.Minute)
	deleteLimiter := NewRateLimiter(10, time.Minute)
	submitLimiter := NewRateLimiter(10, time.Minute)
	pollLimiter := NewRateLimiter(60, time.Minute)

	mux.Handle("POST /pair", createLimiter.Middleware(http.HandlerFunc(handlers.CreateSession)))
	mux.Handle("GET /pair/{code}", getLimiter.Middleware(http.HandlerFunc(handlers.GetSession)))
	mux.Handle("DELETE /pair/{code}", deleteLimiter.Middleware(http.HandlerFunc(handlers.DeleteSession)))
	mux.Handle("POST /pair/{code}/response", submitLimiter.Middleware(http.HandlerFunc(handlers.SubmitResponse)))
	mux.Handle("GET /pair/{code}/response", pollLimiter.Middleware(http.HandlerFunc(handlers.GetResponse)))
	mux.HandleFunc("GET /health", handlers.Health)

	return &Server{
		cfg:          cfg,
		store:        store,
		handlers:     handlers,
		rateLimiters: []*RateLimiter{createLimiter, getLimiter, deleteLimiter, submitLimiter, pollLimiter},
		server: &http.Server{
			Addr:           cfg.Addr,
			Handler:        loggingMiddleware(corsMiddleware(mux)),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 16,
		},
	}
}

func (s *Server) Start() error {
	log.Printf("Starting relay server on %s", s.cfg.Addr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		log.Printf("Received signal %v, shutting down", sig)
		return s.Shutdown(context.Background())
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Printf("Shutting down gracefully (active sessions: %d)", s.store.Len())

	for _, rl := range s.rateLimiters {
		rl.Close()
	}
	s.store.Close()

	return s.server.Shutdown(shutdownCtx)
}

func (s *Server) Close() error {
	for _, rl := range s.rateLimiters {
		rl.Close()
	}
	s.store.Close()
	return s.server.Close()
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start).Round(time.Millisecond))
	})
}
