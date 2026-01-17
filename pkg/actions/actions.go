// Package actions provides mutating operations for release preparation.
package actions

import (
	"github.com/agentplexus/release-agent-team/pkg/config"
)

// Result represents the result of an action.
type Result struct {
	Name    string
	Success bool
	Output  string
	Error   error
	Skipped bool
	Reason  string
}

// Proposal represents a proposed change for user approval.
type Proposal struct {
	Description string            // Human-readable description
	FilePath    string            // File to modify
	OldContent  string            // Current content (for diff)
	NewContent  string            // Proposed new content
	Metadata    map[string]string // Additional context
}

// Action is the interface for mutating operations.
type Action interface {
	// Name returns the action name.
	Name() string

	// Run executes the action directly (non-interactive).
	Run(dir string, opts Options) Result

	// Propose generates proposals for interactive mode.
	Propose(dir string, opts Options) ([]Proposal, error)

	// Apply applies approved proposals.
	Apply(dir string, proposals []Proposal) Result
}

// Options configures action behavior.
type Options struct {
	DryRun      bool           // Don't actually make changes
	Interactive bool           // Enable interactive mode
	Version     string         // Target version (for release)
	Since       string         // Since tag (for changelog)
	Verbose     bool           // Show detailed output
	Config      *config.Config // Configuration
}

// DefaultOptions returns the default action options.
func DefaultOptions() Options {
	return Options{
		DryRun:      false,
		Interactive: false,
		Verbose:     false,
	}
}
