// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/agentplexus/release-agent-team/pkg/checks"
	"github.com/agentplexus/release-agent-team/pkg/config"
	"github.com/agentplexus/release-agent-team/pkg/detect"
	"github.com/agentplexus/release-agent-team/pkg/report"
)

// Validate command flags
var (
	validateVersion  string
	validateSkipQA   bool
	validateSkipDocs bool
	validateSkipSec  bool
	validateFormat   string
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [directory]",
	Short: "Run comprehensive release validation",
	Long: `Run comprehensive release validation across all areas of responsibility.

Validation Areas:
  QA            Build, tests, lint, format, error handling compliance
  Documentation README, PRD, TRD, release notes, CHANGELOG
  Release       Version availability, git status, CI configuration
  Security      LICENSE, vulnerability scan, dependency audit

Assumes Engineering and Product have already signed off.

Examples:
  releaseagent validate                    # Validate current directory
  releaseagent validate --version v0.2.0   # Include version-specific checks
  releaseagent validate --skip-qa          # Skip QA checks
  releaseagent validate --format team      # Team status report format
  releaseagent validate -v                 # Verbose output`,
	Args: cobra.MaximumNArgs(1),
	Run:  runValidate,
}

func init() {
	validateCmd.Flags().StringVar(&validateVersion, "version", "", "Target release version (e.g., v0.2.0)")
	validateCmd.Flags().BoolVar(&validateSkipQA, "skip-qa", false, "Skip QA checks")
	validateCmd.Flags().BoolVar(&validateSkipDocs, "skip-docs", false, "Skip documentation checks")
	validateCmd.Flags().BoolVar(&validateSkipSec, "skip-security", false, "Skip security checks")
	validateCmd.Flags().StringVar(&validateFormat, "format", "default", "Output format (default, team)")

	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) {
	// Get directory to validate
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	// Make sure directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory %s does not exist\n", dir)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: error loading config: %v\n", err)
	}

	// Override config with flags
	if cfgVerbose {
		cfg.Verbose = true
	}

	// Create validation report
	validationReport := &checks.ValidationReport{
		Version: validateVersion,
	}

	// Detect languages for QA checks
	detections, err := detect.Detect(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: error detecting languages: %v\n", err)
	}

	fmt.Println("╔════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                     RELEASE VALIDATION STARTING                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// QA Area
	if !validateSkipQA {
		fmt.Println("▶ Running QA validation...")
		qaResults := runQAChecks(dir, detections, &cfg)
		validationReport.Areas = append(validationReport.Areas, checks.AreaResult{
			Area:    checks.AreaQA,
			Status:  checks.ComputeAreaStatus(qaResults),
			Results: qaResults,
		})
	}

	// Documentation Area
	if !validateSkipDocs {
		fmt.Println("▶ Running Documentation validation...")
		docChecker := &checks.DocChecker{}
		docResults := docChecker.Check(dir, checks.DocOptions{
			Version: validateVersion,
			Verbose: cfg.Verbose,
		})
		validationReport.Areas = append(validationReport.Areas, checks.AreaResult{
			Area:    checks.AreaDocumentation,
			Status:  checks.ComputeAreaStatus(docResults),
			Results: docResults,
		})
	}

	// Release Management Area
	fmt.Println("▶ Running Release Management validation...")
	releaseChecker := &checks.ReleaseChecker{}
	releaseResults := releaseChecker.Check(dir, checks.ReleaseOptions{
		Version: validateVersion,
		Verbose: cfg.Verbose,
	})
	validationReport.Areas = append(validationReport.Areas, checks.AreaResult{
		Area:    checks.AreaRelease,
		Status:  checks.ComputeAreaStatus(releaseResults),
		Results: releaseResults,
	})

	// Security Area
	if !validateSkipSec {
		fmt.Println("▶ Running Security validation...")
		secChecker := &checks.SecurityChecker{}
		secResults := secChecker.Check(dir, checks.SecurityOptions{
			Verbose: cfg.Verbose,
		})
		validationReport.Areas = append(validationReport.Areas, checks.AreaResult{
			Area:    checks.AreaSecurity,
			Status:  checks.ComputeAreaStatus(secResults),
			Results: secResults,
		})
	}

	// Print comprehensive report
	if validateFormat == "team" {
		printTeamStatusReport(validationReport, dir)
	} else {
		checks.PrintValidationReport(validationReport)
	}

	// Exit with error if validation failed
	if !validationReport.IsGo() {
		os.Exit(1)
	}
}

// printTeamStatusReport prints the validation report in team status format.
func printTeamStatusReport(vr *checks.ValidationReport, dir string) {
	// Determine project name from git remote or directory
	project := dir
	if dir == "." {
		if cwd, err := os.Getwd(); err == nil {
			project = cwd
		}
	}

	// Try to get git remote URL
	if gitProject := getGitRemoteProject(); gitProject != "" {
		project = gitProject
	}

	// Build target string
	target := vr.Version
	if target == "" {
		target = "release validation"
	}

	// Convert to team status report
	teamReport := report.FromValidationReport(vr, project, target, "RELEASE VALIDATION")

	// Render the report
	renderer := report.NewRenderer(os.Stdout)
	if err := renderer.Render(teamReport); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering report: %v\n", err)
	}
}

// getGitRemoteProject extracts the project path from git remote origin.
func getGitRemoteProject() string {
	// This is a simplified version - just return empty for now
	// A full implementation would parse .git/config or run git remote get-url
	return ""
}

// runQAChecks runs all QA checks for detected languages.
func runQAChecks(dir string, detections []detect.Detection, cfg *config.Config) []checks.Result {
	var results []checks.Result

	// Go checks
	if detect.HasLanguage(detections, detect.Go) && cfg.IsLanguageEnabled("go") {
		langCfg := cfg.GetLanguageConfig("go")
		opts := checks.Options{
			Test:              *langCfg.Test,
			Lint:              *langCfg.Lint,
			Format:            *langCfg.Format,
			Coverage:          langCfg.Coverage != nil && *langCfg.Coverage,
			Verbose:           cfg.Verbose,
			GoExcludeCoverage: langCfg.ExcludeCoverage,
		}
		if opts.GoExcludeCoverage == "" {
			opts.GoExcludeCoverage = "cmd"
		}

		checker := &checks.GoChecker{}
		for _, d := range detect.GetByLanguage(detections, detect.Go) {
			results = append(results, checker.Check(d.Path, opts)...)
		}
	}

	// TypeScript checks
	if detect.HasLanguage(detections, detect.TypeScript) && cfg.IsLanguageEnabled("typescript") {
		langCfg := cfg.GetLanguageConfig("typescript")
		opts := checks.Options{
			Test:    *langCfg.Test,
			Lint:    *langCfg.Lint,
			Format:  *langCfg.Format,
			Verbose: cfg.Verbose,
		}

		checker := &checks.TypeScriptChecker{}
		for _, d := range detect.GetByLanguage(detections, detect.TypeScript) {
			results = append(results, checker.Check(d.Path, opts)...)
		}
	}

	// JavaScript checks
	if detect.HasLanguage(detections, detect.JavaScript) && cfg.IsLanguageEnabled("javascript") {
		langCfg := cfg.GetLanguageConfig("javascript")
		opts := checks.Options{
			Test:    *langCfg.Test,
			Lint:    *langCfg.Lint,
			Format:  *langCfg.Format,
			Verbose: cfg.Verbose,
		}

		checker := &checks.TypeScriptChecker{}
		for _, d := range detect.GetByLanguage(detections, detect.JavaScript) {
			results = append(results, checker.Check(d.Path, opts)...)
		}
	}

	return results
}
