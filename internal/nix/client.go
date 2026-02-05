package nix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/paths"
)

var log = logging.New("nix")

type Client struct {
	NixPortableBinary   string
	NixPortableLocation string // passed via NP_LOCATION env var
	FlakePath           string
	outputCallback      func(line string)
	runner              CommandRunner
}

func (c *Client) SetOutputCallback(fn func(line string)) {
	c.outputCallback = fn
}

type lineCallbackWriter struct {
	callback    func(string)
	replaceFrom string
	replaceTo   string
	buf         bytes.Buffer
	mu          sync.Mutex
}

func (w *lineCallbackWriter) Write(p []byte) (n int, err error) {
	if w.callback == nil {
		return len(p), nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf.Write(p)

	for {
		line, err := w.buf.ReadString('\n')
		if err != nil {
			w.buf.WriteString(line)
			break
		}
		line = strings.TrimRight(line, "\n\r")
		if w.replaceFrom != "" && w.replaceTo != "" {
			line = strings.ReplaceAll(line, w.replaceFrom, w.replaceTo)
		}
		w.callback(line)
	}

	return len(p), nil
}

func NewClient() (*Client, error) {
	stateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return nil, err
	}

	// Allows dry-run and other non-Nix operations to work without nix-portable.
	nixPortable, findErr := findNixPortable()
	if findErr != nil {
		log.Debug("nix-portable not found: %v", findErr)
	}

	buildDir := filepath.Join(stateDir, "build")
	return &Client{
		NixPortableBinary:   nixPortable,
		NixPortableLocation: filepath.Join(buildDir, "nix"),
		FlakePath:           filepath.Join(buildDir, "flake"),
		runner:              &ExecRunner{},
	}, nil
}

func findNixPortable() (string, error) {
	if override := os.Getenv("KYARABEN_NIX_PORTABLE_PATH"); override != "" {
		if _, err := os.Stat(override); err == nil {
			log.Debug("Using nix-portable from KYARABEN_NIX_PORTABLE_PATH: %s", override)
			return override, nil
		}
		log.Debug("KYARABEN_NIX_PORTABLE_PATH set but file not found: %s", override)
	}

	targetTriple := getTargetTriple()
	binaryName := "nix-portable-" + targetTriple

	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("getting executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	log.Debug("Looking for nix-portable: %s", binaryName)
	log.Debug("Executable dir: %s", execDir)

	// Search locations in order of preference:
	// 1. Same directory as executable (AppImage/installed)
	// 2. ../binaries/ relative to executable (development)
	// 3. ui/binaries/ from project root (development)
	searchPaths := []string{
		filepath.Join(execDir, binaryName),
		filepath.Join(execDir, "..", "binaries", binaryName),
	}

	// For development, also check relative to working directory
	if cwd, err := os.Getwd(); err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(cwd, "ui", "binaries", binaryName),
		)
	}

	for _, path := range searchPaths {
		log.Debug("Checking: %s", path)
		if _, err := os.Stat(path); err == nil {
			log.Debug("Found nix-portable at: %s", path)
			return path, nil
		}
	}

	log.Debug("nix-portable NOT FOUND")
	return "", fmt.Errorf("nix-portable binary not found (searched for %s)", binaryName)
}

func getTargetTriple() string {
	arch := runtime.GOARCH
	os := runtime.GOOS

	switch os {
	case "linux":
		switch arch {
		case "amd64":
			return "x86_64-unknown-linux-gnu"
		case "arm64":
			return "aarch64-unknown-linux-gnu"
		default:
			return "unknown-unknown-linux-gnu"
		}
	case "darwin":
		switch arch {
		case "amd64":
			return "x86_64-apple-darwin"
		case "arm64":
			return "aarch64-apple-darwin"
		default:
			return "unknown-apple-darwin"
		}
	default:
		return "unknown-unknown-unknown"
	}
}

func (c *Client) IsAvailable() bool {
	_, err := os.Stat(c.NixPortableBinary)
	return err == nil
}

