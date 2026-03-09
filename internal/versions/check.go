package versions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
)

type ReleaseInfo struct {
	Tag        string
	Prerelease bool
}

type VersionCheck struct {
	Package        string
	Current        string
	LatestStable   string
	LatestPre      string
	HasUpdate      bool
	Error          error
	ReleasesURL    string
	ReleasesSource string
}

func CheckAllVersions(ctx context.Context) <-chan VersionCheck {
	results := make(chan VersionCheck)

	go func() {
		defer close(results)

		v := MustGet()
		var wg sync.WaitGroup

		for name, spec := range v.Packages {
			if spec.ReleasesURL != "" {
				wg.Add(1)
				go func(name string, spec PackageSpec) {
					defer wg.Done()
					check := checkReleasesURL(ctx, name, spec.Default, spec.ReleasesURL)
					results <- check
				}(name, spec)
			}
		}

		wg.Wait()
	}()

	return results
}

func checkReleasesURL(ctx context.Context, pkg, current, releasesURL string) VersionCheck {
	check := VersionCheck{
		Package:     pkg,
		Current:     current,
		ReleasesURL: releasesURL,
	}

	parts := strings.SplitN(releasesURL, ":", 2)
	if len(parts) != 2 {
		check.Error = fmt.Errorf("invalid releases_url format: %s", releasesURL)
		return check
	}

	source := parts[0]
	path := parts[1]
	check.ReleasesSource = source

	var releases []ReleaseInfo
	var err error

	switch source {
	case "github":
		releases, err = fetchGitHubReleases(ctx, path)
	case "gitlab":
		releases, err = fetchGitLabReleases(ctx, path)
	case "forgejo":
		releases, err = fetchForgejoReleases(ctx, path)
	default:
		check.Error = fmt.Errorf("unknown release source: %s", source)
		return check
	}

	if err != nil {
		check.Error = err
		return check
	}

	for _, rel := range releases {
		if isNonVersionTag(rel.Tag) {
			continue
		}

		isPrerelease := rel.Prerelease || looksLikePrerelease(rel.Tag)

		if isPrerelease {
			if check.LatestPre == "" {
				check.LatestPre = rel.Tag
			}
		} else {
			if check.LatestStable == "" {
				check.LatestStable = rel.Tag
			}
		}
		if check.LatestStable != "" && check.LatestPre != "" {
			break
		}
	}

	if check.LatestStable == "" {
		check.LatestStable = check.LatestPre
	}

	check.HasUpdate = check.LatestStable != "" && isNewerVersion(check.LatestStable, current)

	return check
}

func isNewerVersion(latest, current string) bool {
	latestV, errLatest := semver.NewVersion(latest)
	currentV, errCurrent := semver.NewVersion(current)

	if errLatest == nil && errCurrent == nil {
		return latestV.GreaterThan(currentV)
	}

	latestNum := extractBuildNumber(latest)
	currentNum := extractBuildNumber(current)
	if latestNum > 0 && currentNum > 0 {
		return latestNum > currentNum
	}

	return normalizeVersion(latest) != normalizeVersion(current)
}

func normalizeVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}

func isNonVersionTag(tag string) bool {
	lower := strings.ToLower(tag)
	nonVersionTags := []string{"latest", "preview", "nightly", "canary", "dev", "master", "main"}
	for _, nv := range nonVersionTags {
		if lower == nv {
			return true
		}
	}
	return strings.HasPrefix(lower, "build-")
}

func looksLikePrerelease(tag string) bool {
	v, err := semver.NewVersion(tag)
	if err != nil {
		return false
	}
	return v.Prerelease() != ""
}

func extractBuildNumber(v string) int {
	v = strings.TrimPrefix(v, "v")
	if idx := strings.Index(v, "@"); idx > 0 {
		v = v[:idx]
	}

	var num int
	_, _ = fmt.Sscanf(v, "%d", &num)
	return num
}

type githubRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}

func fetchGitHubReleases(ctx context.Context, repo string) ([]ReleaseInfo, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=20", repo)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("rate limited (set GITHUB_TOKEN for higher limits)")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var ghReleases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&ghReleases); err != nil {
		return nil, err
	}

	var releases []ReleaseInfo
	for _, r := range ghReleases {
		if r.Draft {
			continue
		}
		releases = append(releases, ReleaseInfo{
			Tag:        r.TagName,
			Prerelease: r.Prerelease,
		})
	}

	return releases, nil
}

type gitlabRelease struct {
	TagName string `json:"tag_name"`
}

func fetchGitLabReleases(ctx context.Context, project string) ([]ReleaseInfo, error) {
	encodedPath := url.PathEscape(project)
	apiURL := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/releases?per_page=20", encodedPath)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var glReleases []gitlabRelease
	if err := json.NewDecoder(resp.Body).Decode(&glReleases); err != nil {
		return nil, err
	}

	var releases []ReleaseInfo
	for _, r := range glReleases {
		releases = append(releases, ReleaseInfo{
			Tag:        r.TagName,
			Prerelease: false,
		})
	}

	return releases, nil
}

func fetchForgejoReleases(ctx context.Context, path string) ([]ReleaseInfo, error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid forgejo path: expected host/owner/repo, got %s", path)
	}
	host := parts[0]
	repo := parts[1]

	apiURL := fmt.Sprintf("https://%s/api/v1/repos/%s/releases?limit=20", host, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var fgReleases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&fgReleases); err != nil {
		return nil, err
	}

	var releases []ReleaseInfo
	for _, r := range fgReleases {
		if r.Draft {
			continue
		}
		releases = append(releases, ReleaseInfo{
			Tag:        r.TagName,
			Prerelease: r.Prerelease,
		})
	}

	return releases, nil
}
