package merger

import (
	"fmt"
	"regexp"
	"strings"
)

// PatternMerger handles pattern-based file merging
type PatternMerger struct {
	classifier *Classifier
}

// NewPatternMerger creates a new pattern-based merger
func NewPatternMerger() *PatternMerger {
	return &PatternMerger{
		classifier: NewClassifier(),
	}
}

// Merge performs pattern-based merging
func (m *PatternMerger) Merge(existing string, newContent string, fileType FileType) (string, []Conflict, error) {
	switch fileType {
	case FileTypePermissions:
		return m.mergePermissions(existing, newContent)
	case FileTypePermissionProvider:
		return m.mergePermissionProvider(existing, newContent)
	case FileTypeDbContext, FileTypeIDbContext:
		return m.mergeDbContext(existing, newContent)
	default:
		return "", nil, fmt.Errorf("unsupported file type for pattern merging: %v", fileType)
	}
}

// mergePermissions merges permission constant classes
func (m *PatternMerger) mergePermissions(existing string, newContent string) (string, []Conflict, error) {
	// Extract new permission classes
	newClasses := m.extractPermissionClasses(newContent)

	// Check for existing classes
	var conflicts []Conflict
	var toAdd []string

	for _, newClass := range newClasses {
		className := m.extractClassName(newClass)
		if className == "" {
			continue
		}

		// Check if class already exists
		if m.containsClass(existing, className) {
			conflicts = append(conflicts, Conflict{
				Type:         ConflictTypeDuplicateClass,
				Description:  fmt.Sprintf("Class '%s' already exists", className),
				ExistingCode: m.extractExistingClass(existing, className),
				NewCode:      newClass,
				Line:         m.findClassLine(existing, className),
			})
		} else {
			toAdd = append(toAdd, newClass)
		}
	}

	// If no conflicts, append new classes before closing namespace
	if len(conflicts) == 0 && len(toAdd) > 0 {
		merged := m.insertBeforeClosingBrace(existing, strings.Join(toAdd, "\n\n"))
		return merged, nil, nil
	}

	return existing, conflicts, nil
}

// mergePermissionProvider merges permission provider definitions
func (m *PatternMerger) mergePermissionProvider(existing string, newContent string) (string, []Conflict, error) {
	// Extract the Define method calls from new content
	definePattern := regexp.MustCompile(`context\.AddPermission\([^)]+\);`)
	newDefines := definePattern.FindAllString(newContent, -1)

	var conflicts []Conflict
	var toAdd []string

	for _, define := range newDefines {
		// Check if this permission definition already exists
		if strings.Contains(existing, define) {
			conflicts = append(conflicts, Conflict{
				Type:         ConflictTypeDuplicateMethod,
				Description:  "Permission definition already exists",
				ExistingCode: define,
				NewCode:      define,
			})
		} else {
			toAdd = append(toAdd, "            "+define)
		}
	}

	// Insert new definitions at the end of Define method
	if len(conflicts) == 0 && len(toAdd) > 0 {
		merged := m.insertIntoDefineMethod(existing, strings.Join(toAdd, "\n"))
		return merged, nil, nil
	}

	return existing, conflicts, nil
}

// mergeDbContext merges DbContext or IDbContext files
func (m *PatternMerger) mergeDbContext(existing string, newContent string) (string, []Conflict, error) {
	// Extract DbSet properties from new content
	dbSetPattern := regexp.MustCompile(`public\s+DbSet<(\w+)>\s+(\w+)\s*\{\s*get;\s*set;\s*\}`)
	newDbSets := dbSetPattern.FindAllString(newContent, -1)

	var conflicts []Conflict
	var toAdd []string

	for _, dbSet := range newDbSets {
		// Extract entity name
		matches := dbSetPattern.FindStringSubmatch(dbSet)
		if len(matches) < 3 {
			continue
		}
		propertyName := matches[2]

		// Check if DbSet already exists
		if strings.Contains(existing, fmt.Sprintf("%s {", propertyName)) {
			conflicts = append(conflicts, Conflict{
				Type:         ConflictTypeDuplicateProperty,
				Description:  fmt.Sprintf("DbSet property '%s' already exists", propertyName),
				ExistingCode: m.extractExistingDbSet(existing, propertyName),
				NewCode:      dbSet,
			})
		} else {
			toAdd = append(toAdd, "    "+dbSet)
		}
	}

	// Extract OnModelCreating configuration lines
	onModelCreatingPattern := regexp.MustCompile(`builder\.Entity<(\w+)>\([^)]+\);`)
	newConfigurations := onModelCreatingPattern.FindAllString(newContent, -1)

	for _, config := range newConfigurations {
		if !strings.Contains(existing, config) {
			toAdd = append(toAdd, "            "+config)
		}
	}

	// Insert new DbSets before closing class brace
	if len(conflicts) == 0 && len(toAdd) > 0 {
		merged := m.insertDbSets(existing, toAdd)
		return merged, nil, nil
	}

	return existing, conflicts, nil
}