func (c *Client) prepareNixRun(args []string) (fullArgs []string, baseOpts RunOpts, err error) {
	if !c.IsAvailable() {
		log.Error("nix-portable not available at: %s", c.NixPortableBinary)
		return nil, RunOpts{}, fmt.Errorf("nix-portable is not available (bundled binary not found)")
	}

	if err := c.EnsureNixPortableDir(); err != nil {
		return nil, RunOpts{}, fmt.Errorf("creating nix-portable data directory: %w", err)
	}

	fullArgs = append([]string{"nix"}, args...)
	baseOpts = RunOpts{
		Env: append(os.Environ(), "NP_LOCATION="+c.NixPortableLocation),
	}

	log.Debug("Running: %s %v", c.NixPortableBinary, fullArgs)
	log.Debug("NP_LOCATION=%s", c.NixPortableLocation)

	return fullArgs, baseOpts, nil
}

func (c *Client) EnsureNixPortableDir() error {
	return os.MkdirAll(c.NixPortableLocation, 0755)
}

func (c *Client) Build(ctx context.Context, flakeRef string) (string, error) {
	log.Info("Starting nix build for: %s", flakeRef)

	args := []string{
		"build",
		flakeRef,
		"--no-link",
		"--print-out-paths",
		"-L",
	}

	if os.Getenv("KYARABEN_NIX_NO_SANDBOX") == "1" {
		args = append(args, "--option", "sandbox", "false")
	}

	fullArgs, opts, err := c.prepareNixRun(args)
	if err != nil {
		return "", err
	}

	var stderr bytes.Buffer
	opts.Stderr = io.MultiWriter(&stderr, os.Stderr, logging.Writer())

	log.Info("Executing nix build (this may take a while on first run)...")
	stdout, err := c.runner.Output(ctx, c.NixPortableBinary, fullArgs, opts)
	if err != nil {
		log.Error("nix build FAILED: %v", err)
		return "", fmt.Errorf("nix build failed: %w\nstderr: %s", err, stderr.String())
	}

	storePath := strings.TrimSpace(string(stdout))
	if storePath == "" {
		log.Error("nix build produced no output")
		return "", fmt.Errorf("nix build produced no output")
	}

	log.Info("nix build SUCCESS: %s", storePath)
	return storePath, nil
}

