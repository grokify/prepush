// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package checks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PMChecker validates product management concerns for a release.
type PMChecker struct{}

// PMOptions contains options for PM validation.
type PMOptions struct {
	Version string // Target version (e.g., "v0.5.0")
	Verbose bool
}

// Check runs all PM validation checks.
func (c *PMChecker) Check(dir string, opts PMOptions) []Result {
	var results []Result

	// 1. Version recommendation
	results = append(results, c.checkVersionRecommendation(dir, opts.Version))

	// 2. Release scope
	results = append(results, c.checkReleaseScope(dir, opts.Version))

	// 3. Changelog quality
	results = append(results, c.checkChangelogQuality(dir, opts.Version))

	// 4. Breaking changes
	results = append(results, c.checkBreakingChanges(dir, opts.Version))

	// 5. Roadmap alignment
	results = append(results, c.checkRoadmapAlignment(dir, opts.Version))

	// 6. Deprecation notices
	results = append(results, c.checkDeprecationNotices(dir, opts.Version))

	return results
}

// checkVersionRecommendation validates the version follows semver and is appropriate.
func (c *PMChecker) checkVersionRecommendation(dir, version string) Result {
	name := "PM: version-recommendation"

	if version == "" {
		return Result{
			Name:    name,
			Passed:  false,
			Warning: true,
			Reason:  "No version specified",
		}
	}

	// Validate semver format
	semverRegex := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`)
	if !semverRegex.MatchString(version) {
		return Result{
			Name:   name,
			Passed: false,
			Reason: fmt.Sprintf("Version %s does not follow semver format", version),
		}
	}

	// Determine version type
	versionType := "patch"
	parts := semverRegex.FindStringSubmatch(version)
	if len(parts) >= 3 {
		if parts[1] != "0" && parts[2] == "0" && parts[3] == "0" {
			versionType = "major"
		} else if parts[2] != "0" && parts[3] == "0" {
			versionType = "minor (feature release)"
		}
	}

	return Result{
		Name:   name,
		Passed: true,
		Output: fmt.Sprintf("%s appropriate for %s", version, versionType),
	}
}

// checkReleaseScope validates the release scope matches expectations.
func (c *PMChecker) checkReleaseScope(dir, version string) Result {
	name := "PM: release-scope"

	// Check CHANGELOG.json for the version entry
	changelogPath := filepath.Join(dir, "CHANGELOG.json")
	data, err := os.ReadFile(changelogPath)
	if err != nil {
		return Result{
			Name:    name,
			Passed:  false,
			Warning: true,
			Reason:  "CHANGELOG.json not found",
		}
	}

	var changelog struct {
		Releases []struct {
			Version    string `json:"version"`
			Highlights []struct {
				Description string `json:"description"`
			} `json:"highlights"`
			Added   []interface{} `json:"added"`
			Changed []interface{} `json:"changed"`
			Fixed   []interface{} `json:"fixed"`
		} `json:"releases"`
	}

	if err := json.Unmarshal(data, &changelog); err != nil {
		return Result{
			Name:    name,
			Passed:  false,
			Warning: true,
			Reason:  "Failed to parse CHANGELOG.json",
		}
	}

	// Find the version entry
	for _, release := range changelog.Releases {
		if release.Version == version {
			totalChanges := len(release.Added) + len(release.Changed) + len(release.Fixed)
			return Result{
				Name:   name,
				Passed: true,
				Output: fmt.Sprintf("%d changes documented", totalChanges),
			}
		}
	}

	return Result{
		Name:    name,
		Passed:  false,
		Warning: true,
		Reason:  fmt.Sprintf("Version %s not found in CHANGELOG.json", version),
	}
}

// checkChangelogQuality validates the changelog has highlights and proper descriptions.
func (c *PMChecker) checkChangelogQuality(dir, version string) Result {
	name := "PM: changelog-quality"

	changelogPath := filepath.Join(dir, "CHANGELOG.json")
	data, err := os.ReadFile(changelogPath)
	if err != nil {
		return Result{
			Name:    name,
			Passed:  false,
			Warning: true,
			Reason:  "CHANGELOG.json not found",
		}
	}

	var changelog struct {
		Releases []struct {
			Version    string `json:"version"`
			Highlights []struct {
				Description string `json:"description"`
			} `json:"highlights"`
		} `json:"releases"`
	}

	if err := json.Unmarshal(data, &changelog); err != nil {
		return Result{
			Name:    name,
			Passed:  false,
			Warning: true,
			Reason:  "Failed to parse CHANGELOG.json",
		}
	}

	// Find the version entry
	for _, release := range changelog.Releases {
		if release.Version == version {
			if len(release.Highlights) == 0 {
				return Result{
					Name:    name,
					Passed:  false,
					Warning: true,
					Reason:  "No highlights for this release",
				}
			}
			return Result{
				Name:   name,
				Passed: true,
				Output: fmt.Sprintf("%d highlights present", len(release.Highlights)),
			}
		}
	}

	return Result{
		Name:    name,
		Passed:  false,
		Warning: true,
		Reason:  fmt.Sprintf("Version %s not found in CHANGELOG.json", version),
	}
}

// checkBreakingChanges validates breaking changes are properly documented.
func (c *PMChecker) checkBreakingChanges(dir, version string) Result {
	name := "PM: breaking-changes"

	changelogPath := filepath.Join(dir, "CHANGELOG.json")
	data, err := os.ReadFile(changelogPath)
	if err != nil {
		return Result{
			Name:    name,
			Passed:  false,
			Warning: true,
			Reason:  "CHANGELOG.json not found",
		}
	}

	var changelog struct {
		Releases []struct {
			Version string `json:"version"`
			Changed []struct {
				Description string `json:"description"`
				Breaking    bool   `json:"breaking"`
			} `json:"changed"`
		} `json:"releases"`
	}

	if err := json.Unmarshal(data, &changelog); err != nil {
		return Result{
			Name:    name,
			Passed:  false,
			Warning: true,
			Reason:  "Failed to parse CHANGELOG.json",
		}
	}

	// Find the version entry and count breaking changes
	for _, release := range changelog.Releases {
		if release.Version == version {
			breakingCount := 0
			for _, change := range release.Changed {
				if change.Breaking {
					breakingCount++
				}
			}

			if breakingCount == 0 {
				return Result{
					Name:   name,
					Passed: true,
					Output: "No breaking changes",
				}
			}

			return Result{
				Name:   name,
				Passed: true,
				Output: fmt.Sprintf("%d breaking changes documented", breakingCount),
			}
		}
	}

	return Result{
		Name:   name,
		Passed: true,
		Output: "No breaking changes (version not in changelog)",
	}
}

// checkRoadmapAlignment validates the release aligns with roadmap items.
func (c *PMChecker) checkRoadmapAlignment(dir, version string) Result {
	name := "PM: roadmap-alignment"

	roadmapPath := filepath.Join(dir, "ROADMAP.md")
	data, err := os.ReadFile(roadmapPath)
	if err != nil {
		return Result{
			Name:    name,
			Passed:  false,
			Warning: true,
			Reason:  "ROADMAP.md not found",
		}
	}

	content := string(data)

	// Count completed vs total items for this version
	// Look for patterns like "**Version:** 0.5.0" after "### [x]" or "### [ ]"
	versionNum := strings.TrimPrefix(version, "v")

	completedPattern := regexp.MustCompile(`### \[x\][^\n]+\n[^\n]*\n\*\*Version:\*\* ` + regexp.QuoteMeta(versionNum))
	pendingPattern := regexp.MustCompile(`### \[ \][^\n]+\n[^\n]*\n\*\*Version:\*\* ` + regexp.QuoteMeta(versionNum))

	completed := len(completedPattern.FindAllString(content, -1))
	pending := len(pendingPattern.FindAllString(content, -1))
	total := completed + pending

	if total == 0 {
		return Result{
			Name:    name,
			Passed:  true,
			Warning: true,
			Output:  fmt.Sprintf("No roadmap items tagged for %s", version),
		}
	}

	if pending > 0 {
		return Result{
			Name:    name,
			Passed:  false,
			Warning: true,
			Reason:  fmt.Sprintf("%d/%d roadmap items completed (%d pending)", completed, total, pending),
		}
	}

	return Result{
		Name:   name,
		Passed: true,
		Output: fmt.Sprintf("%d/%d items completed", completed, total),
	}
}

