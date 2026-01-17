package interactive

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/agentplexus/release-agent-team/pkg/actions"
)

// CLIPrompter implements Prompter for terminal interaction.
type CLIPrompter struct {
	reader *bufio.Reader
}

// NewCLIPrompter creates a new CLIPrompter.
func NewCLIPrompter() *CLIPrompter {
	return &CLIPrompter{
		reader: bufio.NewReader(os.Stdin),
	}
}

// Ask presents a question and returns the user's answer.
func (p *CLIPrompter) Ask(q Question) (Answer, error) {
	answer := Answer{QuestionID: q.ID}

	switch q.Type {
	case QuestionTypeSingleChoice:
		return p.askSingleChoice(q)
	case QuestionTypeMultiChoice:
		return p.askMultiChoice(q)
	case QuestionTypeConfirm:
		confirmed, err := p.Confirm(q.Text)
		if err != nil {
			return answer, err
		}
		answer.Confirmed = confirmed
		return answer, nil
	case QuestionTypeText:
		return p.askText(q)
	default:
		return answer, fmt.Errorf("unknown question type: %v", q.Type)
	}
}

func (p *CLIPrompter) askSingleChoice(q Question) (Answer, error) {
	answer := Answer{QuestionID: q.ID}

	// Display the question
	fmt.Println()
	fmt.Println(q.Text)

	if q.Context != "" {
		fmt.Println()
		fmt.Println(q.Context)
	}

	fmt.Println()

	// Display options
	for i, opt := range q.Options {
		marker := " "
		if opt.ID == q.Default {
			marker = "*"
		}
		if opt.Description != "" {
			fmt.Printf("%s %d) %s - %s\n", marker, i+1, opt.Label, opt.Description)
		} else {
			fmt.Printf("%s %d) %s\n", marker, i+1, opt.Label)
		}
	}

	fmt.Println()

	// Get input
	defaultNum := 0
	for i, opt := range q.Options {
		if opt.ID == q.Default {
			defaultNum = i + 1
			break
		}
	}

	prompt := "Enter choice"
	if defaultNum > 0 {
		prompt = fmt.Sprintf("Enter choice [%d]", defaultNum)
	}
	fmt.Printf("%s: ", prompt)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return answer, err
	}

	input = strings.TrimSpace(input)

	// Handle default
	if input == "" && defaultNum > 0 {
		answer.Selected = []string{q.Options[defaultNum-1].ID}
		return answer, nil
	}

	// Parse number
	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(q.Options) {
		return answer, fmt.Errorf("invalid choice: %s", input)
	}

	answer.Selected = []string{q.Options[num-1].ID}
	return answer, nil
}

func (p *CLIPrompter) askMultiChoice(q Question) (Answer, error) {
	answer := Answer{QuestionID: q.ID}

	// Display the question
	fmt.Println()
	fmt.Println(q.Text)
	fmt.Println("(Enter comma-separated numbers, e.g., 1,3,4)")

	if q.Context != "" {
		fmt.Println()
		fmt.Println(q.Context)
	}

	fmt.Println()

	// Display options
	for i, opt := range q.Options {
		if opt.Description != "" {
			fmt.Printf("  %d) %s - %s\n", i+1, opt.Label, opt.Description)
		} else {
			fmt.Printf("  %d) %s\n", i+1, opt.Label)
		}
	}

	fmt.Println()
	fmt.Print("Enter choices: ")

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return answer, err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return answer, nil
	}

	// Parse numbers
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		num, err := strconv.Atoi(part)
		if err != nil || num < 1 || num > len(q.Options) {
			return answer, fmt.Errorf("invalid choice: %s", part)
		}
		answer.Selected = append(answer.Selected, q.Options[num-1].ID)
	}

	return answer, nil
}

func (p *CLIPrompter) askText(q Question) (Answer, error) {
	answer := Answer{QuestionID: q.ID}

	// Display the question
	fmt.Println()
	fmt.Println(q.Text)

	if q.Context != "" {
		fmt.Println()
		fmt.Println(q.Context)
	}

	fmt.Println()

	prompt := "Enter response"
	if q.Default != "" {
		prompt = fmt.Sprintf("Enter response [%s]", q.Default)
	}
	fmt.Printf("%s: ", prompt)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return answer, err
	}

	input = strings.TrimSpace(input)

	// Handle default
	if input == "" && q.Default != "" {
		answer.Text = q.Default
	} else {
		answer.Text = input
	}

	return answer, nil
}

// ShowProposal displays a proposed change for review.
func (p *CLIPrompter) ShowProposal(proposal actions.Proposal) error {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📝 Proposed Change: %s\n", proposal.Description)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if proposal.FilePath != "" {
		fmt.Printf("\nFile: %s\n", proposal.FilePath)
	}

	// Show diff if we have old and new content
	if proposal.OldContent != "" || proposal.NewContent != "" {
		fmt.Println("\nChanges:")
		fmt.Println("─────────")

		if proposal.OldContent != "" && proposal.NewContent != "" {
			// Simple diff display
			oldLines := strings.Split(proposal.OldContent, "\n")
			newLines := strings.Split(proposal.NewContent, "\n")

			// Show abbreviated content
			if len(oldLines) > 10 {
				fmt.Printf("- [%d lines removed/changed]\n", len(oldLines))
			} else {
				for _, line := range oldLines {
					if line != "" {
						fmt.Printf("- %s\n", truncate(line, 70))
					}
				}
			}

			if len(newLines) > 10 {
				fmt.Printf("+ [%d lines added/changed]\n", len(newLines))
			} else {
				for _, line := range newLines {
					if line != "" {
						fmt.Printf("+ %s\n", truncate(line, 70))
					}
				}
			}
		} else if proposal.NewContent != "" {
			fmt.Println(truncate(proposal.NewContent, 500))
		}
	}

	// Show metadata
	if len(proposal.Metadata) > 0 {
		fmt.Println("\nDetails:")
		for k, v := range proposal.Metadata {
			fmt.Printf("  %s: %s\n", k, truncate(v, 60))
		}
	}

	fmt.Println()
	return nil
}

// Confirm asks a yes/no question.
func (p *CLIPrompter) Confirm(message string) (bool, error) {
	fmt.Printf("\n%s [y/N]: ", message)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes", nil
}

// Info displays an informational message.
func (p *CLIPrompter) Info(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// Warn displays a warning message.
func (p *CLIPrompter) Warn(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// Error displays an error message.
func (p *CLIPrompter) Error(message string) {
	fmt.Fprintf(os.Stderr, "❌ %s\n", message)
}

// truncate truncates a string to the specified length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
