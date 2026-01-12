// Package interactive provides user interaction support for release-agent.
package interactive

import (
	"github.com/grokify/release-agent/pkg/actions"
)

// QuestionType defines the type of question.
type QuestionType int

const (
	// QuestionTypeSingleChoice allows selecting one option.
	QuestionTypeSingleChoice QuestionType = iota
	// QuestionTypeMultiChoice allows selecting multiple options.
	QuestionTypeMultiChoice
	// QuestionTypeConfirm is a yes/no question.
	QuestionTypeConfirm
	// QuestionTypeText allows free-form text input.
	QuestionTypeText
)

// String returns the string representation of the question type.
func (qt QuestionType) String() string {
	switch qt {
	case QuestionTypeSingleChoice:
		return "single_choice"
	case QuestionTypeMultiChoice:
		return "multi_choice"
	case QuestionTypeConfirm:
		return "confirm"
	case QuestionTypeText:
		return "text"
	default:
		return "unknown"
	}
}

// Option represents a choice option for questions.
type Option struct {
	ID          string // Unique identifier
	Label       string // Display text
	Description string // Optional description
}

// Question represents a question for the user.
type Question struct {
	ID      string       // Unique identifier
	Text    string       // The question text
	Type    QuestionType // Type of question
	Options []Option     // Available options (for choice types)
	Default string       // Default value or option ID
	Context string       // Additional context (e.g., code snippet)
}

// Answer represents a user's response to a question.
type Answer struct {
	QuestionID string   // ID of the question being answered
	Selected   []string // Selected option IDs (for choice types)
	Text       string   // Text response (for text type)
	Confirmed  bool     // Response for confirm type
}

// Prompter handles user interaction.
type Prompter interface {
	// Ask presents a question and returns the user's answer.
	Ask(q Question) (Answer, error)

	// ShowProposal displays a proposed change for review.
	ShowProposal(p actions.Proposal) error

	// Confirm asks a yes/no question.
	Confirm(message string) (bool, error)

	// Info displays an informational message.
	Info(message string)

	// Warn displays a warning message.
	Warn(message string)

	// Error displays an error message.
	Error(message string)
}

// ProposalAction represents what to do with a proposal.
type ProposalAction int

const (
	// ProposalActionApply applies the proposal.
	ProposalActionApply ProposalAction = iota
	// ProposalActionSkip skips the proposal.
	ProposalActionSkip
	// ProposalActionEdit allows editing before applying.
	ProposalActionEdit
	// ProposalActionAbort aborts the entire operation.
	ProposalActionAbort
)

// String returns the string representation of the proposal action.
func (pa ProposalAction) String() string {
	switch pa {
	case ProposalActionApply:
		return "apply"
	case ProposalActionSkip:
		return "skip"
	case ProposalActionEdit:
		return "edit"
	case ProposalActionAbort:
		return "abort"
	default:
		return "unknown"
	}
}

// ReviewProposal presents a proposal and asks for a decision.
func ReviewProposal(p Prompter, proposal actions.Proposal) (ProposalAction, error) {
	if err := p.ShowProposal(proposal); err != nil {
		return ProposalActionAbort, err
	}

	q := Question{
		ID:   "proposal_action",
		Text: "What would you like to do?",
		Type: QuestionTypeSingleChoice,
		Options: []Option{
			{ID: "apply", Label: "Apply", Description: "Apply this change"},
			{ID: "skip", Label: "Skip", Description: "Skip this change"},
			{ID: "abort", Label: "Abort", Description: "Abort the entire operation"},
		},
		Default: "apply",
	}

	answer, err := p.Ask(q)
	if err != nil {
		return ProposalActionAbort, err
	}

	if len(answer.Selected) == 0 {
		return ProposalActionApply, nil
	}

	switch answer.Selected[0] {
	case "apply":
		return ProposalActionApply, nil
	case "skip":
		return ProposalActionSkip, nil
	case "abort":
		return ProposalActionAbort, nil
	default:
		return ProposalActionApply, nil
	}
}
