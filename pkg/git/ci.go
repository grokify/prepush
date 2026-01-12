package git

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// CIStatus represents the combined status of CI checks.
type CIStatus struct {
	State       string        // "success", "pending", "failure", "error"
	TotalCount  int           // Total number of checks
	Statuses    []CheckStatus // Individual check statuses
	CheckSuites []CheckSuite  // GitHub Actions check suites
}

// CheckStatus represents a single status check.
type CheckStatus struct {
	Context     string // Check name/context
	State       string // "success", "pending", "failure", "error"
	Description string // Status description
	TargetURL   string // URL to check details
}

// CheckSuite represents a GitHub Actions check suite.
type CheckSuite struct {
	App        string // App name (e.g., "GitHub Actions")
	Status     string // "queued", "in_progress", "completed"
	Conclusion string // "success", "failure", "neutral", etc. (only if completed)
}

// ghCombinedStatus is the structure returned by gh api for combined status.
type ghCombinedStatus struct {
	State    string `json:"state"`
	Statuses []struct {
		Context     string `json:"context"`
		State       string `json:"state"`
		Description string `json:"description"`
		TargetURL   string `json:"target_url"`
	} `json:"statuses"`
	TotalCount int `json:"total_count"`
}

// ghCheckRuns is the structure returned by gh api for check runs.
type ghCheckRuns struct {
	TotalCount int `json:"total_count"`
	CheckRuns  []struct {
		Name       string `json:"name"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		App        struct {
			Name string `json:"name"`
		} `json:"app"`
	} `json:"check_runs"`
}

// GetCIStatus retrieves the CI status for a commit.
func (g *Git) GetCIStatus(ref string) (*CIStatus, error) {
	if !commandExists("gh") {
		return nil, fmt.Errorf("gh CLI not found in PATH")
	}

	// Get repository info
	owner, repo, err := g.parseRemoteURL()
	if err != nil {
		return nil, err
	}

	if ref == "" {
		ref, err = g.CurrentCommit()
		if err != nil {
			return nil, err
		}
	}

	status := &CIStatus{
		State: "pending",
	}

	// Get combined status (legacy status checks)
	combinedOutput, err := g.runGH("api", fmt.Sprintf("repos/%s/%s/commits/%s/status", owner, repo, ref))
	if err == nil {
		var combined ghCombinedStatus
		if err := json.Unmarshal([]byte(combinedOutput), &combined); err == nil {
			status.TotalCount = combined.TotalCount
			status.State = combined.State
			for _, s := range combined.Statuses {
				status.Statuses = append(status.Statuses, CheckStatus{
					Context:     s.Context,
					State:       s.State,
					Description: s.Description,
					TargetURL:   s.TargetURL,
				})
			}
		}
	}

	// Get check runs (GitHub Actions)
	checksOutput, err := g.runGH("api", fmt.Sprintf("repos/%s/%s/commits/%s/check-runs", owner, repo, ref))
	if err == nil {
		var checks ghCheckRuns
		if err := json.Unmarshal([]byte(checksOutput), &checks); err == nil {
			for _, run := range checks.CheckRuns {
				status.CheckSuites = append(status.CheckSuites, CheckSuite{
					App:        run.App.Name,
					Status:     run.Status,
					Conclusion: run.Conclusion,
				})

				// Add to statuses for unified view
				state := "pending"
				if run.Status == "completed" {
					switch run.Conclusion {
					case "success", "skipped", "neutral":
						state = "success"
					case "failure", "timed_out", "cancelled":
						state = "failure"
					default:
						state = run.Conclusion
					}
				}

				status.Statuses = append(status.Statuses, CheckStatus{
					Context: run.Name,
					State:   state,
				})
			}
		}
	}

	// Calculate overall state from all checks
	status.State = calculateOverallState(status.Statuses)

	return status, nil
}

// WaitForCI waits for CI to complete with a timeout.
func (g *Git) WaitForCI(timeout time.Duration) error {
	if !commandExists("gh") {
		return fmt.Errorf("gh CLI not found in PATH")
	}

	ref, err := g.CurrentCommit()
	if err != nil {
		return err
	}

	deadline := time.Now().Add(timeout)
	pollInterval := 10 * time.Second

	for time.Now().Before(deadline) {
		status, err := g.GetCIStatus(ref)
		if err != nil {
			return err
		}

		switch status.State {
		case "success":
			return nil
		case "failure", "error":
			return fmt.Errorf("CI failed with state: %s", status.State)
		}

		// Still pending, wait and retry
		time.Sleep(pollInterval)
	}

	return fmt.Errorf("CI timeout after %v", timeout)
}

// IsCIPassing checks if CI is currently passing (without waiting).
func (g *Git) IsCIPassing(ref string) (bool, error) {
	status, err := g.GetCIStatus(ref)
	if err != nil {
		return false, err
	}
	return status.State == "success", nil
}

// parseRemoteURL extracts owner and repo from the remote URL.
func (g *Git) parseRemoteURL() (owner string, repo string, err error) {
	url, err := g.RemoteURL()
	if err != nil {
		return "", "", err
	}

	// Handle SSH format: git@github.com:owner/repo.git
	sshRegex := regexp.MustCompile(`git@github\.com:([^/]+)/(.+?)(?:\.git)?$`)
	if m := sshRegex.FindStringSubmatch(url); m != nil {
		return m[1], m[2], nil
	}

	// Handle HTTPS format: https://github.com/owner/repo.git
	httpsRegex := regexp.MustCompile(`https://github\.com/([^/]+)/(.+?)(?:\.git)?$`)
	if m := httpsRegex.FindStringSubmatch(url); m != nil {
		return m[1], m[2], nil
	}

	return "", "", fmt.Errorf("could not parse GitHub URL: %s", url)
}

// runGH executes a gh command and returns the output.
func (g *Git) runGH(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	cmd.Dir = g.Dir

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// calculateOverallState determines the overall CI state from individual checks.
func calculateOverallState(statuses []CheckStatus) string {
	if len(statuses) == 0 {
		return "pending"
	}

	hasFailure := false
	hasPending := false

	for _, s := range statuses {
		switch s.State {
		case "failure", "error":
			hasFailure = true
		case "pending":
			hasPending = true
		}
	}

	if hasFailure {
		return "failure"
	}
	if hasPending {
		return "pending"
	}
	return "success"
}

// commandExists checks if a command is available in PATH.
func commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// GetPRForBranch gets the PR number for the current branch.
func (g *Git) GetPRForBranch() (int, error) {
	if !commandExists("gh") {
		return 0, fmt.Errorf("gh CLI not found in PATH")
	}

	branch, err := g.CurrentBranch()
	if err != nil {
		return 0, err
	}

	output, err := g.runGH("pr", "view", branch, "--json", "number")
	if err != nil {
		return 0, fmt.Errorf("no PR found for branch %s", branch)
	}

	var result struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return 0, err
	}

	return result.Number, nil
}

// GetPRStatus gets the CI status for a PR.
func (g *Git) GetPRStatus(prNumber int) (*CIStatus, error) {
	if !commandExists("gh") {
		return nil, fmt.Errorf("gh CLI not found in PATH")
	}

	output, err := g.runGH("pr", "checks", fmt.Sprintf("%d", prNumber), "--json", "name,state,conclusion")
	if err != nil {
		return nil, err
	}

	var checks []struct {
		Name       string `json:"name"`
		State      string `json:"state"`
		Conclusion string `json:"conclusion"`
	}
	if err := json.Unmarshal([]byte(output), &checks); err != nil {
		return nil, err
	}

	status := &CIStatus{
		TotalCount: len(checks),
	}

	for _, c := range checks {
		state := strings.ToLower(c.State)
		if c.Conclusion != "" {
			state = strings.ToLower(c.Conclusion)
		}

		// Normalize state names
		switch state {
		case "success", "skipped", "neutral":
			state = "success"
		case "failure", "timed_out", "cancelled", "action_required":
			state = "failure"
		case "pending", "queued", "in_progress", "waiting":
			state = "pending"
		}

		status.Statuses = append(status.Statuses, CheckStatus{
			Context: c.Name,
			State:   state,
		})
	}

	status.State = calculateOverallState(status.Statuses)

	return status, nil
}
