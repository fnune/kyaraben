package nix

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

func TestBuild_CallsNixPortable(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &FakeRunner{
		OutputData: []byte("/nix/store/abc123-package\n"),
	}

	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := c.Build(context.Background(), "/path/to/flake#package")
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	if len(runner.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(runner.Calls))
	}

	call := runner.Calls[0]
	if call.Name != c.NixPortableBinary {
		t.Errorf("called %s, want %s", call.Name, c.NixPortableBinary)
	}
	if call.Args[0] != "nix" || call.Args[1] != "build" {
		t.Errorf("args = %v, want nix build ...", call.Args)
	}
	if !slices.Contains(call.Args, "/path/to/flake#package") {
		t.Errorf("args missing flake ref: %v", call.Args)
	}
}

func TestBuild_ReturnsStorePath(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &FakeRunner{
		OutputData: []byte("/nix/store/abc123-package\n"),
	}

	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	path, err := c.Build(context.Background(), "flake#pkg")
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	want := "/nix/store/abc123-package"
	if path != want {
		t.Errorf("Build() = %q, want %q", path, want)
	}
}

func TestBuild_PropagatesError(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &FakeRunner{
		OutputErr: errors.New("nix build failed"),
	}

	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := c.Build(context.Background(), "flake#pkg")
	if err == nil {
		t.Fatal("Build() expected error, got nil")
	}
}

func TestBuild_FailsWhenNixPortableNotAvailable(t *testing.T) {
	tmpDir := t.TempDir()
	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nonexistent"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              &FakeRunner{},
	}

	_, err := c.Build(context.Background(), "flake#pkg")
	if err == nil {
		t.Fatal("Build() expected error when nix-portable not available")
	}
}

func TestBuildWithLink_SetsOutLink(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &FakeRunner{}

	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	outLink := filepath.Join(tmpDir, "result")
	err := c.BuildWithLink(context.Background(), "flake#pkg", outLink)
	if err != nil {
		t.Fatalf("BuildWithLink() error: %v", err)
	}

	if len(runner.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(runner.Calls))
	}

	call := runner.Calls[0]
	outLinkIdx := slices.Index(call.Args, "--out-link")
	if outLinkIdx == -1 {
		t.Fatalf("args missing --out-link: %v", call.Args)
	}
	if call.Args[outLinkIdx+1] != outLink {
		t.Errorf("--out-link value = %s, want %s", call.Args[outLinkIdx+1], outLink)
	}
}

func TestGetVersion_ParsesOutput(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &FakeRunner{
		OutputData: []byte("nix (Nix) 2.18.1\n"),
	}

	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	version, err := c.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("GetVersion() error: %v", err)
	}

	want := "nix (Nix) 2.18.1"
	if version != want {
		t.Errorf("GetVersion() = %q, want %q", version, want)
	}
}

func TestFlakeUpdate_SetsWorkingDir(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &FakeRunner{}

	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	flakePath := filepath.Join(tmpDir, "my-flake")
	if err := os.MkdirAll(flakePath, 0755); err != nil {
		t.Fatal(err)
	}

	err := c.FlakeUpdate(context.Background(), flakePath)
	if err != nil {
		t.Fatalf("FlakeUpdate() error: %v", err)
	}

	if len(runner.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(runner.Calls))
	}

	call := runner.Calls[0]
	if call.Opts.Dir != flakePath {
		t.Errorf("Dir = %q, want %q", call.Opts.Dir, flakePath)
	}
}

func TestFlakeCheck_CallsFlakeShow(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &FakeRunner{}

	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	err := c.FlakeCheck(context.Background(), "/path/to/flake")
	if err != nil {
		t.Fatalf("FlakeCheck() error: %v", err)
	}

	if len(runner.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(runner.Calls))
	}

	call := runner.Calls[0]
	if !slices.Contains(call.Args, "flake") || !slices.Contains(call.Args, "show") {
		t.Errorf("args should contain 'flake show': %v", call.Args)
	}
}

func TestEval_ReturnsJSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &FakeRunner{
		OutputData: []byte(`{"foo":"bar"}`),
	}

	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	result, err := c.Eval(context.Background(), "{ foo = \"bar\"; }")
	if err != nil {
		t.Fatalf("Eval() error: %v", err)
	}

	want := `{"foo":"bar"}`
	if string(result) != want {
		t.Errorf("Eval() = %q, want %q", string(result), want)
	}
}

func TestNixPortableEnv_IsSet(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &FakeRunner{
		OutputData: []byte("/nix/store/abc123\n"),
	}

	nixLocation := filepath.Join(tmpDir, "nix")
	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: nixLocation,
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := c.Build(context.Background(), "flake#pkg")
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	call := runner.Calls[0]
	wantEnv := "NP_LOCATION=" + nixLocation
	if !slices.Contains(call.Opts.Env, wantEnv) {
		t.Errorf("Env should contain %q, got %v", wantEnv, call.Opts.Env)
	}
}

func TestGetTargetTriple(t *testing.T) {
	triple := getTargetTriple()

	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			if triple != "x86_64-unknown-linux-gnu" {
				t.Errorf("getTargetTriple() = %s, want x86_64-unknown-linux-gnu", triple)
			}
		case "arm64":
			if triple != "aarch64-unknown-linux-gnu" {
				t.Errorf("getTargetTriple() = %s, want aarch64-unknown-linux-gnu", triple)
			}
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			if triple != "x86_64-apple-darwin" {
				t.Errorf("getTargetTriple() = %s, want x86_64-apple-darwin", triple)
			}
		case "arm64":
			if triple != "aarch64-apple-darwin" {
				t.Errorf("getTargetTriple() = %s, want aarch64-apple-darwin", triple)
			}
		}
	}
}

