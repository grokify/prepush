// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package report

import (
	"bytes"
	"strings"
	"testing"
)

func TestStatusIcon(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusGo, "\U0001F7E2"},   // 🟢
		{StatusWarn, "\U0001F7E1"}, // 🟡
		{StatusNoGo, "\U0001F534"}, // 🔴
		{StatusSkip, "\u26AA"},     // ⚪
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.Icon(); got != tt.want {
				t.Errorf("Status(%q).Icon() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestTeamOverallStatus(t *testing.T) {
	tests := []struct {
		name   string
		checks []Check
		want   Status
	}{
		{
			name:   "all GO",
			checks: []Check{{Status: StatusGo}, {Status: StatusGo}},
			want:   StatusGo,
		},
		{
			name:   "one NO-GO",
			checks: []Check{{Status: StatusGo}, {Status: StatusNoGo}},
			want:   StatusNoGo,
		},
		{
			name:   "one WARN",
			checks: []Check{{Status: StatusGo}, {Status: StatusWarn}},
			want:   StatusWarn,
		},
		{
			name:   "all SKIP",
			checks: []Check{{Status: StatusSkip}, {Status: StatusSkip}},
			want:   StatusSkip,
		},
		{
			name:   "NO-GO takes precedence over WARN",
			checks: []Check{{Status: StatusWarn}, {Status: StatusNoGo}},
			want:   StatusNoGo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			team := Team{Checks: tt.checks}
			if got := team.OverallStatus(); got != tt.want {
				t.Errorf("Team.OverallStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReportIsGo(t *testing.T) {
	tests := []struct {
		name  string
		teams []Team
		want  bool
	}{
		{
			name: "all teams GO",
			teams: []Team{
				{Checks: []Check{{Status: StatusGo}}},
				{Checks: []Check{{Status: StatusGo}}},
			},
			want: true,
		},
		{
			name: "one team WARN is still GO",
			teams: []Team{
				{Checks: []Check{{Status: StatusGo}}},
				{Checks: []Check{{Status: StatusWarn}}},
			},
			want: true,
		},
		{
			name: "one team NO-GO",
			teams: []Team{
				{Checks: []Check{{Status: StatusGo}}},
				{Checks: []Check{{Status: StatusNoGo}}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &TeamStatusReport{Teams: tt.teams}
			if got := report.IsGo(); got != tt.want {
				t.Errorf("Report.IsGo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderer(t *testing.T) {
	report := &TeamStatusReport{
		Project: "github.com/agentplexus/release-agent-team",
		Version: "v0.3.0",
		Target:  "v0.3.0 (release automation)",
		Phase:   "PHASE 1: REVIEW",
		Teams: []Team{
			{
				ID:   "qa-validation",
				Name: "qa",
				Checks: []Check{
					{ID: "build", Status: StatusGo, Detail: ""},
					{ID: "tests", Status: StatusGo, Detail: "42 tests passed"},
					{ID: "lint", Status: StatusGo, Detail: ""},
				},
			},
			{
				ID:   "security-validation",
				Name: "security",
				Checks: []Check{
					{ID: "license", Status: StatusGo, Detail: "MIT License"},
					{ID: "vulnerability-scan", Status: StatusWarn, Detail: "1 deprecated"},
				},
			},
		},
	}

	var buf bytes.Buffer
	renderer := NewRenderer(&buf)
	err := renderer.Render(report)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()

	// Check for expected content
	expectedContent := []string{
		"TEAM STATUS REPORT",
		"github.com/agentplexus/release-agent-team",
		"v0.3.0 (release automation)",
		"PHASE 1: REVIEW",
		"qa-validation (qa)",
		"build",
		"42 tests passed",
		"security-validation (security)",
		"MIT License",
		"1 deprecated",
		"GO for v0.3.0",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(output, expected) {
			t.Errorf("Output missing expected content: %q", expected)
		}
	}

	// Check box structure
	if !strings.Contains(output, "╔") {
		t.Error("Output missing top border")
	}
	if !strings.Contains(output, "╚") {
		t.Error("Output missing bottom border")
	}
	if !strings.Contains(output, "║") {
		t.Error("Output missing side borders")
	}
}

func TestRendererNoGo(t *testing.T) {
	report := &TeamStatusReport{
		Version: "v0.3.0",
		Teams: []Team{
			{
				ID:   "qa-validation",
				Name: "qa",
				Checks: []Check{
					{ID: "build", Status: StatusNoGo, Detail: "compilation failed"},
				},
			},
		},
	}

	var buf bytes.Buffer
	renderer := NewRenderer(&buf)
	err := renderer.Render(report)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "NO-GO for v0.3.0") {
		t.Error("Output should contain NO-GO message")
	}
}

func TestVisualLength(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"hello world", 11},
		{"\U0001F7E2", 2}, // 🟢 - emoji counts as 2
		{"GO \U0001F7E2", 5},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := visualLength(tt.input); got != tt.want {
				t.Errorf("visualLength(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// ExampleRenderer demonstrates the report output format.
func ExampleRenderer() {
	report := &TeamStatusReport{
		Project: "github.com/grokify/example",
		Version: "v1.0.0",
		Target:  "v1.0.0 (initial release)",
		Phase:   "PHASE 1: REVIEW",
		Teams: []Team{
			{
				ID:   "qa-validation",
				Name: "qa",
				Checks: []Check{
					{ID: "build", Status: StatusGo, Detail: ""},
					{ID: "tests", Status: StatusGo, Detail: "10 tests passed"},
				},
			},
		},
	}

	var buf bytes.Buffer
	renderer := NewRenderer(&buf)
	if err := renderer.Render(report); err != nil {
		panic(err)
	}
	// Output format is validated by the test
}