// Helper methods

func (m *PatternMerger) extractPermissionClasses(content string) []string {
	// Extract public static classes
	classPattern := regexp.MustCompile(`(?s)public\s+static\s+class\s+\w+\s*\{[^}]*\}`)
	return classPattern.FindAllString(content, -1)
}

func (m *PatternMerger) extractClassName(classCode string) string {
	pattern := regexp.MustCompile(`public\s+static\s+class\s+(\w+)`)
	matches := pattern.FindStringSubmatch(classCode)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (m *PatternMerger) containsClass(content string, className string) bool {
	pattern := regexp.MustCompile(fmt.Sprintf(`public\s+static\s+class\s+%s\s*\{`, regexp.QuoteMeta(className)))
	return pattern.MatchString(content)
}

func (m *PatternMerger) extractExistingClass(content string, className string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`(?s)public\s+static\s+class\s+%s\s*\{[^}]*\}`, regexp.QuoteMeta(className)))
	matches := pattern.FindString(content)
	return matches
}

func (m *PatternMerger) findClassLine(content string, className string) int {
	lines := strings.Split(content, "\n")
	pattern := regexp.MustCompile(fmt.Sprintf(`public\s+static\s+class\s+%s`, regexp.QuoteMeta(className)))
	for i, line := range lines {
		if pattern.MatchString(line) {
			return i + 1
		}
	}
	return 0
}

func (m *PatternMerger) extractExistingDbSet(content string, propertyName string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`public\s+DbSet<\w+>\s+%s\s*\{\s*get;\s*set;\s*\}`, regexp.QuoteMeta(propertyName)))
	return pattern.FindString(content)
}

func (m *PatternMerger) insertBeforeClosingBrace(content string, toInsert string) string {
	// Find the last closing brace of namespace
	lines := strings.Split(content, "\n")
	insertIndex := -1

	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == "}" {
			insertIndex = i
			break
		}
	}

	if insertIndex == -1 {
		return content + "\n" + toInsert
	}

	// Insert before the closing brace
	result := strings.Join(lines[:insertIndex], "\n") + "\n\n" + toInsert + "\n" + strings.Join(lines[insertIndex:], "\n")
	return result
}

func (m *PatternMerger) insertIntoDefineMethod(content string, toInsert string) string {
	// Find the Define method and insert before its closing brace
	definePattern := regexp.MustCompile(`(?s)(public\s+override\s+void\s+Define\([^)]+\)\s*\{)(.*?)(\s*\})`)
	return definePattern.ReplaceAllString(content, fmt.Sprintf("${1}${2}\n%s${3}", toInsert))
}

func (m *PatternMerger) insertDbSets(content string, toAdd []string) string {
	// Find a good insertion point for DbSets (after existing DbSets or before OnModelCreating)
	lines := strings.Split(content, "\n")
	insertIndex := -1

	// Find last DbSet property
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "DbSet<") && strings.Contains(lines[i], "{ get; set; }") {
			insertIndex = i + 1
			break
		}
	}

	// If no DbSet found, find OnModelCreating method
	if insertIndex == -1 {
		for i, line := range lines {
			if strings.Contains(line, "protected override void OnModelCreating") {
				insertIndex = i
				break
			}
		}
	}

	if insertIndex == -1 {
		return content
	}

	// Insert DbSets
	newLines := make([]string, 0, len(lines)+len(toAdd))
	newLines = append(newLines, lines[:insertIndex]...)
	newLines = append(newLines, toAdd...)
	newLines = append(newLines, lines[insertIndex:]...)

	return strings.Join(newLines, "\n")
}
