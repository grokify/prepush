// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package report provides structured team status report generation
// for release validation with consistent, templated output.
package report

// Status represents the validation status of a check.
type Status string

const (
	StatusGo   Status = "GO"
	StatusWarn Status = "WARN"
	StatusNoGo Status = "NO-GO"
	StatusSkip Status = "SKIP"
)

// Icon returns the UTF-8 icon for the status.
func (s Status) Icon() string {
	switch s {
	case StatusGo:
		return "\U0001F7E2" // 🟢
	case StatusWarn:
		return "\U0001F7E1" // 🟡
	case StatusNoGo:
		return "\U0001F534" // 🔴
	case StatusSkip:
		return "\u26AA" // ⚪
	default:
		return "?"
	}
}

// Check represents a single validation check result.
type Check struct {
	ID     string // e.g., "build", "tests", "lint"
	Status Status // GO, WARN, NO-GO, SKIP
	Detail string // Optional detail (e.g., "8 tests passed")
}

// Team represents a validation team/area with its checks.
type Team struct {
	ID     string  // e.g., "qa-validation"
	Name   string  // e.g., "qa"
	Checks []Check // Individual checks for this team
}

// OverallStatus computes the overall status for the team.
func (t *Team) OverallStatus() Status {
	hasNoGo := false
	hasWarn := false
	allSkipped := true

	for _, c := range t.Checks {
		if c.Status != StatusSkip {
			allSkipped = false
		}
		switch c.Status {
		case StatusNoGo:
			hasNoGo = true
		case StatusWarn:
			hasWarn = true
		}
	}

	if allSkipped {
		return StatusSkip
	}
	if hasNoGo {
		return StatusNoGo
	}
	if hasWarn {
		return StatusWarn
	}
	return StatusGo
}

// TeamStatusReport represents the complete validation report.
type TeamStatusReport struct {
	Project string // e.g., "github.com/agentplexus/release-agent-team"
	Version string // e.g., "v0.3.0"
	Target  string // e.g., "v0.3.0 (release automation platform)"
	Phase   string // e.g., "PHASE 1: REVIEW"
	Teams   []Team // Validation teams
}

// IsGo returns true if all teams pass validation.
func (r *TeamStatusReport) IsGo() bool {
	for _, t := range r.Teams {
		if t.OverallStatus() == StatusNoGo {
			return false
		}
	}
	return true
}

// OverallStatus returns the overall status for the report.
func (r *TeamStatusReport) OverallStatus() Status {
	hasNoGo := false
	hasWarn := false

	for _, t := range r.Teams {
		switch t.OverallStatus() {
		case StatusNoGo:
			hasNoGo = true
		case StatusWarn:
			hasWarn = true
		}
	}

	if hasNoGo {
		return StatusNoGo
	}
	if hasWarn {
		return StatusWarn
	}
	return StatusGo
}

// FinalMessage returns the final status message.
func (r *TeamStatusReport) FinalMessage() string {
	if r.IsGo() {
		return "\U0001F680 TEAM: GO for " + r.Version + " \U0001F680" // 🚀 TEAM: GO for vX.Y.Z 🚀
	}
	return "\U0001F6D1 TEAM: NO-GO for " + r.Version + " \U0001F6D1" // 🛑 TEAM: NO-GO for vX.Y.Z 🛑
}