func TestRealStorePath(t *testing.T) {
	c := &Client{
		NixPortableLocation: "/home/user/.local/state/kyaraben/build/nix",
	}

	tests := []struct {
		input string
		want  string
	}{
		{
			input: "/nix/store/abc123-package",
			want:  "/home/user/.local/state/kyaraben/build/nix/.nix-portable/nix/store/abc123-package",
		},
		{
			input: "/some/other/path",
			want:  "/some/other/path",
		},
		{
			input: "/nix/store/xyz789-another/bin/app",
			want:  "/home/user/.local/state/kyaraben/build/nix/.nix-portable/nix/store/xyz789-another/bin/app",
		},
	}

	for _, tt := range tests {
		got := c.RealStorePath(tt.input)
		if got != tt.want {
			t.Errorf("RealStorePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLineCallbackWriter(t *testing.T) {
	var lines []string
	w := &lineCallbackWriter{
		callback: func(line string) {
			lines = append(lines, line)
		},
	}

	w.Write([]byte("line1\nline2\n"))
	w.Write([]byte("partial"))
	w.Write([]byte(" continued\n"))

	want := []string{"line1", "line2", "partial continued"}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d", len(lines), len(want))
	}
	for i, line := range lines {
		if line != want[i] {
			t.Errorf("lines[%d] = %q, want %q", i, line, want[i])
		}
	}
}

func TestLineCallbackWriter_WithReplacement(t *testing.T) {
	var lines []string
	w := &lineCallbackWriter{
		callback:    func(line string) { lines = append(lines, line) },
		replaceFrom: "/nix/store/",
		replaceTo:   "~/.local/state/",
	}

	w.Write([]byte("copying /nix/store/abc123 to output\n"))

	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	want := "copying ~/.local/state/abc123 to output"
	if lines[0] != want {
		t.Errorf("line = %q, want %q", lines[0], want)
	}
}

func TestLineCallbackWriter_NilCallback(t *testing.T) {
	w := &lineCallbackWriter{}
	n, err := w.Write([]byte("some data\n"))
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	if n != 10 {
		t.Errorf("Write() = %d, want 10", n)
	}
}

func TestIsAvailable(t *testing.T) {
	tmpDir := t.TempDir()

	c := &Client{
		NixPortableBinary: filepath.Join(tmpDir, "nix-portable"),
	}

	if c.IsAvailable() {
		t.Error("IsAvailable() = true before binary exists")
	}

	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	if !c.IsAvailable() {
		t.Error("IsAvailable() = false after binary created")
	}
}

func TestEnsureNixPortableDir(t *testing.T) {
	tmpDir := t.TempDir()
	nixDir := filepath.Join(tmpDir, "nix", "portable", "dir")

	c := &Client{
		NixPortableLocation: nixDir,
	}

	if err := c.EnsureNixPortableDir(); err != nil {
		t.Fatalf("EnsureNixPortableDir() error: %v", err)
	}

	info, err := os.Stat(nixDir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("created path is not a directory")
	}
}

func TestSetOutputCallback(t *testing.T) {
	tmpDir := t.TempDir()
	var captured []string
	runner := &FakeRunner{}

	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	c.SetOutputCallback(func(line string) {
		captured = append(captured, line)
	})

	if c.outputCallback == nil {
		t.Error("outputCallback not set")
	}
}

func TestBuildMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	callCount := 0
	runner := &FakeRunner{}

	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	originalOutput := func(ctx context.Context, name string, args []string, opts RunOpts) ([]byte, error) {
		callCount++
		return []byte("/nix/store/result" + string(rune('0'+callCount)) + "\n"), nil
	}

	runner.OutputData = []byte("/nix/store/result1\n")
	results, err := c.BuildMultiple(context.Background(), []string{"flake#a", "flake#b"})
	if err != nil {
		t.Fatalf("BuildMultiple() error: %v", err)
	}
	_ = originalOutput

	if len(results) != 2 {
		t.Errorf("BuildMultiple() returned %d results, want 2", len(results))
	}
}

func TestBuild_SetsStderrWriter(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &FakeRunner{
		OutputData: []byte("/nix/store/abc123\n"),
	}

	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := c.Build(context.Background(), "flake#pkg")
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	call := runner.Calls[0]
	if call.Opts.Stderr == nil {
		t.Error("Stderr writer not set")
	}
}

func TestBuild_EmptyOutputError(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &FakeRunner{
		OutputData: []byte(""),
	}

	c := &Client{
		NixPortableBinary:   filepath.Join(tmpDir, "nix-portable"),
		NixPortableLocation: filepath.Join(tmpDir, "nix"),
		runner:              runner,
	}
	if err := os.WriteFile(c.NixPortableBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := c.Build(context.Background(), "flake#pkg")
	if err == nil {
		t.Fatal("Build() expected error for empty output")
	}
}

func TestExecRunner_Run(t *testing.T) {
	r := &ExecRunner{}
	var stdout bytes.Buffer

	err := r.Run(context.Background(), "echo", []string{"hello"}, RunOpts{
		Stdout: &stdout,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if got := stdout.String(); got != "hello\n" {
		t.Errorf("stdout = %q, want %q", got, "hello\n")
	}
}

func TestExecRunner_Output(t *testing.T) {
	r := &ExecRunner{}

	out, err := r.Output(context.Background(), "echo", []string{"hello"}, RunOpts{})
	if err != nil {
		t.Fatalf("Output() error: %v", err)
	}

	if got := string(out); got != "hello\n" {
		t.Errorf("output = %q, want %q", got, "hello\n")
	}
}
