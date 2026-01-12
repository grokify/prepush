package interactive

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/grokify/release-agent/pkg/actions"
)

// JSONPrompter implements Prompter with JSON input/output for Claude Code integration.
type JSONPrompter struct {
	writer  io.Writer
	reader  *bufio.Reader
	encoder *json.Encoder
}

// NewJSONPrompter creates a new JSONPrompter.
func NewJSONPrompter(w io.Writer, r io.Reader) *JSONPrompter {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return &JSONPrompter{
		writer:  w,
		reader:  bufio.NewReader(r),
		encoder: encoder,
	}
}

// DefaultJSONPrompter returns a JSONPrompter using stdout/stdin.
func DefaultJSONPrompter() *JSONPrompter {
	return NewJSONPrompter(os.Stdout, os.Stdin)
}

// jsonMessage is the base JSON protocol message.
type jsonMessage struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
}

// jsonQuestionMessage represents a question in JSON format.
type jsonQuestionMessage struct {
	jsonMessage
	Question   string       `json:"question"`
	InputType  string       `json:"input_type"`
	Options    []jsonOption `json:"options,omitempty"`
	Default    string       `json:"default,omitempty"`
	Context    string       `json:"context,omitempty"`
	Required   bool         `json:"required"`
	WaitingFor string       `json:"waiting_for"`
}

// jsonOption represents a choice option in JSON format.
type jsonOption struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// jsonAnswerMessage represents a user's answer in JSON format.
type jsonAnswerMessage struct {
	QuestionID string   `json:"question_id"`
	Selected   []string `json:"selected,omitempty"`
	Text       string   `json:"text,omitempty"`
	Confirmed  *bool    `json:"confirmed,omitempty"`
}

// jsonProposalMessage represents a proposed change in JSON format.
type jsonProposalMessage struct {
	jsonMessage
	Description string            `json:"description"`
	FilePath    string            `json:"file_path,omitempty"`
	OldContent  string            `json:"old_content,omitempty"`
	NewContent  string            `json:"new_content,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	WaitingFor  string            `json:"waiting_for"`
	Actions     []string          `json:"actions"`
}

// jsonInfoMessage represents an informational message.
type jsonInfoMessage struct {
	jsonMessage
	Text string `json:"text"`
}

// Ask presents a question and returns the user's answer via JSON.
func (p *JSONPrompter) Ask(q Question) (Answer, error) {
	// Convert options
	options := make([]jsonOption, len(q.Options))
	for i, opt := range q.Options {
		options[i] = jsonOption(opt)
	}

	// Write question as JSON
	msg := jsonQuestionMessage{
		jsonMessage: jsonMessage{
			Type: "question",
			ID:   q.ID,
		},
		Question:   q.Text,
		InputType:  q.Type.String(),
		Options:    options,
		Default:    q.Default,
		Context:    q.Context,
		Required:   true,
		WaitingFor: "user_input",
	}

	if err := p.encoder.Encode(msg); err != nil {
		return Answer{}, fmt.Errorf("failed to write question: %w", err)
	}

	// Read answer from stdin
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return Answer{}, fmt.Errorf("failed to read answer: %w", err)
	}

	var answerMsg jsonAnswerMessage
	if err := json.Unmarshal([]byte(line), &answerMsg); err != nil {
		return Answer{}, fmt.Errorf("failed to parse answer: %w", err)
	}

	answer := Answer{
		QuestionID: answerMsg.QuestionID,
		Selected:   answerMsg.Selected,
		Text:       answerMsg.Text,
	}
	if answerMsg.Confirmed != nil {
		answer.Confirmed = *answerMsg.Confirmed
	}

	return answer, nil
}

// ShowProposal displays a proposed change for review via JSON.
func (p *JSONPrompter) ShowProposal(proposal actions.Proposal) error {
	msg := jsonProposalMessage{
		jsonMessage: jsonMessage{
			Type: "proposal",
		},
		Description: proposal.Description,
		FilePath:    proposal.FilePath,
		OldContent:  proposal.OldContent,
		NewContent:  proposal.NewContent,
		Metadata:    proposal.Metadata,
		WaitingFor:  "user_approval",
		Actions:     []string{"apply", "skip", "abort"},
	}

	return p.encoder.Encode(msg)
}

// Confirm asks a yes/no question via JSON.
func (p *JSONPrompter) Confirm(message string) (bool, error) {
	q := Question{
		ID:   "confirm",
		Text: message,
		Type: QuestionTypeConfirm,
		Options: []Option{
			{ID: "yes", Label: "Yes"},
			{ID: "no", Label: "No"},
		},
		Default: "yes",
	}

	answer, err := p.Ask(q)
	if err != nil {
		return false, err
	}

	return answer.Confirmed || (len(answer.Selected) > 0 && answer.Selected[0] == "yes"), nil
}

// Info displays an informational message via JSON.
func (p *JSONPrompter) Info(message string) {
	msg := jsonInfoMessage{
		jsonMessage: jsonMessage{Type: "info"},
		Text:        message,
	}
	_ = p.encoder.Encode(msg)
}

// Warn displays a warning message via JSON.
func (p *JSONPrompter) Warn(message string) {
	msg := jsonInfoMessage{
		jsonMessage: jsonMessage{Type: "warning"},
		Text:        message,
	}
	_ = p.encoder.Encode(msg)
}

// Error displays an error message via JSON.
func (p *JSONPrompter) Error(message string) {
	msg := jsonInfoMessage{
		jsonMessage: jsonMessage{Type: "error"},
		Text:        message,
	}
	_ = p.encoder.Encode(msg)
}
