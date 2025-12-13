package writer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mohamedhabibwork/abp-gen/internal/merger"
)

// FileOperation represents a file operation to be performed
type FileOperation struct {
	Type     OperationType
	Path     string
	Content  string
	Existing bool // Whether file already exists
}

// OperationType represents the type of file operation
type OperationType string

const (
	OperationCreate OperationType = "CREATE"
	OperationUpdate OperationType = "UPDATE"
	OperationSkip   OperationType = "SKIP"
)

// Writer handles file writing with support for dry-run and force modes
type Writer struct {
	DryRun      bool
	Force       bool
	Verbose     bool
	MergeMode   bool
	Operations  []FileOperation
	mergeEngine *merger.Engine
}

// NewWriter creates a new file writer
func NewWriter(dryRun, force, verbose bool) *Writer {
	return &Writer{
		DryRun:      dryRun,
		Force:       force,
		Verbose:     verbose,
		MergeMode:   false,
		Operations:  []FileOperation{},
		mergeEngine: merger.NewEngine(force, verbose),
	}
}

// NewWriterWithMerge creates a new file writer with merge support
func NewWriterWithMerge(dryRun, force, verbose, mergeMode bool) *Writer {
	return &Writer{
		DryRun:      dryRun,
		Force:       force,
		Verbose:     verbose,
		MergeMode:   mergeMode,
		Operations:  []FileOperation{},
		mergeEngine: merger.NewEngine(force, verbose),
	}
}

// WriteFile writes content to a file
func (w *Writer) WriteFile(path string, content string) error {
	// Normalize path
	path = filepath.Clean(path)

	// Check if file exists
	exists := fileExists(path)

	// If merge mode is enabled and file exists, try to merge
	if w.MergeMode && exists && !w.Force {
		mergedContent, shouldWrite, err := w.mergeEngine.MergeFile(path, content)
		if err != nil {
			return fmt.Errorf("merge failed for %s: %w", path, err)
		}
		
		if !shouldWrite {
			// User chose to skip
			w.Operations = append(w.Operations, FileOperation{
				Type:     OperationSkip,
				Path:     path,
				Content:  content,
				Existing: true,
			})
			w.logOperation(OperationSkip, path)
			return nil
		}
		
		// Use merged content
		content = mergedContent
	}

	// Determine operation type
	opType := OperationCreate
	if exists {
		if w.Force || w.MergeMode {
			opType = OperationUpdate
		} else {
			opType = OperationSkip
			w.logOperation(opType, path)
			return nil
		}
	}

	// Record operation
	w.Operations = append(w.Operations, FileOperation{
		Type:     opType,
		Path:     path,
		Content:  content,
		Existing: exists,
	})

	// Log operation
	w.logOperation(opType, path)

	// If dry-run, don't actually write
	if w.DryRun {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// UpdateFile updates an existing file by applying a modification function
func (w *Writer) UpdateFile(path string, modifyFunc func(string) (string, error)) error {
	// Read existing content
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return err
	}

	// Apply modification
	newContent, err := modifyFunc(string(content))
	if err != nil {
		return err
	}

	// Write back
	return w.WriteFile(path, newContent)
}

// UpdateFileIdempotent updates a file only if the specified content doesn't already exist
func (w *Writer) UpdateFileIdempotent(path string, searchPattern string, insertFunc func(string) (string, error)) error {
	// Read existing content
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return err
	}

	contentStr := string(content)

	// Check if pattern already exists
	if strings.Contains(contentStr, searchPattern) {
		w.logOperation(OperationSkip, path+" (already contains pattern)")
		return nil
	}

	// Apply modification
	newContent, err := insertFunc(contentStr)
	if err != nil {
		return err
	}

	// Write back
	return w.WriteFile(path, newContent)
}

// EnsureDirectory ensures a directory exists
func (w *Writer) EnsureDirectory(path string) error {
	if w.DryRun {
		w.logOperation("CREATE_DIR", path)
		return nil
	}

	return os.MkdirAll(path, 0755)
}

// PrintSummary prints a summary of operations
func (w *Writer) PrintSummary() {
	if len(w.Operations) == 0 {
		fmt.Println("No operations performed.")
		return
	}

	created := 0
	updated := 0
	skipped := 0

	for _, op := range w.Operations {
		switch op.Type {
		case OperationCreate:
			created++
		case OperationUpdate:
			updated++
		case OperationSkip:
			skipped++
		}
	}

	fmt.Println("\n=== Summary ===")
	fmt.Printf("Created: %d\n", created)
	fmt.Printf("Updated: %d\n", updated)
	fmt.Printf("Skipped: %d\n", skipped)
	fmt.Printf("Total:   %d\n", len(w.Operations))

	if w.DryRun {
		fmt.Println("\nDRY RUN: No files were actually modified.")
	}
}

// logOperation logs a file operation
func (w *Writer) logOperation(opType interface{}, path string) {
	if !w.Verbose && !w.DryRun {
		return
	}

	var prefix string
	switch opType {
	case OperationCreate:
		prefix = "[CREATE]"
	case OperationUpdate:
		prefix = "[UPDATE]"
	case OperationSkip:
		prefix = "[SKIP]  "
	case "CREATE_DIR":
		prefix = "[MKDIR] "
	default:
		prefix = "[INFO]  "
	}

	fmt.Printf("%s %s\n", prefix, path)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// ReadFile reads a file and returns its content
func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

