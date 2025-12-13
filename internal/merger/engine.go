package merger

import (
	"fmt"
	"os"

	"github.com/mohamedhabibwork/abp-gen/internal/prompts"
)

// MergeDecision type alias
type MergeDecision = prompts.MergeDecision

const (
	MergeDecisionOverwrite = prompts.MergeDecisionOverwrite
	MergeDecisionMerge     = prompts.MergeDecisionMerge
	MergeDecisionSkip      = prompts.MergeDecisionSkip
	MergeDecisionShowDiff  = prompts.MergeDecisionShowDiff
)

// Engine orchestrates merge operations
type Engine struct {
	detector         *Detector
	classifier       *Classifier
	patternMerger    *PatternMerger
	astMerger        *ASTMerger
	jsonMerger       *JSONMerger
	conflictResolver *ConflictResolver
	
	// Configuration
	Force      bool
	MergeAll   bool
	MergeMode  MergeDecision
	Verbose    bool
}

// NewEngine creates a new merge engine
func NewEngine(force bool, verbose bool) *Engine {
	return &Engine{
		detector:         NewDetector(),
		classifier:       NewClassifier(),
		patternMerger:    NewPatternMerger(),
		astMerger:        NewASTMerger(),
		jsonMerger:       NewJSONMerger(),
		conflictResolver: NewConflictResolver(),
		Force:            force,
		Verbose:          verbose,
	}
}

// MergeFile merges a new file with an existing file if it exists
func (e *Engine) MergeFile(path string, newContent string) (string, bool, error) {
	// Check if file exists
	fileExists, err := e.detector.CheckFile(path)
	if err != nil {
		return "", false, err
	}
	
	// If file doesn't exist, return new content
	if !fileExists.Exists {
		return newContent, true, nil
	}
	
	// If force mode, overwrite
	if e.Force {
		if e.Verbose {
			fmt.Printf("[OVERWRITE] %s\n", path)
		}
		return newContent, true, nil
	}
	
	// Check if file can be merged
	if !e.detector.CanMerge(fileExists.FileType) {
		// File type doesn't support merging
		if e.Verbose {
			fmt.Printf("[SKIP] %s (file type doesn't support merging)\n", path)
		}
		return "", false, nil
	}
	
	// Prompt user for merge decision if not in merge-all mode
	var decision MergeDecision
	if e.MergeAll && e.MergeMode != "" {
		decision = e.MergeMode
	} else {
		fileTypeName := e.classifier.GetFileTypeName(fileExists.FileType)
		decision, err = prompts.PromptMergeDecision(path, fileTypeName)
		if err != nil {
			return "", false, err
		}
		
		// Ask if user wants to apply to all
		if !e.MergeAll {
			applyToAll, err := prompts.PromptMergeAll()
			if err != nil {
				return "", false, err
			}
			if applyToAll {
				e.MergeAll = true
				e.MergeMode = decision
			}
		}
	}
	
	// Handle user decision
	switch decision {
	case MergeDecisionOverwrite:
		if e.Verbose {
			fmt.Printf("[OVERWRITE] %s\n", path)
		}
		return newContent, true, nil
		
	case MergeDecisionSkip:
		if e.Verbose {
			fmt.Printf("[SKIP] %s\n", path)
		}
		return "", false, nil
		
	case MergeDecisionShowDiff:
		// TODO: Show diff
		fmt.Println("Diff display not yet implemented")
		return "", false, nil
		
	case MergeDecisionMerge:
		return e.performMerge(path, fileExists, newContent)
		
	default:
		return "", false, nil
	}
}

// performMerge performs the actual merge operation
func (e *Engine) performMerge(path string, fileExists *FileExistence, newContent string) (string, bool, error) {
	// Read existing content
	existingContent, err := os.ReadFile(path)
	if err != nil {
		return "", false, fmt.Errorf("failed to read existing file: %w", err)
	}
	
	existing := string(existingContent)
	
	// Select merge strategy
	strategy := e.classifier.GetMergeStrategy(fileExists.FileType)
	
	var merged string
	var conflicts []Conflict
	
	// Perform merge based on strategy
	switch strategy {
	case MergeStrategyPattern:
		merged, conflicts, err = e.patternMerger.Merge(existing, newContent, fileExists.FileType)
		
	case MergeStrategyAST:
		merged, conflicts, err = e.astMerger.Merge(existing, newContent, fileExists.FileType)
		
	case MergeStrategyJSON:
		merged, conflicts, err = e.jsonMerger.Merge(existing, newContent)
		
	default:
		return "", false, fmt.Errorf("unsupported merge strategy for file type %v", fileExists.FileType)
	}
	
	if err != nil {
		return "", false, fmt.Errorf("merge failed: %w", err)
	}
	
	// Handle conflicts if any
	if len(conflicts) > 0 {
		if e.Verbose {
			fmt.Printf("[CONFLICTS] %s - %d conflict(s) detected\n", path, len(conflicts))
		}
		
		// Prompt user to resolve conflicts
		resolutions, err := prompts.PromptConflictBatch(conflicts)
		if err != nil {
			return "", false, fmt.Errorf("failed to resolve conflicts: %w", err)
		}
		
		// Apply resolutions
		merged, err = e.conflictResolver.ResolveConflicts(existing, conflicts, resolutions, newContent)
		if err != nil {
			return "", false, fmt.Errorf("failed to apply conflict resolutions: %w", err)
		}
	}
	
	if e.Verbose {
		fmt.Printf("[MERGED] %s\n", path)
	}
	
	return merged, true, nil
}

// SetMergeAll sets merge-all mode with a specific decision
func (e *Engine) SetMergeAll(decision MergeDecision) {
	e.MergeAll = true
	e.MergeMode = decision
}

// ResetMergeAll resets merge-all mode
func (e *Engine) ResetMergeAll() {
	e.MergeAll = false
	e.MergeMode = ""
}

