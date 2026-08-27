package syncmeta

import (
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/testutil"
)

func TestPublishReadRoundTrip(t *testing.T) {
	fs := testutil.NewTestFS(t, nil)
	store := NewStore(fs, "/c")

	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	if err := store.Publish("DEVICE-A", "0.1.4", true, now); err != nil {
		t.Fatalf("Publish A: %v", err)
	}
	if err := store.Publish("DEVICE-B", "0.1.4", false, now); err != nil {
		t.Fatalf("Publish B: %v", err)
	}

	all, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("got %d devices, want 2", len(all))
	}
	if !all["DEVICE-A"].IgnoreDeleteROMs {
		t.Errorf("DEVICE-A IgnoreDeleteROMs = false, want true")
	}
	if all["DEVICE-B"].IgnoreDeleteROMs {
		t.Errorf("DEVICE-B IgnoreDeleteROMs = true, want false")
	}
}

func TestPublishIdempotentOnUnchangedContent(t *testing.T) {
	fs := testutil.NewTestFS(t, nil)
	store := NewStore(fs, "/c")

	t1 := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)

	if err := store.Publish("DEVICE-A", "0.1.4", true, t1); err != nil {
		t.Fatalf("Publish 1: %v", err)
	}
	if err := store.Publish("DEVICE-A", "0.1.4", true, t2); err != nil {
		t.Fatalf("Publish 2: %v", err)
	}

	all, _ := store.ReadAll()
	if got := all["DEVICE-A"].PublishedAt; got != t1.Format(time.RFC3339) {
		t.Errorf("publishedAt = %q, want %q (unchanged content should not rewrite)", got, t1.Format(time.RFC3339))
	}

	if err := store.Publish("DEVICE-A", "0.1.4", false, t2); err != nil {
		t.Fatalf("Publish 3: %v", err)
	}
	all, _ = store.ReadAll()
	if got := all["DEVICE-A"].PublishedAt; got != t2.Format(time.RFC3339) {
		t.Errorf("publishedAt = %q, want %q (changed content should rewrite)", got, t2.Format(time.RFC3339))
	}
}

func TestRemove(t *testing.T) {
	fs := testutil.NewTestFS(t, nil)
	store := NewStore(fs, "/c")
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	_ = store.Publish("DEVICE-A", "0.1.4", true, now)
	if err := store.Remove("DEVICE-A"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if err := store.Remove("DEVICE-A"); err != nil {
		t.Fatalf("Remove of missing marker should be nil, got: %v", err)
	}
	all, _ := store.ReadAll()
	if len(all) != 0 {
		t.Errorf("got %d devices after remove, want 0", len(all))
	}
}

func TestReadAllStrictAndForwardCompatible(t *testing.T) {
	good := `{"schemaVersion":1,"deviceId":"GOOD","publishedAt":"2026-06-25T12:00:00Z","kyarabenVersion":"0.1.4","rom":{"ignoreDelete":true}}`
	additive := `{"schemaVersion":1,"deviceId":"ADDITIVE","futureField":42,"rom":{"ignoreDelete":false,"futureNested":"x"}}`

	fs := testutil.NewTestFS(t, map[string]any{
		"/c/.kyaraben/GOOD.json":       good,
		"/c/.kyaraben/ADDITIVE.json":   additive,
		"/c/.kyaraben/NOVERSION.json":  `{"deviceId":"NOVERSION","rom":{"ignoreDelete":true}}`,
		"/c/.kyaraben/FUTURE.json":     `{"schemaVersion":999,"deviceId":"FUTURE","rom":{"ignoreDelete":true}}`,
		"/c/.kyaraben/WRONGTYPE.json":  `{"schemaVersion":1,"deviceId":"WRONGTYPE","rom":{"ignoreDelete":"yes"}}`,
		"/c/.kyaraben/MISMATCH.json":   `{"schemaVersion":1,"deviceId":"OTHER","rom":{"ignoreDelete":true}}`,
		"/c/.kyaraben/NOROMFIELD.json": `{"schemaVersion":1,"deviceId":"NOROMFIELD","rom":{}}`,
		"/c/.kyaraben/MALFORMED.json":  `{not json`,
		"/c/.kyaraben/notes.txt":       `ignored, not a marker`,
	})
	store := NewStore(fs, "/c")

	all, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if _, ok := all["GOOD"]; !ok {
		t.Errorf("GOOD marker should parse")
	}
	if s, ok := all["ADDITIVE"]; !ok || s.IgnoreDeleteROMs {
		t.Errorf("ADDITIVE marker should parse and ignore unknown fields, got %+v ok=%v", s, ok)
	}
	for _, rejected := range []string{"NOVERSION", "FUTURE", "WRONGTYPE", "OTHER", "MISMATCH", "NOROMFIELD"} {
		if _, ok := all[rejected]; ok {
			t.Errorf("%s marker should have been rejected", rejected)
		}
	}
	if len(all) != 2 {
		t.Errorf("got %d valid markers, want 2 (GOOD, ADDITIVE): %v", len(all), all)
	}
}
