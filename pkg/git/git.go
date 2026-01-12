// Package git provides a wrapper for git operations.
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Git provides git operations for a repository.
type Git struct {
	Dir    string // Repository directory
	Remote string // Remote name (default: origin)
}

// New creates a new Git instance for the given directory.
func New(dir string) *Git {
	return &Git{
		Dir:    dir,
		Remote: "origin",
	}
}

// Status represents the current git status.
type Status struct {
	Branch       string   // Current branch name
	Ahead        int      // Commits ahead of remote
	Behind       int      // Commits behind remote
	Staged       []string // Staged files
	Modified     []string // Modified but unstaged files
	Untracked    []string // Untracked files
	HasRemote    bool     // Whether branch has a remote tracking branch
	RemoteBranch string   // Remote tracking branch
	IsClean      bool     // No uncommitted changes
}

// LatestTag returns the most recent tag reachable from HEAD.
func (g *Git) LatestTag() (string, error) {
	output, err := g.run("describe", "--tags", "--abbrev=0")
	if err != nil {
		return "", fmt.Errorf("no tags found: %w", err)
	}
	return strings.TrimSpace(output), nil
}

// AllTags returns all tags in the repository, sorted by version.
func (g *Git) AllTags() ([]string, error) {
	output, err := g.run("tag", "--sort=-version:refname")
	if err != nil {
		return nil, err
	}
	if output == "" {
		return nil, nil
	}
	return strings.Split(strings.TrimSpace(output), "\n"), nil
}

// CreateTag creates a new tag at HEAD.
func (g *Git) CreateTag(tag string, message string, sign bool) error {
	args := []string{"tag"}
	if sign {
		args = append(args, "-s")
	}
	if message != "" {
		args = append(args, "-m", message)
	}
	args = append(args, tag)

	_, err := g.run(args...)
	if err != nil {
		return fmt.Errorf("failed to create tag %s: %w", tag, err)
	}
	return nil
}

// DeleteTag deletes a local tag.
func (g *Git) DeleteTag(tag string) error {
	_, err := g.run("tag", "-d", tag)
	return err
}

// Push pushes refs to the remote.
func (g *Git) Push(refs ...string) error {
	args := []string{"push", g.Remote}
	args = append(args, refs...)

	_, err := g.run(args...)
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	return nil
}

// PushTag pushes a specific tag to the remote.
func (g *Git) PushTag(tag string) error {
	_, err := g.run("push", g.Remote, tag)
	if err != nil {
		return fmt.Errorf("failed to push tag %s: %w", tag, err)
	}
	return nil
}

// PushWithUpstream pushes the current branch and sets upstream.
func (g *Git) PushWithUpstream() error {
	branch, err := g.CurrentBranch()
	if err != nil {
		return err
	}

	_, err = g.run("push", "-u", g.Remote, branch)
	if err != nil {
		return fmt.Errorf("failed to push with upstream: %w", err)
	}
	return nil
}

// CommitAll stages all changes and creates a commit.
func (g *Git) CommitAll(message string, sign bool) error {
	// Stage all changes
	_, err := g.run("add", "-A")
	if err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Create commit
	args := []string{"commit", "-m", message}
	if sign {
		args = append(args, "-S")
	}

	_, err = g.run(args...)
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

// Commit creates a commit with currently staged changes.
func (g *Git) Commit(message string, sign bool) error {
	args := []string{"commit", "-m", message}
	if sign {
		args = append(args, "-S")
	}

	_, err := g.run(args...)
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

// Status returns the current git status.
func (g *Git) Status() (*Status, error) {
	status := &Status{}

	// Get branch name
	branch, err := g.CurrentBranch()
	if err != nil {
		return nil, err
	}
	status.Branch = branch

	// Get status porcelain
	output, err := g.run("status", "--porcelain", "-b")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if i == 0 {
			// Parse branch line: ## branch...origin/branch [ahead N, behind M]
			status.parseBranchLine(line)
			continue
		}

		if len(line) < 3 {
			continue
		}

		xy := line[:2]
		file := strings.TrimSpace(line[3:])

		switch {
		case xy[0] != ' ' && xy[0] != '?':
			status.Staged = append(status.Staged, file)
		case xy[1] != ' ' && xy[1] != '?':
			status.Modified = append(status.Modified, file)
		case xy == "??":
			status.Untracked = append(status.Untracked, file)
		}
	}

	status.IsClean = len(status.Staged) == 0 && len(status.Modified) == 0

	return status, nil
}

func (s *Status) parseBranchLine(line string) {
	// Example: ## main...origin/main [ahead 2, behind 1]
	// or: ## main
	// or: ## HEAD (no branch)

	line = strings.TrimPrefix(line, "## ")

	// Check for tracking info
	if idx := strings.Index(line, "..."); idx != -1 {
		s.Branch = line[:idx]
		rest := line[idx+3:]

		// Parse remote branch and ahead/behind
		if spaceIdx := strings.Index(rest, " "); spaceIdx != -1 {
			s.RemoteBranch = rest[:spaceIdx]
			s.HasRemote = true

			// Parse ahead/behind
			aheadBehind := rest[spaceIdx:]
			aheadRe := regexp.MustCompile(`ahead (\d+)`)
			behindRe := regexp.MustCompile(`behind (\d+)`)

			if m := aheadRe.FindStringSubmatch(aheadBehind); m != nil {
				_, _ = fmt.Sscanf(m[1], "%d", &s.Ahead)
			}
			if m := behindRe.FindStringSubmatch(aheadBehind); m != nil {
				_, _ = fmt.Sscanf(m[1], "%d", &s.Behind)
			}
		} else {
			s.RemoteBranch = rest
			s.HasRemote = true
		}
	} else {
		s.Branch = line
	}
}

// IsDirty returns true if there are uncommitted changes.
func (g *Git) IsDirty() (bool, error) {
	output, err := g.run("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != "", nil
}

// CurrentBranch returns the current branch name.
func (g *Git) CurrentBranch() (string, error) {
	output, err := g.run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// CurrentCommit returns the current commit SHA.
func (g *Git) CurrentCommit() (string, error) {
	output, err := g.run("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// ShortCommit returns the short form of the current commit SHA.
func (g *Git) ShortCommit() (string, error) {
	output, err := g.run("rev-parse", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// RemoteURL returns the URL of the remote.
func (g *Git) RemoteURL() (string, error) {
	output, err := g.run("remote", "get-url", g.Remote)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// IsAncestor checks if ancestor is an ancestor of descendant.
func (g *Git) IsAncestor(ancestor, descendant string) (bool, error) {
	_, err := g.run("merge-base", "--is-ancestor", ancestor, descendant)
	if err != nil {
		// Exit code 1 means not an ancestor, which is not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Fetch fetches from the remote.
func (g *Git) Fetch() error {
	_, err := g.run("fetch", g.Remote)
	return err
}

// FetchTags fetches tags from the remote.
func (g *Git) FetchTags() error {
	_, err := g.run("fetch", "--tags", g.Remote)
	return err
}

// Log returns commit messages between two refs.
func (g *Git) Log(from, to string, format string) (string, error) {
	if format == "" {
		format = "%h %s"
	}
	ref := from + ".." + to
	output, err := g.run("log", "--format="+format, ref)
	if err != nil {
		return "", err
	}
	return output, nil
}

// run executes a git command and returns the output.
func (g *Git) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.Dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = stdout.String()
		}
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(errMsg))
	}

	return stdout.String(), nil
}
