package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fnune/kyaraben/relay/internal/pairing"
)

func TestHandlers_CreateSession(t *testing.T) {
	store := pairing.NewStore(time.Minute)
	defer store.Close()
	h := NewHandlers(store)

	body := `{"deviceId": "AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA"}`
	req := httptest.NewRequest(http.MethodPost, "/pair", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateSession(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var resp CreateSessionResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if len(resp.Code) != pairing.CodeLength {
		t.Errorf("expected code length %d, got %d", pairing.CodeLength, len(resp.Code))
	}
	if resp.ExpiresIn != int(pairing.DefaultTTL.Seconds()) {
		t.Errorf("expected expiresIn %d, got %d", int(pairing.DefaultTTL.Seconds()), resp.ExpiresIn)
	}
}

func TestHandlers_CreateSession_InvalidDeviceID(t *testing.T) {
	store := pairing.NewStore(time.Minute)
	defer store.Close()
	h := NewHandlers(store)

	body := `{"deviceId": "invalid"}`
	req := httptest.NewRequest(http.MethodPost, "/pair", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.CreateSession(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlers_GetSession(t *testing.T) {
	store := pairing.NewStore(time.Minute)
	defer store.Close()
	h := NewHandlers(store)

	deviceID := "AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA"
	session, _ := store.Create(deviceID, "127.0.0.1")

	req := httptest.NewRequest(http.MethodGet, "/pair/"+session.Code, nil)
	w := httptest.NewRecorder()

	h.GetSession(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp GetSessionResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp.DeviceID != deviceID {
		t.Errorf("expected device ID %s, got %s", deviceID, resp.DeviceID)
	}
}

func TestHandlers_GetSession_NotFound(t *testing.T) {
	store := pairing.NewStore(time.Minute)
	defer store.Close()
	h := NewHandlers(store)

	req := httptest.NewRequest(http.MethodGet, "/pair/NOTFND", nil)
	w := httptest.NewRecorder()

	h.GetSession(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandlers_SubmitResponse(t *testing.T) {
	store := pairing.NewStore(time.Minute)
	defer store.Close()
	h := NewHandlers(store)

	session, _ := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "127.0.0.1")

	responderID := "BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB"
	body := `{"deviceId": "` + responderID + `"}`
	req := httptest.NewRequest(http.MethodPost, "/pair/"+session.Code+"/response", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.SubmitResponse(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d: %s", http.StatusNoContent, w.Code, w.Body.String())
	}

	updated, _ := store.Get(session.Code)
	if updated.ResponderDeviceID != responderID {
		t.Errorf("expected responder device ID %s, got %s", responderID, updated.ResponderDeviceID)
	}
}

func TestHandlers_GetResponse_NotReady(t *testing.T) {
	store := pairing.NewStore(time.Minute)
	defer store.Close()
	h := NewHandlers(store)

	session, _ := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "127.0.0.1")

	req := httptest.NewRequest(http.MethodGet, "/pair/"+session.Code+"/response", nil)
	w := httptest.NewRecorder()

	h.GetResponse(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp GetResponseResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp.Ready {
		t.Error("expected ready to be false")
	}
	if resp.DeviceID != "" {
		t.Errorf("expected empty device ID, got %s", resp.DeviceID)
	}
}

func TestHandlers_GetResponse_Ready(t *testing.T) {
	store := pairing.NewStore(time.Minute)
	defer store.Close()
	h := NewHandlers(store)

	session, _ := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "127.0.0.1")
	responderID := "BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB"
	_ = store.SetResponse(session.Code, responderID)

	req := httptest.NewRequest(http.MethodGet, "/pair/"+session.Code+"/response", nil)
	w := httptest.NewRecorder()

	h.GetResponse(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp GetResponseResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Ready {
		t.Error("expected ready to be true")
	}
	if resp.DeviceID != responderID {
		t.Errorf("expected device ID %s, got %s", responderID, resp.DeviceID)
	}
}

func TestHandlers_DeleteSession(t *testing.T) {
	store := pairing.NewStore(time.Minute)
	defer store.Close()
	h := NewHandlers(store)

	session, _ := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "127.0.0.1")

	req := httptest.NewRequest(http.MethodDelete, "/pair/"+session.Code, nil)
	w := httptest.NewRecorder()

	h.DeleteSession(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	_, err := store.Get(session.Code)
	if err != pairing.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestHandlers_Health(t *testing.T) {
	store := pairing.NewStore(time.Minute)
	defer store.Close()
	h := NewHandlers(store)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestIsValidDeviceID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", true},
		{"1234567-1234567-1234567-1234567-1234567-1234567-1234567-1234567", true},
		{"ABC1234-DEF5678-GHI9012-JKL3456-MNO7890-PQR1234-STU5678-VWX9012", true},
		{"invalid", false},
		{"AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", false},
		{"AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-EXTRA", false},
		{"aaaaaaa-aaaaaaa-aaaaaaa-aaaaaaa-aaaaaaa-aaaaaaa-aaaaaaa-aaaaaaa", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := isValidDeviceID(tt.id); got != tt.valid {
				t.Errorf("isValidDeviceID(%q) = %v, want %v", tt.id, got, tt.valid)
			}
		})
	}
}
