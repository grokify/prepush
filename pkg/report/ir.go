// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package report

import (
	"encoding/json"
	"time"
)

// AgentValidationResult is the JSON-serializable output from each validation agent.
// This is the intermediate representation that agents produce and the Release
// Coordinator consumes to build the final TeamStatusReportIR.
type AgentValidationResult struct {
	// Schema is the JSON schema URL for validation
	Schema string `json:"$schema,omitempty"`

	// AgentID identifies the agent (e.g., "pm", "qa", "documentation")
	AgentID string `json:"agent_id"`

	// StepID is the workflow step name (e.g., "pm-validation", "qa-validation")
	StepID string `json:"step_id"`

	// Inputs are values received from upstream agents in the DAG
	// For example, QA receives {"version": "v0.4.0"} from PM
	Inputs map[string]interface{} `json:"inputs,omitempty"`

	// Outputs are values produced by this agent for downstream agents
	// For example, PM outputs {"version-recommendation": "v0.4.0"}
	Outputs map[string]interface{} `json:"outputs,omitempty"`

	// Checks are the individual validation checks performed
	Checks []CheckIR `json:"checks"`

	// Status is the overall status for this agent (computed from checks)
	Status Status `json:"status"`

	// ExecutedAt is when the agent completed execution
	ExecutedAt time.Time `json:"executed_at"`

	// AgentModel is the LLM model used (e.g., "sonnet", "haiku", "opus")
	AgentModel string `json:"agent_model,omitempty"`

	// Duration is how long the agent took to execute
	Duration string `json:"duration,omitempty"`

	// Error is set if the agent failed to execute
	Error string `json:"error,omitempty"`
}

// CheckIR is a JSON-serializable check result.
type CheckIR struct {
	// ID is the check identifier (e.g., "build", "tests", "version-recommendation")
	ID string `json:"id"`

	// Status is GO, WARN, NO-GO, or SKIP
	Status Status `json:"status"`

	// Detail is optional additional information
	Detail string `json:"detail,omitempty"`

	// Metadata allows checks to include structured data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TeamIR is a JSON-serializable team/agent section in the report.
type TeamIR struct {
	// ID is the workflow step ID (e.g., "pm-validation")
	ID string `json:"id"`

	// Name is the agent name (e.g., "pm")
	Name string `json:"name"`

	// AgentID matches the agent definition in team.json
	AgentID string `json:"agent_id"`

	// Model is the LLM model used
	Model string `json:"model,omitempty"`

	// Checks are the validation checks for this team
	Checks []CheckIR `json:"checks"`

	// Status is the overall status (computed from checks)
	Status Status `json:"status"`
}

// TeamStatusReportIR is the complete JSON-serializable report.
// This is what the Release Coordinator produces by aggregating AgentValidationResults.
type TeamStatusReportIR struct {
	// Schema is the JSON schema URL for validation
	Schema string `json:"$schema,omitempty"`

	// Project is the repository identifier
	Project string `json:"project"`

	// Version is the target release version
	Version string `json:"version"`

	// Target is a human-readable target description
	Target string `json:"target"`

	// Phase is the workflow phase (e.g., "RELEASE VALIDATION")
	Phase string `json:"phase"`

	// Teams are the validation teams/agents
	Teams []TeamIR `json:"teams"`

	// Status is the overall status (computed from teams)
	Status Status `json:"status"`

	// GeneratedAt is when the report was generated
	GeneratedAt time.Time `json:"generated_at"`

	// GeneratedBy identifies the coordinator
	GeneratedBy string `json:"generated_by,omitempty"`
}

// ComputeStatus computes the overall status from checks.
func (a *AgentValidationResult) ComputeStatus() Status {
	hasNoGo := false
	hasWarn := false
	allSkipped := true

	for _, c := range a.Checks {
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

// ToTeamIR converts an AgentValidationResult to a TeamIR for the report.
func (a *AgentValidationResult) ToTeamIR() TeamIR {
	return TeamIR{
		ID:      a.StepID,
		Name:    a.AgentID,
		AgentID: a.AgentID,
		Model:   a.AgentModel,
		Checks:  a.Checks,
		Status:  a.ComputeStatus(),
	}
}

// AggregateResults combines multiple AgentValidationResults into a TeamStatusReportIR.
func AggregateResults(results []AgentValidationResult, project, version, phase string) *TeamStatusReportIR {
	teams := make([]TeamIR, 0, len(results))
	for _, r := range results {
		teams = append(teams, r.ToTeamIR())
	}

	report := &TeamStatusReportIR{
		Schema:      "https://agentplexus.github.io/release-agent-team/schemas/team-status-report.json",
		Project:     project,
		Version:     version,
		Target:      version,
		Phase:       phase,
		Teams:       teams,
		GeneratedAt: time.Now().UTC(),
		GeneratedBy: "release-coordinator",
	}

	// Compute overall status
	report.Status = report.ComputeOverallStatus()

	return report
}

// ComputeOverallStatus computes the overall status from all teams.
func (r *TeamStatusReportIR) ComputeOverallStatus() Status {
	hasNoGo := false
	hasWarn := false

	for _, t := range r.Teams {
		switch t.Status {
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

// ToJSON serializes the report to JSON.
func (r *TeamStatusReportIR) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// ToTeamStatusReport converts the IR to the internal TeamStatusReport for rendering.
func (r *TeamStatusReportIR) ToTeamStatusReport() *TeamStatusReport {
	teams := make([]Team, 0, len(r.Teams))
	for _, t := range r.Teams {
		checks := make([]Check, 0, len(t.Checks))
		for _, c := range t.Checks {
			checks = append(checks, Check{
				ID:     c.ID,
				Status: c.Status,
				Detail: c.Detail,
			})
		}
		teams = append(teams, Team{
			ID:     t.ID,
			Name:   t.Name,
			Checks: checks,
		})
	}

	return &TeamStatusReport{
		Project: r.Project,
		Version: r.Version,
		Target:  r.Target,
		Phase:   r.Phase,
		Teams:   teams,
	}
}

// ParseAgentResult parses JSON into an AgentValidationResult.
func ParseAgentResult(data []byte) (*AgentValidationResult, error) {
	var result AgentValidationResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MarshalJSON implements json.Marshaler for Status.
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON implements json.Unmarshaler for Status.
func (s *Status) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = Status(str)
	return nil
}
