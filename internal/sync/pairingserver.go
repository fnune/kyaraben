package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	gosync "sync"
	"time"
)

type PairingRequest struct {
	Code     string `json:"code"`
	DeviceID string `json:"deviceId"`
	Name     string `json:"name"`
	Mode     string `json:"mode,omitempty"`
}

type PairingResponse struct {
	DeviceID string `json:"deviceId"`
	Name     string `json:"name"`
}

const maxPairingAttempts = 5

type PairingServer struct {
	code         string
	localID      string
	localName    string
	listener     net.Listener
	server       *http.Server
	result       chan PairResult
	mu           gosync.Mutex
	attempts     int
	onPairAccept func(peerDeviceID, peerName string) error
}

func NewPairingServer(code, localID, localName string) *PairingServer {
	return &PairingServer{
		code:      code,
		localID:   localID,
		localName: localName,
		result:    make(chan PairResult, 1),
	}
}

func (s *PairingServer) SetOnPairAccept(fn func(peerDeviceID, peerName string) error) {
	s.onPairAccept = fn
}

func (s *PairingServer) Start(listener net.Listener) error {
	s.listener = listener
	mux := http.NewServeMux()
	mux.HandleFunc("/pair", s.handlePair)

	s.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Error("pairing server error: %v", err)
		}
	}()

	return nil
}

func (s *PairingServer) Port() int {
	if s.listener == nil {
		return 0
	}
	return s.listener.Addr().(*net.TCPAddr).Port
}

func (s *PairingServer) Result() <-chan PairResult {
	return s.result
}

func (s *PairingServer) Stop() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.server.Shutdown(ctx)
	}
}

func (s *PairingServer) handlePair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	s.attempts++
	if s.attempts > maxPairingAttempts {
		s.mu.Unlock()
		http.Error(w, "too many attempts", http.StatusTooManyRequests)
		return
	}
	s.mu.Unlock()

	var req PairingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Code != s.code {
		http.Error(w, "invalid code", http.StatusForbidden)
		return
	}

	if req.Mode == "primary" {
		http.Error(w, "cannot pair two primary devices", http.StatusConflict)
		return
	}

	if req.DeviceID == "" {
		http.Error(w, "deviceId required", http.StatusBadRequest)
		return
	}

	if s.onPairAccept != nil {
		if err := s.onPairAccept(req.DeviceID, req.Name); err != nil {
			http.Error(w, fmt.Sprintf("pairing failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	resp := PairingResponse{
		DeviceID: s.localID,
		Name:     s.localName,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)

	select {
	case s.result <- PairResult{PeerDeviceID: req.DeviceID, PeerName: req.Name}:
	default:
	}
}