// checkDeprecationNotices validates deprecated features are properly documented.
func (c *PMChecker) checkDeprecationNotices(dir, version string) Result {
	name := "PM: deprecation-notices"

	changelogPath := filepath.Join(dir, "CHANGELOG.json")
	data, err := os.ReadFile(changelogPath)
	if err != nil {
		return Result{
			Name:   name,
			Passed: true,
			Output: "No deprecations (CHANGELOG.json not found)",
		}
	}

	var changelog struct {
		Releases []struct {
			Version    string `json:"version"`
			Deprecated []struct {
				Description string `json:"description"`
			} `json:"deprecated"`
		} `json:"releases"`
	}

	if err := json.Unmarshal(data, &changelog); err != nil {
		return Result{
			Name:   name,
			Passed: true,
			Output: "No deprecations (could not parse changelog)",
		}
	}

	// Find the version entry
	for _, release := range changelog.Releases {
		if release.Version == version {
			if len(release.Deprecated) == 0 {
				return Result{
					Name:   name,
					Passed: true,
					Output: "No deprecations",
				}
			}
			return Result{
				Name:   name,
				Passed: true,
				Output: fmt.Sprintf("%d deprecation notices", len(release.Deprecated)),
			}
		}
	}

	return Result{
		Name:   name,
		Passed: true,
		Output: "No deprecations",
	}
}
