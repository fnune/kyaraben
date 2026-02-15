package sync

import (
	"strings"
	"testing"
)

func TestGeneratePairingCode(t *testing.T) {
	code, err := GeneratePairingCode()
	if err != nil {
		t.Fatalf("GeneratePairingCode: %v", err)
	}

	if len(code) != PairingCodeLength {
		t.Errorf("expected length %d, got %d", PairingCodeLength, len(code))
	}

	for _, c := range code {
		if !strings.ContainsRune(PairingCodeCharset, c) {
			t.Errorf("code contains invalid character %q", c)
		}
	}
}

func TestGeneratePairingCodeUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := GeneratePairingCode()
		if err != nil {
			t.Fatalf("GeneratePairingCode: %v", err)
		}
		if seen[code] {
			t.Errorf("duplicate code generated: %s", code)
		}
		seen[code] = true
	}
}
