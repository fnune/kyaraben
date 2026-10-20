package model

import (
	"fmt"
	"net/url"
)

type CollectionConfig struct {
	URL      string                          `toml:"url"`
	Vertical string                          `toml:"vertical,omitempty"`
	VRR      bool                            `toml:"vrr,omitempty"`
	Games    map[string]CollectionGameConfig `toml:"games,omitempty"`
}

type CollectionGameConfig struct {
	Emulator EmulatorID `toml:"emulator,omitempty"`
	Vertical string     `toml:"vertical,omitempty"`
}

const (
	VerticalRotate = "rotate"
	VerticalNative = "native"
)

func (c CollectionConfig) VerticalFor(gameID string) string {
	if g, ok := c.Games[gameID]; ok && g.Vertical != "" {
		return g.Vertical
	}
	if c.Vertical != "" {
		return c.Vertical
	}
	return VerticalRotate
}

func (c CollectionConfig) normalize() (CollectionConfig, ConfigWarnings) {
	var warnings ConfigWarnings
	if u, err := url.Parse(c.URL); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		warnings = append(warnings, ConfigWarning{Field: "url", Message: fmt.Sprintf("invalid manifest URL %q", c.URL)})
	}
	if !validVertical(c.Vertical) {
		warnings = append(warnings, ConfigWarning{Field: "vertical", Message: fmt.Sprintf("unknown value %q, using %q", c.Vertical, VerticalRotate)})
		c.Vertical = ""
	}
	games := make(map[string]CollectionGameConfig, len(c.Games))
	for gameID, g := range c.Games {
		if !validVertical(g.Vertical) {
			warnings = append(warnings, ConfigWarning{Field: "games." + gameID + ".vertical", Message: fmt.Sprintf("unknown value %q, ignoring", g.Vertical)})
			g.Vertical = ""
		}
		games[gameID] = g
	}
	c.Games = games
	return c, warnings
}

func validVertical(v string) bool {
	return v == "" || v == VerticalRotate || v == VerticalNative
}
