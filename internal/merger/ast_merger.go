package merger

import (
	"fmt"
	"strings"
)

// ASTMerger handles AST-based file merging for complex C# files
type ASTMerger struct {
	parser *CSharpParser
}

// NewASTMerger creates a new AST-based merger
func NewASTMerger() *ASTMerger {
	return &ASTMerger{
		parser: NewCSharpParser(),
	}
}

// Merge performs AST-based merging
func (m *ASTMerger) Merge(existing string, newContent string, fileType FileType) (string, []Conflict, error) {
	// Parse both files
	existingClass, err := m.parser.ParseClass(existing)
	if err != nil || existingClass == nil {
		return "", nil, fmt.Errorf("failed to parse existing class: %w", err)
	}

	newClass, err := m.parser.ParseClass(newContent)
	if err != nil || newClass == nil {
		return "", nil, fmt.Errorf("failed to parse new class: %w", err)
	}

	// Detect conflicts
	conflicts := m.detectConflicts(existingClass, newClass)

	// If there are conflicts, return them for user resolution
	if len(conflicts) > 0 {
		return existing, conflicts, nil
	}

	// Merge classes
	merged, err := m.mergeClasses(existing, existingClass, newClass, fileType)
	if err != nil {
		return "", nil, err
	}

	return merged, nil, nil
}

// detectConflicts detects conflicts between two classes
func (m *ASTMerger) detectConflicts(existing *CSharpClass, new *CSharpClass) []Conflict {
	var conflicts []Conflict

	// Check for duplicate properties with different types
	for _, newProp := range new.Properties {
		existingProp := m.parser.FindProperty(existing, newProp.Name)
		if existingProp != nil && existingProp.Type != newProp.Type {
			conflicts = append(conflicts, Conflict{
				Type:         ConflictTypeDuplicateProperty,
				Description:  fmt.Sprintf("Property '%s' exists with different type", newProp.Name),
				ExistingCode: existingProp.RawContent,
				NewCode:      newProp.RawContent,
				Line:         existingProp.Line,
			})
		}
	}

	// Check for duplicate methods with same signature but different implementation
	for _, newMethod := range new.Methods {
		existingMethod := m.parser.FindMethod(existing, newMethod.Signature)
		if existingMethod != nil {
			// Method exists - this is a conflict if implementations differ
			if strings.TrimSpace(existingMethod.Body) != strings.TrimSpace(newMethod.Body) {
				conflicts = append(conflicts, Conflict{
					Type:         ConflictTypeDuplicateMethod,
					Description:  fmt.Sprintf("Method '%s' exists with different implementation", newMethod.Name),
					ExistingCode: existingMethod.RawContent,
					NewCode:      newMethod.RawContent,
					Line:         existingMethod.StartLine,
				})
			}
		}
	}

	return conflicts
}

// mergeClasses merges two C# classes
func (m *ASTMerger) mergeClasses(original string, existingClass *CSharpClass, newClass *CSharpClass, fileType FileType) (string, error) {
	result := original

	// Merge usings
	existingUsings := m.parser.ExtractUsings(original)
	newUsings := m.parser.ExtractUsings(newClass.RawContent)
	mergedUsings := m.parser.MergeUsings(existingUsings, newUsings)

	// Update usings in result
	result = m.replaceUsings(result, mergedUsings)

	// Merge properties
	result = m.mergeProperties(result, existingClass, newClass)

	// Merge methods
	result = m.mergeMethods(result, existingClass, newClass, fileType)

	return result, nil
}

// mergeProperties merges properties from new class into existing
func (m *ASTMerger) mergeProperties(content string, existing *CSharpClass, new *CSharpClass) string {
	var toAdd []string

	// Find properties that don't exist in existing class
	for _, newProp := range new.Properties {
		if !m.parser.HasProperty(existing, newProp.Name) {
			toAdd = append(toAdd, "    "+strings.TrimSpace(newProp.RawContent))
		}
	}

	if len(toAdd) == 0 {
		return content
	}

	// Find insertion point (after last property or before first method)
	lines := strings.Split(content, "\n")
	insertIndex := -1

	// Find last property
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "{ get; set; }") {
			insertIndex = i + 1
			break
		}
	}

	// If no properties found, insert after class opening brace
	if insertIndex == -1 {
		for i, line := range lines {
			if strings.Contains(line, "class "+existing.Name) {
				// Find opening brace
				for j := i; j < len(lines); j++ {
					if strings.Contains(lines[j], "{") {
						insertIndex = j + 1
						break
					}
				}
				break
			}
		}
	}

	if insertIndex == -1 {
		return content
	}

	// Insert properties
	newLines := make([]string, 0, len(lines)+len(toAdd))
	newLines = append(newLines, lines[:insertIndex]...)
	newLines = append(newLines, toAdd...)
	newLines = append(newLines, lines[insertIndex:]...)

	return strings.Join(newLines, "\n")
}

// mergeMethods merges methods from new class into existing
func (m *ASTMerger) mergeMethods(content string, existing *CSharpClass, new *CSharpClass, fileType FileType) string {
	var toAdd []string

	// Find methods that don't exist in existing class
	for _, newMethod := range new.Methods {
		// Skip constructors for entities (they're usually specific)
		if newMethod.Name == existing.Name && fileType == FileTypeEntity {
			continue
		}

		if !m.parser.HasMethod(existing, newMethod.Signature) {
			toAdd = append(toAdd, "\n"+strings.TrimSpace(newMethod.RawContent))
		}
	}

	if len(toAdd) == 0 {
		return content
	}

	// Find insertion point (before closing class brace)
	lines := strings.Split(content, "\n")
	insertIndex := -1

	// Find closing brace of class
	braceCount := 0
	inClass := false

	for i, line := range lines {
		if strings.Contains(line, "class "+existing.Name) {
			inClass = true
		}
		if inClass {
			braceCount += strings.Count(line, "{")
			braceCount -= strings.Count(line, "}")
			if braceCount == 0 && strings.TrimSpace(line) == "}" {
				insertIndex = i
				break
			}
		}
	}

	if insertIndex == -1 {
		return content
	}

	// Insert methods
	newLines := make([]string, 0, len(lines)+len(toAdd))
	newLines = append(newLines, lines[:insertIndex]...)
	newLines = append(newLines, toAdd...)
	newLines = append(newLines, lines[insertIndex:]...)

	return strings.Join(newLines, "\n")
}

// replaceUsings replaces using statements in the content
func (m *ASTMerger) replaceUsings(content string, usings []string) string {
	// Find where usings end
	lines := strings.Split(content, "\n")
	lastUsingIndex := -1

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "using ") {
			lastUsingIndex = i
		}
	}

	if lastUsingIndex == -1 {
		return content
	}

	// Build new using statements
	var usingLines []string
	for _, u := range usings {
		usingLines = append(usingLines, "using "+u+";")
	}

	// Replace existing usings
	newLines := make([]string, 0)

	// Add usings
	newLines = append(newLines, usingLines...)

	// Add rest of content (skip old usings)
	for i := lastUsingIndex + 1; i < len(lines); i++ {
		newLines = append(newLines, lines[i])
	}

	return strings.Join(newLines, "\n")
}
