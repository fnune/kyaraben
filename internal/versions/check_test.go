package versions

import "testing"

func TestIsNewerVersion(t *testing.T) {
	for _, tc := range []struct {
		name    string
		latest  string
		current string
		want    bool
	}{
		{name: "newer patch", latest: "0.1.7", current: "0.1.6", want: true},
		{name: "same version", latest: "0.1.6", current: "0.1.6", want: false},
		{name: "older release does not trigger a downgrade", latest: "0.1.5", current: "0.1.6", want: false},
		{name: "v prefix is tolerated", latest: "v0.1.7", current: "0.1.6", want: true},
		{name: "dev builds can still update", latest: "0.1.6", current: "dev", want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsNewerVersion(tc.latest, tc.current); got != tc.want {
				t.Errorf("IsNewerVersion(%q, %q) = %v, want %v", tc.latest, tc.current, got, tc.want)
			}
		})
	}
}