// BuildWithLink builds a flake and creates a symlink to the result.
// The symlink is created by nix itself, which ensures it works correctly
// with nix-portable's virtualized store.
func (c *Client) BuildWithLink(ctx context.Context, flakeRef string, outLink string) error {
	log.Info("Starting nix build for: %s (link: %s)", flakeRef, outLink)

	if err := os.MkdirAll(filepath.Dir(outLink), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	args := []string{
		"build",
		flakeRef,
		"--out-link", outLink,
		"-L",
	}

	if os.Getenv("KYARABEN_NIX_NO_SANDBOX") == "1" {
		args = append(args, "--option", "sandbox", "false")
	}

	fullArgs, opts, err := c.prepareNixRun(args)
	if err != nil {
		return err
	}

	var stderr bytes.Buffer
	writers := []io.Writer{&stderr, os.Stderr, logging.Writer()}
	if c.outputCallback != nil {
		writers = append(writers, &lineCallbackWriter{
			callback:    c.outputCallback,
			replaceFrom: "/nix/store/",
			replaceTo:   "~/.local/state/kyaraben/",
		})
	}
	opts.Stderr = io.MultiWriter(writers...)

	log.Info("Executing nix build (this may take a while on first run)...")
	if err := c.runner.Run(ctx, c.NixPortableBinary, fullArgs, opts); err != nil {
		log.Error("nix build FAILED: %v", err)
		return fmt.Errorf("nix build failed: %w\nstderr: %s", err, stderr.String())
	}

	log.Info("nix build SUCCESS: created link at %s", outLink)
	return nil
}

func (c *Client) BuildMultiple(ctx context.Context, flakeRefs []string) (map[string]string, error) {
	results := make(map[string]string)

	for _, ref := range flakeRefs {
		path, err := c.Build(ctx, ref)
		if err != nil {
			return nil, fmt.Errorf("building %s: %w", ref, err)
		}
		results[ref] = path
	}

	return results, nil
}

func (c *Client) Eval(ctx context.Context, expr string) (json.RawMessage, error) {
	args := []string{
		"eval",
		"--json",
		"--expr", expr,
	}

	fullArgs, opts, err := c.prepareNixRun(args)
	if err != nil {
		return nil, err
	}

	var stderr bytes.Buffer
	opts.Stderr = &stderr

	stdout, err := c.runner.Output(ctx, c.NixPortableBinary, fullArgs, opts)
	if err != nil {
		return nil, fmt.Errorf("nix eval failed: %w\nstderr: %s", err, stderr.String())
	}

	return json.RawMessage(stdout), nil
}

func (c *Client) FlakeUpdate(ctx context.Context, flakePath string) error {
	args := []string{
		"flake",
		"update",
	}

	fullArgs, opts, err := c.prepareNixRun(args)
	if err != nil {
		return err
	}

	opts.Dir = flakePath
	var stderr bytes.Buffer
	opts.Stderr = &stderr

	if err := c.runner.Run(ctx, c.NixPortableBinary, fullArgs, opts); err != nil {
		return fmt.Errorf("nix flake update failed: %w\nstderr: %s", err, stderr.String())
	}

	return nil
}

func (c *Client) GetVersion(ctx context.Context) (string, error) {
	fullArgs, opts, err := c.prepareNixRun([]string{"--version"})
	if err != nil {
		return "", err
	}

	stdout, err := c.runner.Output(ctx, c.NixPortableBinary, fullArgs, opts)
	if err != nil {
		return "", fmt.Errorf("getting nix version: %w", err)
	}

	return strings.TrimSpace(string(stdout)), nil
}

func (c *Client) EnsureFlakeDir() error {
	return os.MkdirAll(c.FlakePath, 0755)
}

func (c *Client) RealStorePath(virtualPath string) string {
	const nixStorePrefix = "/nix/store/"
	if !strings.HasPrefix(virtualPath, nixStorePrefix) {
		return virtualPath
	}
	hashAndName := strings.TrimPrefix(virtualPath, nixStorePrefix)
	return filepath.Join(c.NixPortableLocation, ".nix-portable", "nix", "store", hashAndName)
}

func (c *Client) GetNixPortableBinary() string {
	return c.NixPortableBinary
}

func (c *Client) GetPersistentNixPortablePath() string {
	return filepath.Join(c.NixPortableLocation, "nix-portable")
}

func (c *Client) EnsurePersistentNixPortable() (string, error) {
	persistentPath := c.GetPersistentNixPortablePath()

	if _, err := os.Stat(persistentPath); err == nil {
		currentBinaryInfo, _ := os.Stat(c.NixPortableBinary)
		persistentInfo, _ := os.Stat(persistentPath)
		if currentBinaryInfo != nil && persistentInfo != nil &&
			currentBinaryInfo.Size() == persistentInfo.Size() {
			return persistentPath, nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(persistentPath), 0755); err != nil {
		return "", fmt.Errorf("creating directory for persistent nix-portable: %w", err)
	}

	if err := copyFile(c.NixPortableBinary, persistentPath); err != nil {
		return "", fmt.Errorf("copying nix-portable: %w", err)
	}

	log.Info("Copied nix-portable to persistent location: %s", persistentPath)
	return persistentPath, nil
}

func copyFile(srcPath, dstPath string) (err error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := src.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := dst.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(dst, src)
	return err
}

func (c *Client) GetNixPortableLocation() string {
	return c.NixPortableLocation
}

func (c *Client) FlakeCheck(ctx context.Context, flakePath string) error {
	args := []string{
		"flake",
		"show",
		"--json",
		flakePath,
	}

	fullArgs, opts, err := c.prepareNixRun(args)
	if err != nil {
		return err
	}

	var stderr bytes.Buffer
	opts.Stderr = &stderr

	if err := c.runner.Run(ctx, c.NixPortableBinary, fullArgs, opts); err != nil {
		return fmt.Errorf("nix flake show failed: %w\nstderr: %s", err, stderr.String())
	}

	return nil
}
