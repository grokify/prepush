package workflow

import (
	"errors"
	"strings"
	"testing"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext("/tmp", "v1.0.0")

	if ctx.Dir != "/tmp" {
		t.Errorf("Dir = %s, want /tmp", ctx.Dir)
	}
	if ctx.Version != "v1.0.0" {
		t.Errorf("Version = %s, want v1.0.0", ctx.Version)
	}
	if ctx.Data == nil {
		t.Error("Data is nil, want initialized map")
	}
	if ctx.Output == nil {
		t.Error("Output is nil, want initialized builder")
	}
}

func TestContextLog(t *testing.T) {
	ctx := NewContext("/tmp", "v1.0.0")

	ctx.Log("Hello %s", "world")
	ctx.Log("Line 2")

	output := ctx.Output.String()
	if !strings.Contains(output, "Hello world") {
		t.Errorf("Output should contain 'Hello world', got: %s", output)
	}
	if !strings.Contains(output, "Line 2") {
		t.Errorf("Output should contain 'Line 2', got: %s", output)
	}
}

func TestNewRunner(t *testing.T) {
	runner := NewRunner()

	if runner.DryRun {
		t.Error("DryRun should be false by default")
	}
	if runner.Verbose {
		t.Error("Verbose should be false by default")
	}
	if runner.Interactive {
		t.Error("Interactive should be false by default")
	}
}

func TestRunnerRun_Success(t *testing.T) {
	wf := &Workflow{
		Name:        "Test Workflow",
		Description: "A test workflow",
		Steps: []Step{
			{
				Name:     "Step 1",
				Type:     StepTypeFunc,
				Required: true,
				Func: func(ctx *Context) error {
					ctx.Log("Step 1 executed")
					return nil
				},
			},
			{
				Name:     "Step 2",
				Type:     StepTypeFunc,
				Required: true,
				Func: func(ctx *Context) error {
					ctx.Log("Step 2 executed")
					return nil
				},
			},
		},
	}

	runner := NewRunner()
	ctx := NewContext("/tmp", "v1.0.0")
	result := runner.Run(wf, ctx)

	if !result.Success {
		t.Error("Workflow should succeed")
	}
	if len(result.Steps) != 2 {
		t.Errorf("Should have 2 step results, got %d", len(result.Steps))
	}
	for i, step := range result.Steps {
		if !step.Success {
			t.Errorf("Step %d should succeed", i+1)
		}
	}
}

func TestRunnerRun_RequiredStepFailure(t *testing.T) {
	wf := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				Name:     "Failing Step",
				Type:     StepTypeFunc,
				Required: true,
				Func: func(ctx *Context) error {
					return errors.New("intentional failure")
				},
			},
			{
				Name:     "Never Reached",
				Type:     StepTypeFunc,
				Required: true,
				Func: func(ctx *Context) error {
					t.Error("This step should not be executed")
					return nil
				},
			},
		},
	}

	runner := NewRunner()
	ctx := NewContext("/tmp", "v1.0.0")
	result := runner.Run(wf, ctx)

	if result.Success {
		t.Error("Workflow should fail when required step fails")
	}
	if len(result.Steps) != 1 {
		t.Errorf("Should stop after first failed step, got %d steps", len(result.Steps))
	}
}

func TestRunnerRun_OptionalStepFailure(t *testing.T) {
	step2Executed := false

	wf := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				Name:     "Optional Failing Step",
				Type:     StepTypeFunc,
				Required: false, // Not required
				Func: func(ctx *Context) error {
					return errors.New("intentional failure")
				},
			},
			{
				Name:     "Should Still Run",
				Type:     StepTypeFunc,
				Required: true,
				Func: func(ctx *Context) error {
					step2Executed = true
					return nil
				},
			},
		},
	}

	runner := NewRunner()
	ctx := NewContext("/tmp", "v1.0.0")
	result := runner.Run(wf, ctx)

	if !result.Success {
		t.Error("Workflow should succeed even when optional step fails")
	}
	if !step2Executed {
		t.Error("Second step should still execute after optional step failure")
	}
}

func TestRunnerRun_SkippedStep(t *testing.T) {
	wf := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				Name:     "No Function",
				Type:     StepTypeFunc,
				Required: true,
				Func:     nil, // No function defined
			},
		},
	}

	runner := NewRunner()
	ctx := NewContext("/tmp", "v1.0.0")
	result := runner.Run(wf, ctx)

	if !result.Success {
		t.Error("Workflow should succeed with skipped step")
	}
	if len(result.Steps) != 1 {
		t.Fatal("Should have 1 step result")
	}
	if !result.Steps[0].Skipped {
		t.Error("Step should be marked as skipped")
	}
}

func TestRunnerRun_CompositeStep(t *testing.T) {
	wf := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				Name:     "Composite",
				Type:     StepTypeComposite,
				Required: true,
				SubSteps: []Step{
					{
						Name:     "Sub 1",
						Type:     StepTypeFunc,
						Required: true,
						Func: func(ctx *Context) error {
							return nil
						},
					},
					{
						Name:     "Sub 2",
						Type:     StepTypeFunc,
						Required: true,
						Func: func(ctx *Context) error {
							return nil
						},
					},
				},
			},
		},
	}

	runner := NewRunner()
	ctx := NewContext("/tmp", "v1.0.0")
	result := runner.Run(wf, ctx)

	if !result.Success {
		t.Error("Workflow should succeed")
	}
	if len(result.Steps) != 1 {
		t.Fatal("Should have 1 top-level step result")
	}
	if len(result.Steps[0].SubSteps) != 2 {
		t.Errorf("Should have 2 sub-step results, got %d", len(result.Steps[0].SubSteps))
	}
}

func TestRunnerRun_DryRunPassedToContext(t *testing.T) {
	var capturedDryRun bool

	wf := &Workflow{
		Name: "Test",
		Steps: []Step{
			{
				Name: "Check DryRun",
				Type: StepTypeFunc,
				Func: func(ctx *Context) error {
					capturedDryRun = ctx.DryRun
					return nil
				},
			},
		},
	}

	runner := NewRunner()
	runner.DryRun = true

	ctx := NewContext("/tmp", "v1.0.0")
	runner.Run(wf, ctx)

	if !capturedDryRun {
		t.Error("DryRun should be passed to context")
	}
}

func TestWorkflowResultSummary(t *testing.T) {
	result := &WorkflowResult{
		Name:    "Test Workflow",
		Success: true,
		Steps: []StepResult{
			{Name: "Step 1", Success: true},
			{Name: "Step 2", Success: false},
			{Name: "Step 3", Skipped: true},
		},
	}

	summary := result.Summary()

	if !strings.Contains(summary, "Test Workflow") {
		t.Error("Summary should contain workflow name")
	}
	if !strings.Contains(summary, "Success") {
		t.Error("Summary should contain success status")
	}
	if !strings.Contains(summary, "Step 1") {
		t.Error("Summary should contain step names")
	}
}
