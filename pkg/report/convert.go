// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package report

import (
	"fmt"
	"strings"

	"github.com/agentplexus/release-agent-team/pkg/checks"
)

// TeamConfig maps validation areas to team IDs and names.
type TeamConfig struct {
	Area checks.ValidationArea
	ID   string
	Name string
}

// DefaultTeamConfigs returns the default team configurations.
func DefaultTeamConfigs() []TeamConfig {
	return []TeamConfig{
		{Area: checks.AreaPM, ID: "pm-validation", Name: "pm"},
		{Area: checks.AreaQA, ID: "qa-validation", Name: "qa"},
		{Area: checks.AreaDocumentation, ID: "docs-validation", Name: "documentation"},
		{Area: checks.AreaSecurity, ID: "security-validation", Name: "security"},
		{Area: checks.AreaRelease, ID: "release-validation", Name: "release"},
	}
}

// FromValidationReport converts a checks.ValidationReport to a TeamStatusReport.
func FromValidationReport(vr *checks.ValidationReport, project, target, phase string) *TeamStatusReport {
	configs := DefaultTeamConfigs()
	configMap := make(map[checks.ValidationArea]TeamConfig)
	for _, c := range configs {
		configMap[c.Area] = c
	}

	var teams []Team
	for _, ar := range vr.Areas {
		config, ok := configMap[ar.Area]
		if !ok {
			config = TeamConfig{
				ID:   strings.ToLower(string(ar.Area)) + "-validation",
				Name: strings.ToLower(string(ar.Area)),
			}
		}

		var teamChecks []Check
		for _, r := range ar.Results {
			status := StatusGo
			if r.Skipped {
				status = StatusSkip
			} else if r.Warning && !r.Passed {
				status = StatusWarn
			} else if !r.Passed {
				status = StatusNoGo
			}

			// Extract check ID from name (e.g., "Go: build" -> "build")
			id := r.Name
			if idx := strings.Index(id, ": "); idx >= 0 {
				id = id[idx+2:]
			}
			// Convert to kebab-case
			id = strings.ToLower(strings.ReplaceAll(id, " ", "-"))

			// Use output as detail, truncate if needed
			detail := ""
			if r.Output != "" {
				detail = r.Output
				// Take first line only
				if idx := strings.Index(detail, "\n"); idx >= 0 {
					detail = detail[:idx]
				}
				// Truncate
				if len(detail) > 40 {
					detail = detail[:37] + "..."
				}
			}
			if r.Reason != "" && detail == "" {
				detail = r.Reason
			}

			teamChecks = append(teamChecks, Check{
				ID:     id,
				Status: status,
				Detail: detail,
			})
		}

		teams = append(teams, Team{
			ID:     config.ID,
			Name:   config.Name,
			Checks: teamChecks,
		})
	}

	return &TeamStatusReport{
		Project: project,
		Version: vr.Version,
		Target:  target,
		Phase:   phase,
		Teams:   teams,
	}
}

// PMTeam creates a Product Management validation team.
func PMTeam(version string, roadmapTotal, roadmapCompleted int, hasHighlights, hasBreaking, hasDeprecations bool) Team {
	checks := []Check{
		{
			ID:     "version-recommendation",
			Status: StatusGo,
			Detail: version + " appropriate",
		},
		{
			ID:     "release-scope",
			Status: StatusGo,
			Detail: "Phase complete",
		},
		{
			ID:     "changelog-quality",
			Status: boolToStatus(hasHighlights),
			Detail: boolToDetail(hasHighlights, "Highlights present", "Missing highlights"),
		},
		{
			ID:     "breaking-changes",
			Status: boolToStatus(!hasBreaking),
			Detail: boolToDetail(!hasBreaking, "None", "Has breaking changes"),
		},
		{
			ID:     "roadmap-alignment",
			Status: StatusGo,
			Detail: formatFraction(roadmapCompleted, roadmapTotal) + " items completed",
		},
		{
			ID:     "deprecation-notices",
			Status: boolToStatus(!hasDeprecations),
			Detail: boolToDetail(!hasDeprecations, "No deprecations", "Has deprecations"),
		},
	}

	return Team{
		ID:     "pm-validation",
		Name:   "pm",
		Checks: checks,
	}
}

func boolToStatus(ok bool) Status {
	if ok {
		return StatusGo
	}
	return StatusWarn
}

func boolToDetail(ok bool, okDetail, notOkDetail string) string {
	if ok {
		return okDetail
	}
	return notOkDetail
}

func formatFraction(num, total int) string {
	return fmt.Sprintf("%d/%d", num, total)
}
