package interactive

import (
	"bytes"
	"strings"
	"testing"

	"github.com/grokify/release-agent/pkg/actions"
)

func TestQuestionTypeString(t *testing.T) {
	tests := []struct {
		qt   QuestionType
		want string
	}{
		{QuestionTypeSingleChoice, "single_choice"},
		{QuestionTypeMultiChoice, "multi_choice"},
		{QuestionTypeConfirm, "confirm"},
		{QuestionTypeText, "text"},
		{QuestionType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.qt.String(); got != tt.want {
			t.Errorf("QuestionType(%d).String() = %s, want %s", tt.qt, got, tt.want)
		}
	}
}

func TestProposalActionString(t *testing.T) {
	tests := []struct {
		pa   ProposalAction
		want string
	}{
		{ProposalActionApply, "apply"},
		{ProposalActionSkip, "skip"},
		{ProposalActionEdit, "edit"},
		{ProposalActionAbort, "abort"},
		{ProposalAction(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.pa.String(); got != tt.want {
			t.Errorf("ProposalAction(%d).String() = %s, want %s", tt.pa, got, tt.want)
		}
	}
}

// MockPrompter implements Prompter for testing.
type MockPrompter struct {
	AskFunc          func(q Question) (Answer, error)
	ShowProposalFunc func(p actions.Proposal) error
	ConfirmFunc      func(message string) (bool, error)
	Messages         []string
}

func (m *MockPrompter) Ask(q Question) (Answer, error) {
	if m.AskFunc != nil {
		return m.AskFunc(q)
	}
	return Answer{QuestionID: q.ID, Selected: []string{"apply"}}, nil
}

func (m *MockPrompter) ShowProposal(p actions.Proposal) error {
	if m.ShowProposalFunc != nil {
		return m.ShowProposalFunc(p)
	}
	return nil
}

func (m *MockPrompter) Confirm(message string) (bool, error) {
	if m.ConfirmFunc != nil {
		return m.ConfirmFunc(message)
	}
	return true, nil
}

func (m *MockPrompter) Info(message string) {
	m.Messages = append(m.Messages, "info: "+message)
}

func (m *MockPrompter) Warn(message string) {
	m.Messages = append(m.Messages, "warn: "+message)
}

func (m *MockPrompter) Error(message string) {
	m.Messages = append(m.Messages, "error: "+message)
}

func TestReviewProposal_Apply(t *testing.T) {
	mock := &MockPrompter{
		AskFunc: func(q Question) (Answer, error) {
			return Answer{QuestionID: q.ID, Selected: []string{"apply"}}, nil
		},
	}

	proposal := actions.Proposal{
		Description: "Test proposal",
		FilePath:    "test.txt",
	}

	action, err := ReviewProposal(mock, proposal)
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	if action != ProposalActionApply {
		t.Errorf("ReviewProposal() = %v, want %v", action, ProposalActionApply)
	}
}

func TestReviewProposal_Skip(t *testing.T) {
	mock := &MockPrompter{
		AskFunc: func(q Question) (Answer, error) {
			return Answer{QuestionID: q.ID, Selected: []string{"skip"}}, nil
		},
	}

	proposal := actions.Proposal{Description: "Test proposal"}

	action, err := ReviewProposal(mock, proposal)
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	if action != ProposalActionSkip {
		t.Errorf("ReviewProposal() = %v, want %v", action, ProposalActionSkip)
	}
}

func TestReviewProposal_Abort(t *testing.T) {
	mock := &MockPrompter{
		AskFunc: func(q Question) (Answer, error) {
			return Answer{QuestionID: q.ID, Selected: []string{"abort"}}, nil
		},
	}

	proposal := actions.Proposal{Description: "Test proposal"}

	action, err := ReviewProposal(mock, proposal)
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	if action != ProposalActionAbort {
		t.Errorf("ReviewProposal() = %v, want %v", action, ProposalActionAbort)
	}
}

func TestReviewProposal_Default(t *testing.T) {
	mock := &MockPrompter{
		AskFunc: func(q Question) (Answer, error) {
			return Answer{QuestionID: q.ID, Selected: []string{}}, nil
		},
	}

	proposal := actions.Proposal{Description: "Test proposal"}

	action, err := ReviewProposal(mock, proposal)
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	if action != ProposalActionApply {
		t.Errorf("ReviewProposal() = %v, want %v (default)", action, ProposalActionApply)
	}
}

func TestQuestion(t *testing.T) {
	q := Question{
		ID:      "test",
		Text:    "Test question?",
		Type:    QuestionTypeSingleChoice,
		Options: []Option{{ID: "a", Label: "Option A"}},
		Default: "a",
		Context: "Some context",
	}

	if q.ID != "test" {
		t.Errorf("Question.ID = %s, want test", q.ID)
	}
	if len(q.Options) != 1 {
		t.Errorf("Question.Options length = %d, want 1", len(q.Options))
	}
}

func TestOption(t *testing.T) {
	opt := Option{
		ID:          "opt1",
		Label:       "Option 1",
		Description: "First option",
	}

	if opt.ID != "opt1" {
		t.Errorf("Option.ID = %s, want opt1", opt.ID)
	}
	if opt.Label != "Option 1" {
		t.Errorf("Option.Label = %s, want Option 1", opt.Label)
	}
}

func TestAnswer(t *testing.T) {
	ans := Answer{
		QuestionID: "q1",
		Selected:   []string{"opt1", "opt2"},
		Text:       "custom text",
		Confirmed:  true,
	}

	if ans.QuestionID != "q1" {
		t.Errorf("Answer.QuestionID = %s, want q1", ans.QuestionID)
	}
	if len(ans.Selected) != 2 {
		t.Errorf("Answer.Selected length = %d, want 2", len(ans.Selected))
	}
	if !ans.Confirmed {
		t.Error("Answer.Confirmed = false, want true")
	}
}

func TestJSONPrompter_Ask(t *testing.T) {
	// Prepare input (answer JSON)
	input := `{"question_id": "test-q", "selected": ["opt1"]}` + "\n"
	reader := strings.NewReader(input)

	var output bytes.Buffer
	prompter := NewJSONPrompter(&output, reader)

	q := Question{
		ID:   "test-q",
		Text: "Test question?",
		Type: QuestionTypeSingleChoice,
		Options: []Option{
			{ID: "opt1", Label: "Option 1"},
			{ID: "opt2", Label: "Option 2"},
		},
	}

	answer, err := prompter.Ask(q)
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}

	if answer.QuestionID != "test-q" {
		t.Errorf("QuestionID = %s, want test-q", answer.QuestionID)
	}
	if len(answer.Selected) != 1 || answer.Selected[0] != "opt1" {
		t.Errorf("Selected = %v, want [opt1]", answer.Selected)
	}

	// Check output contains JSON question
	outStr := output.String()
	if !strings.Contains(outStr, `"type": "question"`) {
		t.Error("Output should contain question JSON")
	}
}

func TestJSONPrompter_ShowProposal(t *testing.T) {
	var output bytes.Buffer
	reader := strings.NewReader("")
	prompter := NewJSONPrompter(&output, reader)

	proposal := actions.Proposal{
		Description: "Test proposal",
		FilePath:    "test.txt",
		NewContent:  "new content",
	}

	err := prompter.ShowProposal(proposal)
	if err != nil {
		t.Fatalf("ShowProposal() error = %v", err)
	}

	outStr := output.String()
	if !strings.Contains(outStr, `"type": "proposal"`) {
		t.Error("Output should contain proposal JSON")
	}
	if !strings.Contains(outStr, `"waiting_for": "user_approval"`) {
		t.Error("Output should contain waiting_for field")
	}
}

func TestJSONPrompter_Confirm(t *testing.T) {
	input := `{"question_id": "confirm", "confirmed": true}` + "\n"
	reader := strings.NewReader(input)

	var output bytes.Buffer
	prompter := NewJSONPrompter(&output, reader)

	result, err := prompter.Confirm("Continue?")
	if err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}

	if !result {
		t.Error("Confirm() = false, want true")
	}
}

func TestJSONPrompter_InfoWarnError(t *testing.T) {
	var output bytes.Buffer
	reader := strings.NewReader("")
	prompter := NewJSONPrompter(&output, reader)

	prompter.Info("info message")
	prompter.Warn("warn message")
	prompter.Error("error message")

	outStr := output.String()
	if !strings.Contains(outStr, `"type": "info"`) {
		t.Error("Output should contain info message")
	}
	if !strings.Contains(outStr, `"type": "warning"`) {
		t.Error("Output should contain warning message")
	}
	if !strings.Contains(outStr, `"type": "error"`) {
		t.Error("Output should contain error message")
	}
}

func TestDefaultJSONPrompter(t *testing.T) {
	prompter := DefaultJSONPrompter()
	if prompter == nil {
		t.Error("DefaultJSONPrompter() returned nil")
	}
}
