package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/fnune/kyaraben/relay/internal/pairing"
)

var deviceIDPattern = regexp.MustCompile(`^[A-Z0-9]{7}(-[A-Z0-9]{7}){7}$`)

type CreateSessionRequest struct {
	DeviceID string `json:"deviceId"`
}

type CreateSessionResponse struct {
	Code      string `json:"code"`
	ExpiresIn int    `json:"expiresIn"`
}

type GetSessionResponse struct {
	DeviceID string `json:"deviceId"`
}

type SubmitResponseRequest struct {
	DeviceID string `json:"deviceId"`
}

type GetResponseResponse struct {
	DeviceID string `json:"deviceId,omitempty"`
	Ready    bool   `json:"ready"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Handlers struct {
	store *pairing.Store
}

func NewHandlers(store *pairing.Store) *Handlers {
	return &Handlers{store: store}
}

func (h *Handlers) CreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	deviceID := normalizeDeviceID(req.DeviceID)
	if !isValidDeviceID(deviceID) {
		writeError(w, "invalid device ID format", http.StatusBadRequest)
		return
	}

	session, err := h.store.Create(deviceID, getClientIP(r))
	if err != nil {
		if errors.Is(err, pairing.ErrTooManySessions) || errors.Is(err, pairing.ErrTooManySessionsForIP) {
			log.Printf("Rate limit hit for IP %s", getClientIP(r))
			writeError(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		writeError(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	log.Printf("Session created: code=%s (active sessions: %d)", session.Code, h.store.Len())

	writeJSON(w, CreateSessionResponse{
		Code:      session.Code,
		ExpiresIn: int(pairing.DefaultTTL.Seconds()),
	}, http.StatusCreated)
}

func (h *Handlers) GetSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := extractCode(r.URL.Path)
	if code == "" {
		writeError(w, "missing code", http.StatusBadRequest)
		return
	}

	session, err := h.store.Get(code)
	if err != nil {
		if errors.Is(err, pairing.ErrSessionNotFound) {
			writeError(w, "session not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, pairing.ErrSessionExpired) {
			writeError(w, "session expired", http.StatusGone)
			return
		}
		writeError(w, "failed to get session", http.StatusInternalServerError)
		return
	}

	writeJSON(w, GetSessionResponse{
		DeviceID: session.PrimaryDeviceID,
	}, http.StatusOK)
}

func (h *Handlers) SubmitResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := extractCodeFromResponsePath(r.URL.Path)
	if code == "" {
		writeError(w, "missing code", http.StatusBadRequest)
		return
	}

	var req SubmitResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	deviceID := normalizeDeviceID(req.DeviceID)
	if !isValidDeviceID(deviceID) {
		writeError(w, "invalid device ID format", http.StatusBadRequest)
		return
	}

	err := h.store.SetResponse(code, deviceID)
	if err != nil {
		if errors.Is(err, pairing.ErrSessionNotFound) {
			writeError(w, "session not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, pairing.ErrSessionExpired) {
			writeError(w, "session expired", http.StatusGone)
			return
		}
		if errors.Is(err, pairing.ErrResponseAlreadySet) {
			writeError(w, "response already submitted", http.StatusConflict)
			return
		}
		writeError(w, "failed to submit response", http.StatusInternalServerError)
		return
	}

	log.Printf("Pairing complete: code=%s", code)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) GetResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := extractCodeFromResponsePath(r.URL.Path)
	if code == "" {
		writeError(w, "missing code", http.StatusBadRequest)
		return
	}

	session, err := h.store.Get(code)
	if err != nil {
		if errors.Is(err, pairing.ErrSessionNotFound) {
			writeError(w, "session not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, pairing.ErrSessionExpired) {
			writeError(w, "session expired", http.StatusGone)
			return
		}
		writeError(w, "failed to get session", http.StatusInternalServerError)
		return
	}

	writeJSON(w, GetResponseResponse{
		DeviceID: session.SecondaryDeviceID,
		Ready:    session.HasResponse(),
	}, http.StatusOK)
}

func (h *Handlers) DeleteSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := extractCode(r.URL.Path)
	if code == "" {
		writeError(w, "missing code", http.StatusBadRequest)
		return
	}

	err := h.store.Delete(code)
	if err != nil {
		if errors.Is(err, pairing.ErrSessionNotFound) {
			writeError(w, "session not found", http.StatusNotFound)
			return
		}
		writeError(w, "failed to delete session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func extractCode(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "pair" {
		return parts[1]
	}
	return ""
}

func extractCodeFromResponsePath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "pair" && parts[2] == "response" {
		return parts[1]
	}
	return ""
}

func normalizeDeviceID(id string) string {
	return strings.ToUpper(strings.TrimSpace(id))
}

func isValidDeviceID(id string) bool {
	return deviceIDPattern.MatchString(id)
}

func writeJSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, ErrorResponse{Error: message}, status)
}
