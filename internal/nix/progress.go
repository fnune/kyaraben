package nix

import (
	"regexp"
	"strings"
)

type BuildPhase string

const (
	PhaseEvaluating BuildPhase = "evaluating"
	PhaseInstalling BuildPhase = "installing"
)

type BuildProgress struct {
	Phase           BuildPhase
	PackageName     string
	ProgressPercent int
}

type ExpectedPackage struct {
	Name        string
	DisplayName string
	SizeBytes   int64
}

var (
	evaluatingPattern  = regexp.MustCompile(`^evaluating`)
	buildingPattern    = regexp.MustCompile(`^building '[^']+/[a-z0-9]+-([^/]+)\.drv'`)
	copyingFromPattern = regexp.MustCompile(`^copying path '[^']+/[a-z0-9]+-([^']+)' from`)
)

type ProgressParser struct {
	expectedPackages  []ExpectedPackage
	seenPackages      map[string]bool
	totalBytes        int64
	completedBytes    int64
	lastPackageName   string
	evaluationPercent int
}

func NewProgressParser() *ProgressParser {
	return &ProgressParser{
		seenPackages:      make(map[string]bool),
		evaluationPercent: 5,
	}
}

func (p *ProgressParser) SetExpectedPackages(packages []ExpectedPackage) {
	p.expectedPackages = packages
	p.seenPackages = make(map[string]bool)
	p.totalBytes = 0
	p.completedBytes = 0

	for _, pkg := range packages {
		if pkg.SizeBytes > 0 {
			p.totalBytes += pkg.SizeBytes
		} else {
			p.totalBytes += 50 * 1024 * 1024
		}
	}
}

func (p *ProgressParser) Parse(line string) *BuildProgress {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	if evaluatingPattern.MatchString(line) {
		return &BuildProgress{
			Phase:           PhaseEvaluating,
			ProgressPercent: 0,
		}
	}

	if m := buildingPattern.FindStringSubmatch(line); m != nil {
		return p.handlePackage(extractPackageName(m[1]))
	}

	if m := copyingFromPattern.FindStringSubmatch(line); m != nil {
		return p.handlePackage(extractPackageName(m[1]))
	}

	return nil
}

func (p *ProgressParser) handlePackage(nixPkgName string) *BuildProgress {
	for _, expected := range p.expectedPackages {
		if p.seenPackages[expected.Name] {
			continue
		}

		if matchesExpected(nixPkgName, expected.Name) {
			p.seenPackages[expected.Name] = true
			if expected.SizeBytes > 0 {
				p.completedBytes += expected.SizeBytes
			} else {
				p.completedBytes += 50 * 1024 * 1024
			}
			p.lastPackageName = expected.DisplayName

			percent := p.calculatePercent()
			return &BuildProgress{
				Phase:           PhaseInstalling,
				PackageName:     expected.DisplayName,
				ProgressPercent: percent,
			}
		}
	}

	return nil
}

func (p *ProgressParser) calculatePercent() int {
	if p.totalBytes == 0 {
		return p.evaluationPercent
	}

	buildPercent := int(float64(p.completedBytes) / float64(p.totalBytes) * float64(100-p.evaluationPercent))
	return p.evaluationPercent + buildPercent
}

func matchesExpected(nixPkgName, expectedName string) bool {
	nixLower := strings.ToLower(nixPkgName)
	expectedLower := strings.ToLower(expectedName)

	if strings.HasPrefix(nixLower, expectedLower) {
		return true
	}

	if strings.HasPrefix(expectedLower, "libretro-") {
		coreName := strings.TrimPrefix(expectedLower, "libretro-")
		if strings.Contains(nixLower, "libretro") && strings.Contains(nixLower, coreName) {
			return true
		}
	}

	return false
}

func extractPackageName(nixStoreName string) string {
	parts := strings.Split(nixStoreName, "-")
	if len(parts) == 0 {
		return nixStoreName
	}

	for i, part := range parts {
		if !looksLikeVersion(part) {
			return strings.Join(parts[i:], "-")
		}
	}

	return nixStoreName
}

func looksLikeVersion(s string) bool {
	if s == "" {
		return false
	}
	return s[0] >= '0' && s[0] <= '9'
}
