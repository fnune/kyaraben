package sdlmappings

import (
	"strings"
	"testing"
)

const devices = `I: Bus=0003 Vendor=054c Product=09cc Version=8111
N: Name="Sony Interactive Entertainment Wireless Controller"
H: Handlers=event20 js0

I: Bus=0003 Vendor=054c Product=09cc Version=8111
N: Name="Sony Interactive Entertainment Wireless Controller Touchpad"
H: Handlers=event22 mouse2

I: Bus=0003 Vendor=28de Product=11ff Version=0001
N: Name="Steam Virtual Gamepad"
H: Handlers=event30 js1

I: Bus=0003 Vendor=046d Product=c52b Version=0111
N: Name="Logitech USB Receiver"
H: Handlers=sysrq kbd event3

`

func TestForDevicesPutsConnectedPadsFirst(t *testing.T) {
	got, err := ForDevices(strings.NewReader(devices), Budget)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(got, "\n")
	if !strings.HasPrefix(lines[0], "030000004c050000cc09000011810000,PS4 Controller,") {
		t.Fatalf("expected the connected DualShock 4 first, got %q", lines[0])
	}
	if !strings.Contains(lines[0], "back:b8") || !strings.Contains(lines[0], "leftshoulder:b4") {
		t.Errorf("unexpected DualShock 4 mapping %q", lines[0])
	}
}

func TestForDevicesStaysWithinBudget(t *testing.T) {
	got, err := ForDevices(strings.NewReader(devices), Budget)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) > Budget {
		t.Fatalf("%d bytes exceed the %d byte budget", len(got), Budget)
	}
	small, _ := ForDevices(strings.NewReader(devices), 400)
	if !strings.HasPrefix(small, "030000004c050000cc09000011810000,") || strings.Count(small, "\n") > 1 {
		t.Errorf("a tight budget should keep the connected pad: %q", small)
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(got, "\n") {
		if seen[line[:32]] {
			t.Fatalf("duplicate GUID %s", line[:32])
		}
		seen[line[:32]] = true
	}
}

func TestGUIDMatchesSDL2Linux(t *testing.T) {
	got := guidFromIDLine("I: Bus=0003 Vendor=054c Product=09cc Version=8111")
	if got != "030000004c050000cc09000011810000" {
		t.Errorf("got %s", got)
	}
}

func TestVersionlessEntriesApplyToEveryVersion(t *testing.T) {
	var versionless string
	for _, line := range strings.Split(database, "\n") {
		if len(line) > 32 && !strings.HasPrefix(line, "#") && line[24:28] == "0000" {
			versionless = line
			break
		}
	}
	if versionless == "" {
		t.Skip("no versionless entry in the database")
	}
	device := versionless[:24] + "4242" + versionless[28:32]
	got := lookup([]string{device})
	if len(got) != 1 || got[0] != device+versionless[32:] {
		t.Errorf("expected %q rewritten onto %s, got %q", versionless, device, got)
	}
}

func TestDatabaseLeavesValveDevicesAlone(t *testing.T) {
	for _, line := range strings.Split(database, "\n") {
		if len(line) > 12 && line[8:12] == "de28" {
			t.Fatalf("Valve device in database: %s", line)
		}
	}
}
