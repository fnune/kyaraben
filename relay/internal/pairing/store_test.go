package pairing

import (
	"testing"
	"time"
)

func TestStore_Create(t *testing.T) {
	store := NewStore(time.Minute)
	defer store.Close()

	session, err := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "192.168.1.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(session.Code) != CodeLength {
		t.Errorf("expected code length %d, got %d", CodeLength, len(session.Code))
	}
	if session.InitiatorDeviceID != "AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA" {
		t.Errorf("unexpected device ID: %s", session.InitiatorDeviceID)
	}
	if session.ResponderDeviceID != "" {
		t.Errorf("responder device ID should be empty")
	}
}

func TestStore_Get(t *testing.T) {
	store := NewStore(time.Minute)
	defer store.Close()

	session, _ := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "192.168.1.1")

	retrieved, err := store.Get(session.Code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved.Code != session.Code {
		t.Errorf("expected code %s, got %s", session.Code, retrieved.Code)
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	store := NewStore(time.Minute)
	defer store.Close()

	_, err := store.Get("NOTFND")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestStore_Get_Expired(t *testing.T) {
	store := NewStore(1 * time.Millisecond)
	defer store.Close()

	session, _ := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "192.168.1.1")
	time.Sleep(5 * time.Millisecond)

	_, err := store.Get(session.Code)
	if err != ErrSessionExpired {
		t.Errorf("expected ErrSessionExpired, got %v", err)
	}
}

func TestStore_Get_CaseInsensitive(t *testing.T) {
	store := NewStore(time.Minute)
	defer store.Close()

	session, _ := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "192.168.1.1")

	retrieved, err := store.Get("  " + session.Code + "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved.Code != session.Code {
		t.Errorf("expected code %s, got %s", session.Code, retrieved.Code)
	}
}

func TestStore_SetResponse(t *testing.T) {
	store := NewStore(time.Minute)
	defer store.Close()

	session, _ := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "192.168.1.1")
	responderID := "BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB"

	err := store.SetResponse(session.Code, responderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, _ := store.Get(session.Code)
	if retrieved.ResponderDeviceID != responderID {
		t.Errorf("expected responder device ID %s, got %s", responderID, retrieved.ResponderDeviceID)
	}
}

func TestStore_SetResponse_AlreadySet(t *testing.T) {
	store := NewStore(time.Minute)
	defer store.Close()

	session, _ := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "192.168.1.1")
	_ = store.SetResponse(session.Code, "BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB")

	err := store.SetResponse(session.Code, "CCCCCCC-CCCCCCC-CCCCCCC-CCCCCCC-CCCCCCC-CCCCCCC-CCCCCCC-CCCCCCC")
	if err != ErrResponseAlreadySet {
		t.Errorf("expected ErrResponseAlreadySet, got %v", err)
	}
}

func TestStore_Delete(t *testing.T) {
	store := NewStore(time.Minute)
	defer store.Close()

	session, _ := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "192.168.1.1")

	err := store.Delete(session.Code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = store.Get(session.Code)
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestStore_MaxSessionsPerIP(t *testing.T) {
	store := NewStore(time.Minute)
	defer store.Close()

	for i := 0; i < MaxSessionsPerIP; i++ {
		_, err := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "192.168.1.1")
		if err != nil {
			t.Fatalf("unexpected error creating session %d: %v", i, err)
		}
	}

	_, err := store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "192.168.1.1")
	if err != ErrTooManySessionsForIP {
		t.Errorf("expected ErrTooManySessionsForIP, got %v", err)
	}

	_, err = store.Create("AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA", "192.168.1.2")
	if err != nil {
		t.Errorf("should allow sessions from different IP: %v", err)
	}
}

func TestSession_HasResponse(t *testing.T) {
	session := &Session{
		Code:              "ABC123",
		InitiatorDeviceID: "AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA-AAAAAAA",
	}

	if session.HasResponse() {
		t.Error("expected HasResponse to be false")
	}

	session.ResponderDeviceID = "BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB-BBBBBBB"

	if !session.HasResponse() {
		t.Error("expected HasResponse to be true")
	}
}
