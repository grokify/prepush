package checks

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// GoChecker implements checks for Go projects.
type GoChecker struct{}

// Name returns the checker name.
func (c *GoChecker) Name() string {
	return "Go"
}

// Check runs Go checks on the specified directory.
func (c *GoChecker) Check(dir string, opts Options) []Result {
	var results []Result

	// Check for local replace directives
	results = append(results, c.checkNoLocalReplace(dir))

	// Check go mod tidy
	results = append(results, c.checkModTidy(dir))

	// Check build
	results = append(results, c.checkBuild(dir))

	// Format check
	if opts.Format {
		results = append(results, c.checkFormat(dir))
	}

	// Lint check
	if opts.Lint {
		results = append(results, c.checkLint(dir))
	}

	// Test check
	if opts.Test {
		results = append(results, c.checkTest(dir))
	}

	// Soft checks (warnings, don't fail build)
	// Untracked file references
	results = append(results, c.checkUntrackedReferences(dir))

	// Coverage check
	if opts.Coverage {
		results = append(results, c.checkCoverage(dir, opts.GoExcludeCoverage))
	}

	return results
}

func (c *GoChecker) checkNoLocalReplace(dir string) Result {
	name := "Go: no local replace directives"

	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return Result{
			Name:   name,
			Passed: false,
			Error:  err,
		}
	}

	// Check for local paths in replace directives
	// Local replaces typically have paths starting with . or /
	localReplacePattern := regexp.MustCompile(`"Path":\s*"[./]`)
	if localReplacePattern.Match(output) {
		return Result{
			Name:   name,
			Passed: false,
			Output: "go.mod contains local replace directives",
		}
	}

	return Result{
		Name:   name,
		Passed: true,
	}
}

func (c *GoChecker) checkFormat(dir string) Result {
	name := "Go: gofmt"

	// Check if any files need formatting
	cmd := exec.Command("gofmt", "-l", ".")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return Result{
			Name:   name,
			Passed: false,
			Error:  err,
		}
	}

	unformatted := strings.TrimSpace(string(output))
	if unformatted != "" {
		return Result{
			Name:   name,
			Passed: false,
			Output: "Files need formatting:\n" + unformatted,
		}
	}

	return Result{
		Name:   name,
		Passed: true,
	}
}

func (c *GoChecker) checkLint(dir string) Result {
	name := "Go: golangci-lint"

	if !CommandExists("golangci-lint") {
		return Result{
			Name:    name,
			Skipped: true,
			Reason:  "golangci-lint not installed",
		}
	}

	return RunCommand(name, dir, "golangci-lint", "run")
}

func (c *GoChecker) checkTest(dir string) Result {
	name := "Go: tests"
	return RunCommand(name, dir, "go", "test", "./...")
}

func (c *GoChecker) checkCoverage(dir string, exclude string) Result {
	name := "Go: coverage"

	if !CommandExists("gocoverbadge") {
		return Result{
			Name:    name,
			Skipped: true,
			Reason:  "gocoverbadge not installed",
		}
	}

	args := []string{"-dir", dir, "-badge-only"}
	if exclude != "" {
		args = append(args, "-exclude", exclude)
	}

	result := RunCommand(name, dir, "gocoverbadge", args...)
	// Coverage is informational (soft check)
	result.Warning = true
	result.Passed = true
	return result
}

func (c *GoChecker) checkModTidy(dir string) Result {
	name := "Go: mod tidy"

	// Run go mod tidy -diff to check if go.mod/go.sum need updating
	// This is available in Go 1.21+
	cmd := exec.Command("go", "mod", "tidy", "-diff")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()

	// If -diff is not supported, fall back to checking manually
	if err != nil && strings.Contains(string(output), "unknown flag") {
		// Fall back: run go mod tidy and check for changes
		return c.checkModTidyFallback(dir)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr != "" {
		return Result{
			Name:   name,
			Passed: false,
			Output: "go.mod or go.sum needs updating. Run: go mod tidy",
		}
	}

	return Result{
		Name:   name,
		Passed: true,
	}
}

func (c *GoChecker) checkModTidyFallback(dir string) Result {
	name := "Go: mod tidy"

	// Get current state of go.mod and go.sum
	cmd := exec.Command("git", "diff", "--quiet", "go.mod", "go.sum")
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		return Result{
			Name:   name,
			Passed: false,
			Output: "go.mod or go.sum has uncommitted changes",
		}
	}

	// Run go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = dir
	if err := tidyCmd.Run(); err != nil {
		return Result{
			Name:   name,
			Passed: false,
			Error:  err,
		}
	}

	// Check if anything changed
	checkCmd := exec.Command("git", "diff", "--quiet", "go.mod", "go.sum")
	checkCmd.Dir = dir
	err = checkCmd.Run()
	if err != nil {
		// Restore the original files
		restoreCmd := exec.Command("git", "checkout", "go.mod", "go.sum")
		restoreCmd.Dir = dir
		_ = restoreCmd.Run() // Best effort restore, ignore error

		return Result{
			Name:   name,
			Passed: false,
			Output: "go.mod or go.sum needs updating. Run: go mod tidy",
		}
	}

	return Result{
		Name:   name,
		Passed: true,
	}
}

func (c *GoChecker) checkBuild(dir string) Result {
	name := "Go: build"
	return RunCommand(name, dir, "go", "build", "./...")
}

func (c *GoChecker) checkUntrackedReferences(dir string) Result {
	name := "Go: untracked references"

	// Get list of untracked files
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return Result{
			Name:    name,
			Warning: true,
			Passed:  true, // Can't check, so pass
		}
	}

	untrackedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(untrackedFiles) == 0 || (len(untrackedFiles) == 1 && untrackedFiles[0] == "") {
		return Result{
			Name:    name,
			Warning: true,
			Passed:  true,
		}
	}

	// Get list of tracked files
	trackedCmd := exec.Command("git", "ls-files")
	trackedCmd.Dir = dir
	trackedOutput, err := trackedCmd.Output()
	if err != nil {
		return Result{
			Name:    name,
			Warning: true,
			Passed:  true,
		}
	}

	trackedFiles := strings.Split(strings.TrimSpace(string(trackedOutput)), "\n")

	// Check if any tracked file references an untracked file
	var references []string
	for _, tracked := range trackedFiles {
		// Only check Go files and go.mod
		if !strings.HasSuffix(tracked, ".go") && tracked != "go.mod" {
			continue
		}

		for _, untracked := range untrackedFiles {
			if untracked == "" {
				continue
			}
			// Simple check: see if the untracked filename appears in the tracked file
			// This is a heuristic and may have false positives
			baseName := strings.TrimSuffix(untracked, ".go")
			if strings.Contains(baseName, "/") {
				parts := strings.Split(baseName, "/")
				baseName = parts[len(parts)-1]
			}

			// Skip common patterns that are likely false positives
			if baseName == "main" || baseName == "test" || baseName == "doc" {
				continue
			}

			grepCmd := exec.Command("grep", "-l", baseName, tracked)
			grepCmd.Dir = dir
			if grepOutput, err := grepCmd.Output(); err == nil && len(grepOutput) > 0 {
				references = append(references, fmt.Sprintf("%s may reference untracked %s", tracked, untracked))
			}
		}
	}

	if len(references) > 0 {
		return Result{
			Name:    name,
			Warning: true,
			Passed:  false,
			Output:  strings.Join(references, "\n"),
		}
	}

	return Result{
		Name:    name,
		Warning: true,
		Passed:  true,
	}
}
