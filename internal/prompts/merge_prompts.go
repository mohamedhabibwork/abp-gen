package prompts

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

// MergeDecision represents the user's decision for handling existing files
type MergeDecision string

const (
	MergeDecisionOverwrite MergeDecision = "overwrite"
	MergeDecisionMerge     MergeDecision = "merge"
	MergeDecisionSkip      MergeDecision = "skip"
	MergeDecisionShowDiff  MergeDecision = "showdiff"
)

// ConflictResolution represents how a conflict should be resolved
type ConflictResolution int

const (
	ResolutionKeepExisting ConflictResolution = iota
	ResolutionUseNew
	ResolutionKeepBoth
	ResolutionSkip
	ResolutionShowContext
)

// ConflictType represents the type of merge conflict
type ConflictType int

const (
	ConflictTypeDuplicateClass ConflictType = iota
	ConflictTypeDuplicateMethod
	ConflictTypeDuplicateProperty
	ConflictTypeDifferentValue
	ConflictTypeStructural
)

// Conflict represents a merge conflict
type Conflict struct {
	Type         ConflictType
	Description  string
	ExistingCode string
	NewCode      string
	Line         int
}

// PromptMergeDecision prompts the user for a merge decision
func PromptMergeDecision(filePath string, fileTypeName string) (MergeDecision, error) {
	var decision string

	prompt := &survey.Select{
		Message: fmt.Sprintf("File exists: %s (%s). What would you like to do?", filePath, fileTypeName),
		Options: []string{
			"Merge intelligently (recommended)",
			"Overwrite with new content",
			"Skip this file",
			"Show diff first",
		},
		Default: "Merge intelligently (recommended)",
	}

	if err := survey.AskOne(prompt, &decision); err != nil {
		return "", err
	}

	switch decision {
	case "Merge intelligently (recommended)":
		return MergeDecisionMerge, nil
	case "Overwrite with new content":
		return MergeDecisionOverwrite, nil
	case "Skip this file":
		return MergeDecisionSkip, nil
	case "Show diff first":
		return MergeDecisionShowDiff, nil
	default:
		return MergeDecisionSkip, nil
	}
}

// PromptMergeAll prompts the user if they want to apply the same decision to all files
func PromptMergeAll() (bool, error) {
	var applyToAll bool

	prompt := &survey.Confirm{
		Message: "Apply this decision to all remaining files?",
		Default: false,
	}

	if err := survey.AskOne(prompt, &applyToAll); err != nil {
		return false, err
	}

	return applyToAll, nil
}

// PromptConflictResolution prompts the user for conflict resolution
func PromptConflictResolution(conflict Conflict, index int, total int) (ConflictResolution, error) {
	fmt.Printf("\n⚠️  Merge conflict %d of %d\n", index+1, total)
	fmt.Printf("Type: %s\n", getConflictTypeName(conflict.Type))
	fmt.Printf("Description: %s\n", conflict.Description)

	if conflict.Line > 0 {
		fmt.Printf("Line: %d\n", conflict.Line)
	}

	fmt.Println("\nExisting code:")
	fmt.Println("───────────────")
	printCode(conflict.ExistingCode)

	fmt.Println("\nNew code:")
	fmt.Println("─────────")
	printCode(conflict.NewCode)

	var resolution string

	prompt := &survey.Select{
		Message: "How would you like to resolve this conflict?",
		Options: []string{
			"Keep existing",
			"Use new",
			"Keep both (rename new)",
			"Skip this conflict",
		},
		Default: "Keep existing",
	}

	if err := survey.AskOne(prompt, &resolution); err != nil {
		return ResolutionKeepExisting, err
	}

	switch resolution {
	case "Keep existing":
		return ResolutionKeepExisting, nil
	case "Use new":
		return ResolutionUseNew, nil
	case "Keep both (rename new)":
		return ResolutionKeepBoth, nil
	case "Skip this conflict":
		return ResolutionSkip, nil
	default:
		return ResolutionKeepExisting, nil
	}
}

// PromptConflictBatch prompts for multiple conflicts at once
func PromptConflictBatch(conflicts []Conflict) (map[int]ConflictResolution, error) {
	resolutions := make(map[int]ConflictResolution)

	for i, conflict := range conflicts {
		resolution, err := PromptConflictResolution(conflict, i, len(conflicts))
		if err != nil {
			return nil, err
		}

		resolutions[i] = resolution

		// Ask if user wants to apply same resolution to all remaining conflicts
		if i < len(conflicts)-1 {
			var applyToAll bool
			prompt := &survey.Confirm{
				Message: "Apply this resolution to all remaining conflicts of this type?",
				Default: false,
			}

			if err := survey.AskOne(prompt, &applyToAll); err != nil {
				return nil, err
			}

			if applyToAll {
				// Apply same resolution to remaining conflicts of same type
				for j := i + 1; j < len(conflicts); j++ {
					if conflicts[j].Type == conflict.Type {
						resolutions[j] = resolution
					}
				}
				break
			}
		}
	}

	return resolutions, nil
}

// Helper functions

func getConflictTypeName(conflictType ConflictType) string {
	switch conflictType {
	case ConflictTypeDuplicateClass:
		return "Duplicate Class"
	case ConflictTypeDuplicateMethod:
		return "Duplicate Method"
	case ConflictTypeDuplicateProperty:
		return "Duplicate Property"
	case ConflictTypeDifferentValue:
		return "Different Value"
	case ConflictTypeStructural:
		return "Structural Conflict"
	default:
		return "Unknown"
	}
}

func printCode(code string) {
	lines := splitLines(code)
	for _, line := range lines {
		if len(line) > 100 {
			fmt.Println("  " + line[:97] + "...")
		} else {
			fmt.Println("  " + line)
		}
	}
}

func splitLines(text string) []string {
	var lines []string
	current := ""

	for _, ch := range text {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}
