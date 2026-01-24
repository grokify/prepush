// Package kiro provides embedded Kiro CLI agent and steering files.
package kiro

import "embed"

// AgentFiles contains the embedded Kiro agent JSON files.
//
//go:embed agents/*.json
var AgentFiles embed.FS

// SteeringFiles contains the embedded Kiro steering markdown files.
//
//go:embed steering/*.md
var SteeringFiles embed.FS
