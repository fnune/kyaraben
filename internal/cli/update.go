package cli

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fnune/kyaraben/internal/version"
)

type UpdateCmd struct {
	Check bool `help:"Check for updates without installing."`
	Yes   bool `short:"y" help:"Skip confirmation prompt."`
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (cmd *UpdateCmd) Run(ctx *Context) error {
	releasesURL := os.Getenv("KYARABEN_RELEASES_URL")
	if releasesURL == "" {
		releasesURL = "https://api.github.com/repos/fnune/kyaraben/releases/latest"
	}

	fmt.Println("Checking for CLI updates...")

	release, err := fetchRelease(releasesURL)
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	currentVersion := version.Version
	latestVersion := strings.TrimPrefix(release.TagName, "v")

	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Printf("Latest version:  %s\n", latestVersion)

	if currentVersion == latestVersion {
		fmt.Println("\nYou're already on the latest version.")
		return nil
	}

	assetName := expectedAssetName()
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no release asset found for %s", assetName)
	}

	if cmd.Check {
		fmt.Printf("\nUpdate available: %s -> %s\n", currentVersion, latestVersion)
		fmt.Println("Run 'kyaraben update' to install.")
		return nil
	}

	if !cmd.Yes {
		fmt.Printf("\nDownload and install update? [y/N] ")
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return nil
		}
		if strings.ToLower(response) != "y" {
			fmt.Println("Update cancelled.")
			return nil
		}
	}

	fmt.Printf("\nDownloading %s...\n", assetName)

	tempFile, err := downloadAsset(downloadURL)
	if err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}
	defer func() { _ = os.Remove(tempFile) }()

	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}
	binaryPath, err = filepath.EvalSymlinks(binaryPath)
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	if err := installUpdate(tempFile, binaryPath); err != nil {
		return fmt.Errorf("installing update: %w", err)
	}

	fmt.Println("\nCLI update installed successfully.")
	fmt.Println("Run 'kyaraben status' to verify.")
	fmt.Println()
	fmt.Println("Note: This updates the CLI only. The desktop app updates separately.")
	return nil
}

func fetchRelease(url string) (*githubRelease, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "kyaraben-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func expectedAssetName() string {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "amd64"
	}
	return fmt.Sprintf("kyaraben-cli-%s-%s.tar.gz", runtime.GOOS, arch)
}

func downloadAsset(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	tempFile, err := os.CreateTemp("", "kyaraben-update-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer func() { _ = tempFile.Close() }()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		_ = os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

func installUpdate(tarPath, binaryPath string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if hdr.Name == "kyaraben" || strings.HasSuffix(hdr.Name, "/kyaraben") {
			tempBinary, err := os.CreateTemp(filepath.Dir(binaryPath), "kyaraben-new-*")
			if err != nil {
				return fmt.Errorf("creating temp binary: %w", err)
			}
			tempPath := tempBinary.Name()

			if _, err := io.Copy(tempBinary, tr); err != nil {
				_ = tempBinary.Close()
				_ = os.Remove(tempPath)
				return fmt.Errorf("extracting binary: %w", err)
			}
			_ = tempBinary.Close()

			if err := os.Chmod(tempPath, 0755); err != nil {
				_ = os.Remove(tempPath)
				return fmt.Errorf("setting permissions: %w", err)
			}

			if err := os.Rename(tempPath, binaryPath); err != nil {
				_ = os.Remove(tempPath)
				return fmt.Errorf("replacing binary: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("kyaraben binary not found in archive")
}
