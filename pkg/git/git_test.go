package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	g := New("/tmp/test")

	if g.Dir != "/tmp/test" {
		t.Errorf("Dir = %s, want /tmp/test", g.Dir)
	}
	if g.Remote != "origin" {
		t.Errorf("Remote = %s, want origin", g.Remote)
	}
}

func TestStatusParseBranchLine(t *testing.T) {
	tests := []struct {
		name          string
		line          string
		wantBranch    string
		wantRemote    string
		wantHasRemote bool
		wantAhead     int
		wantBehind    int
	}{
		{
			name:       "simple branch",
			line:       "## main",
			wantBranch: "main",
		},
		{
			name:          "with remote",
			line:          "## main...origin/main",
			wantBranch:    "main",
			wantRemote:    "origin/main",
			wantHasRemote: true,
		},
		{
			name:          "ahead only",
			line:          "## main...origin/main [ahead 2]",
			wantBranch:    "main",
			wantRemote:    "origin/main",
			wantHasRemote: true,
			wantAhead:     2,
		},
		{
			name:          "behind only",
			line:          "## main...origin/main [behind 3]",
			wantBranch:    "main",
			wantRemote:    "origin/main",
			wantHasRemote: true,
			wantBehind:    3,
		},
		{
			name:          "ahead and behind",
			line:          "## feature...origin/feature [ahead 1, behind 2]",
			wantBranch:    "feature",
			wantRemote:    "origin/feature",
			wantHasRemote: true,
			wantAhead:     1,
			wantBehind:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Status{}
			s.parseBranchLine(tt.line)

			if s.Branch != tt.wantBranch {
				t.Errorf("Branch = %s, want %s", s.Branch, tt.wantBranch)
			}
			if s.RemoteBranch != tt.wantRemote {
				t.Errorf("RemoteBranch = %s, want %s", s.RemoteBranch, tt.wantRemote)
			}
			if s.HasRemote != tt.wantHasRemote {
				t.Errorf("HasRemote = %v, want %v", s.HasRemote, tt.wantHasRemote)
			}
			if s.Ahead != tt.wantAhead {
				t.Errorf("Ahead = %d, want %d", s.Ahead, tt.wantAhead)
			}
			if s.Behind != tt.wantBehind {
				t.Errorf("Behind = %d, want %d", s.Behind, tt.wantBehind)
			}
		})
	}
}

func TestCalculateOverallState(t *testing.T) {
	tests := []struct {
		name     string
		statuses []CheckStatus
		want     string
	}{
		{
			name:     "empty",
			statuses: nil,
			want:     "pending",
		},
		{
			name: "all success",
			statuses: []CheckStatus{
				{Context: "build", State: "success"},
				{Context: "test", State: "success"},
			},
			want: "success",
		},
		{
			name: "one failure",
			statuses: []CheckStatus{
				{Context: "build", State: "success"},
				{Context: "test", State: "failure"},
			},
			want: "failure",
		},
		{
			name: "one pending",
			statuses: []CheckStatus{
				{Context: "build", State: "success"},
				{Context: "test", State: "pending"},
			},
			want: "pending",
		},
		{
			name: "failure takes precedence over pending",
			statuses: []CheckStatus{
				{Context: "build", State: "pending"},
				{Context: "test", State: "failure"},
			},
			want: "failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateOverallState(tt.statuses)
			if got != tt.want {
				t.Errorf("calculateOverallState() = %s, want %s", got, tt.want)
			}
		})
	}
}

// Integration tests that require a real git repo
func TestGitIntegration(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Initialize a git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user for the test repo
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git name: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	g := New(tmpDir)

	t.Run("IsDirty", func(t *testing.T) {
		dirty, err := g.IsDirty()
		if err != nil {
			t.Fatalf("IsDirty() error: %v", err)
		}
		if !dirty {
			t.Error("IsDirty() = false, want true (untracked file exists)")
		}
	})

	t.Run("CommitAll", func(t *testing.T) {
		err := g.CommitAll("Initial commit", false)
		if err != nil {
			t.Fatalf("CommitAll() error: %v", err)
		}

		dirty, _ := g.IsDirty()
		if dirty {
			t.Error("IsDirty() = true after commit, want false")
		}
	})

	t.Run("CurrentBranch", func(t *testing.T) {
		branch, err := g.CurrentBranch()
		if err != nil {
			t.Fatalf("CurrentBranch() error: %v", err)
		}
		// Could be "main" or "master" depending on git config
		if branch != "main" && branch != "master" {
			t.Errorf("CurrentBranch() = %s, want main or master", branch)
		}
	})

	t.Run("CurrentCommit", func(t *testing.T) {
		commit, err := g.CurrentCommit()
		if err != nil {
			t.Fatalf("CurrentCommit() error: %v", err)
		}
		if len(commit) != 40 {
			t.Errorf("CurrentCommit() length = %d, want 40", len(commit))
		}
	})

	t.Run("ShortCommit", func(t *testing.T) {
		commit, err := g.ShortCommit()
		if err != nil {
			t.Fatalf("ShortCommit() error: %v", err)
		}
		if len(commit) < 7 {
			t.Errorf("ShortCommit() length = %d, want >= 7", len(commit))
		}
	})

	t.Run("CreateTag", func(t *testing.T) {
		err := g.CreateTag("v0.1.0", "Test tag", false)
		if err != nil {
			t.Fatalf("CreateTag() error: %v", err)
		}

		tag, err := g.LatestTag()
		if err != nil {
			t.Fatalf("LatestTag() error: %v", err)
		}
		if tag != "v0.1.0" {
			t.Errorf("LatestTag() = %s, want v0.1.0", tag)
		}
	})

	t.Run("AllTags", func(t *testing.T) {
		tags, err := g.AllTags()
		if err != nil {
			t.Fatalf("AllTags() error: %v", err)
		}
		if len(tags) != 1 || tags[0] != "v0.1.0" {
			t.Errorf("AllTags() = %v, want [v0.1.0]", tags)
		}
	})

	t.Run("Status", func(t *testing.T) {
		status, err := g.Status()
		if err != nil {
			t.Fatalf("Status() error: %v", err)
		}
		if !status.IsClean {
			t.Error("Status.IsClean = false, want true")
		}
	})
}
