package checks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if !opts.Test {
		t.Error("expected Test to be true by default")
	}
	if !opts.Lint {
		t.Error("expected Lint to be true by default")
	}
	if !opts.Format {
		t.Error("expected Format to be true by default")
	}
	if opts.Coverage {
		t.Error("expected Coverage to be false by default")
	}
	if opts.GoExcludeCoverage != "cmd" {
		t.Errorf("expected GoExcludeCoverage to be 'cmd', got %s", opts.GoExcludeCoverage)
	}
}

func TestRunCommand_Success(t *testing.T) {
	result := RunCommand("test", ".", "echo", "hello")

	if !result.Passed {
		t.Error("expected command to pass")
	}
	if result.Output != "hello" {
		t.Errorf("expected output 'hello', got %q", result.Output)
	}
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestRunCommand_Failure(t *testing.T) {
	result := RunCommand("test", ".", "false")

	if result.Passed {
		t.Error("expected command to fail")
	}
	if result.Error == nil {
		t.Error("expected error")
	}
}

func TestRunCommand_NotFound(t *testing.T) {
	result := RunCommand("test", ".", "nonexistent-command-12345")

	if result.Passed {
		t.Error("expected command to fail")
	}
	if result.Error == nil {
		t.Error("expected error for non-existent command")
	}
}

func TestCommandExists(t *testing.T) {
	// echo should exist on all systems
	if !CommandExists("echo") {
		t.Error("expected 'echo' command to exist")
	}

	// This should not exist
	if CommandExists("nonexistent-command-12345") {
		t.Error("expected fake command to not exist")
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")

	// File doesn't exist yet
	if FileExists(testFile) {
		t.Error("expected file to not exist")
	}

	// Create file
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	// Now it should exist
	if !FileExists(testFile) {
		t.Error("expected file to exist")
	}
}

func TestPrintResults(t *testing.T) {
	results := []Result{
		{Name: "test1", Passed: true},
		{Name: "test2", Passed: false, Output: "failed"},
		{Name: "test3", Skipped: true, Reason: "not configured"},
	}

	passed, failed, skipped, warnings := PrintResults(results, false)

	if passed != 1 {
		t.Errorf("expected 1 passed, got %d", passed)
	}
	if failed != 1 {
		t.Errorf("expected 1 failed, got %d", failed)
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
	if warnings != 0 {
		t.Errorf("expected 0 warnings, got %d", warnings)
	}
}

func TestPrintResults_Warnings(t *testing.T) {
	results := []Result{
		{Name: "test1", Passed: true},
		{Name: "test2", Warning: true, Passed: false, Output: "something to check"},
		{Name: "test3", Warning: true, Passed: true},
	}

	passed, failed, skipped, warnings := PrintResults(results, false)

	if passed != 2 {
		t.Errorf("expected 2 passed, got %d", passed)
	}
	if failed != 0 {
		t.Errorf("expected 0 failed, got %d", failed)
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}
	if warnings != 1 {
		t.Errorf("expected 1 warning, got %d", warnings)
	}
}
