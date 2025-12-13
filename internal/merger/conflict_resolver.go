package merger

import (
	"fmt"
	"strings"
)

// ConflictResolver handles interactive conflict resolution
type ConflictResolver struct {
	parser *CSharpParser
}

// NewConflictResolver creates a new conflict resolver
func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{
		parser: NewCSharpParser(),
	}
}

// ResolveConflicts resolves conflicts based on resolution strategy
func (r *ConflictResolver) ResolveConflicts(existing string, conflicts []Conflict, resolutions map[int]ConflictResolution, newContent string) (string, error) {
	if len(conflicts) == 0 {
		return existing, nil
	}

	result := existing

	for i, conflict := range conflicts {
		resolution, ok := resolutions[i]
		if !ok {
			// Default to keeping existing
			resolution = ResolutionKeepExisting
		}

		switch resolution {
		case ResolutionKeepExisting:
			// Do nothing, keep existing
			continue

		case ResolutionUseNew:
			// Replace existing with new
			result = r.replaceCode(result, conflict.ExistingCode, conflict.NewCode)

		case ResolutionKeepBoth:
			// Keep both, rename new one
			newCodeRenamed := r.renameConflictingItem(conflict.NewCode, conflict.Type)
			result = r.insertAfter(result, conflict.ExistingCode, newCodeRenamed)

		case ResolutionSkip:
			// Skip this conflict
			continue
		}
	}

	return result, nil
}

// replaceCode replaces old code with new code
func (r *ConflictResolver) replaceCode(content string, oldCode string, newCode string) string {
	return strings.Replace(content, oldCode, newCode, 1)
}

// insertAfter inserts new code after existing code
func (r *ConflictResolver) insertAfter(content string, existingCode string, newCode string) string {
	index := strings.Index(content, existingCode)
	if index == -1 {
		return content
	}

	endIndex := index + len(existingCode)
	return content[:endIndex] + "\n\n" + newCode + content[endIndex:]
}

// renameConflictingItem renames a conflicting item (property, method, class)
func (r *ConflictResolver) renameConflictingItem(code string, conflictType ConflictType) string {
	switch conflictType {
	case ConflictTypeDuplicateProperty:
		return r.renameProperty(code)
	case ConflictTypeDuplicateMethod:
		return r.renameMethod(code)
	case ConflictTypeDuplicateClass:
		return r.renameClass(code)
	default:
		return code + "_New"
	}
}

// renameProperty renames a property by appending a suffix
func (r *ConflictResolver) renameProperty(code string) string {
	// Extract property name
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if strings.Contains(line, "{ get; set; }") {
			// Find property name
			parts := strings.Fields(line)
			for j, part := range parts {
				if j > 0 && parts[j-1] != "public" && !strings.HasPrefix(part, "[") {
					// This is likely the property name
					newName := part + "2"
					lines[i] = strings.Replace(line, part, newName, 1)
					break
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}

// renameMethod renames a method by appending a suffix
func (r *ConflictResolver) renameMethod(code string) string {
	// Find method name in signature
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if strings.Contains(line, "(") && (strings.Contains(line, "public") || strings.Contains(line, "private") || strings.Contains(line, "protected")) {
			// Extract method name (word before opening parenthesis)
			parenIndex := strings.Index(line, "(")
			if parenIndex > 0 {
				beforeParen := line[:parenIndex]
				parts := strings.Fields(beforeParen)
				if len(parts) > 0 {
					methodName := parts[len(parts)-1]
					newName := methodName + "2"
					lines[i] = strings.Replace(line, methodName+"(", newName+"(", 1)
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}

// renameClass renames a class by appending a suffix
func (r *ConflictResolver) renameClass(code string) string {
	// Find class name in declaration
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if strings.Contains(line, "class ") {
			// Extract class name
			classIndex := strings.Index(line, "class ")
			if classIndex >= 0 {
				afterClass := line[classIndex+6:]
				parts := strings.Fields(afterClass)
				if len(parts) > 0 {
					className := parts[0]
					// Remove any inheritance syntax
					if colonIndex := strings.Index(className, ":"); colonIndex > 0 {
						className = className[:colonIndex]
					}
					if braceIndex := strings.Index(className, "{"); braceIndex > 0 {
						className = className[:braceIndex]
					}
					className = strings.TrimSpace(className)
					newName := className + "2"
					lines[i] = strings.Replace(line, "class "+className, "class "+newName, 1)
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}

// FormatConflict formats a conflict for display to the user
func (r *ConflictResolver) FormatConflict(conflict Conflict, index int) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("\n⚠️  Conflict %d: %s\n", index+1, conflict.Description))

	if conflict.Line > 0 {
		builder.WriteString(fmt.Sprintf("Line: %d\n", conflict.Line))
	}

	builder.WriteString("\nExisting code:\n")
	builder.WriteString("  " + strings.ReplaceAll(conflict.ExistingCode, "\n", "\n  "))
	builder.WriteString("\n\nNew code:\n")
	builder.WriteString("  " + strings.ReplaceAll(conflict.NewCode, "\n", "\n  "))
	builder.WriteString("\n")

	return builder.String()
}

// GetConflictTypeName returns a human-readable name for a conflict type
func (r *ConflictResolver) GetConflictTypeName(conflictType ConflictType) string {
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
