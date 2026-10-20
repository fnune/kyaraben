package model

import "testing"

func TestCollectionConfigNormalize(t *testing.T) {
	c := CollectionConfig{
		URL:      "ftp://example.com/x.json",
		Vertical: "sideways",
		Games:    map[string]CollectionGameConfig{"ddp": {Vertical: "upside"}, "esp": {Vertical: VerticalNative}},
	}
	got, warnings := c.normalize()
	if len(warnings) != 3 {
		t.Errorf("expected 3 warnings, got %v", warnings)
	}
	if got.VerticalFor("ddp") != VerticalRotate || got.VerticalFor("esp") != VerticalNative {
		t.Errorf("unexpected vertical modes: %+v", got)
	}
}
