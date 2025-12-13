package merger

import (
	"regexp"
	"strings"
)

// CSharpClass represents a parsed C# class
type CSharpClass struct {
	Name       string
	BaseClass  string
	Properties []CSharpProperty
	Methods    []CSharpMethod
	RawContent string
	StartLine  int
	EndLine    int
}

// CSharpProperty represents a C# property
type CSharpProperty struct {
	Name         string
	Type         string
	Attributes   []string
	GetSet       string
	RawContent   string
	Line         int
}

// CSharpMethod represents a C# method
type CSharpMethod struct {
	Name       string
	ReturnType string
	Parameters string
	Attributes []string
	Body       string
	RawContent string
	Signature  string
	StartLine  int
	EndLine    int
}

// CSharpParser parses C# code structures
type CSharpParser struct{}

// NewCSharpParser creates a new C# parser
func NewCSharpParser() *CSharpParser {
	return &CSharpParser{}
}

// ParseClass parses a C# class from source code
func (p *CSharpParser) ParseClass(content string) (*CSharpClass, error) {
	// Extract class declaration
	classPattern := regexp.MustCompile(`(?s)public\s+(?:(?:abstract|sealed|static)\s+)?class\s+(\w+)(?:\s*:\s*([^{]+))?\s*\{`)
	matches := classPattern.FindStringSubmatch(content)
	
	if len(matches) < 2 {
		return nil, nil
	}
	
	class := &CSharpClass{
		Name:       matches[1],
		RawContent: content,
	}
	
	if len(matches) > 2 {
		class.BaseClass = strings.TrimSpace(matches[2])
	}
	
	// Parse properties
	class.Properties = p.parseProperties(content)
	
	// Parse methods
	class.Methods = p.parseMethods(content)
	
	return class, nil
}

// parseProperties extracts all properties from C# code
func (p *CSharpParser) parseProperties(content string) []CSharpProperty {
	var properties []CSharpProperty
	
	// Pattern for properties with get/set
	propertyPattern := regexp.MustCompile(`(?m)^\s*(?:\[([^\]]+)\]\s*)*public\s+(\w+(?:<[^>]+>)?(?:\?)?)\s+(\w+)\s*\{\s*(get;\s*set;|get;\s*private\s+set;|get;\s*init;)\s*\}`)
	
	matches := propertyPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		prop := CSharpProperty{
			Type:       match[2],
			Name:       match[3],
			GetSet:     match[4],
			RawContent: match[0],
		}
		
		if match[1] != "" {
			prop.Attributes = strings.Split(match[1], "][")
		}
		
		properties = append(properties, prop)
	}
	
	return properties
}

// parseMethods extracts all methods from C# code
func (p *CSharpParser) parseMethods(content string) []CSharpMethod {
	var methods []CSharpMethod
	
	// Pattern for method signatures
	methodPattern := regexp.MustCompile(`(?s)(?:\[([^\]]+)\]\s*)*public\s+(?:(?:virtual|override|async|static)\s+)*(\w+(?:<[^>]+>)?)\s+(\w+)\s*\(([^)]*)\)`)
	
	matches := methodPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		// Extract method body
		methodIndex := strings.Index(content, match[0])
		if methodIndex == -1 {
			continue
		}
		
		body := p.extractMethodBody(content[methodIndex:])
		
		method := CSharpMethod{
			ReturnType: match[2],
			Name:       match[3],
			Parameters: match[4],
			RawContent: match[0] + body,
			Signature:  p.createMethodSignature(match[2], match[3], match[4]),
		}
		
		if match[1] != "" {
			method.Attributes = strings.Split(match[1], "][")
		}
		
		methods = append(methods, method)
	}
	
	return methods
}

// extractMethodBody extracts the method body including braces
func (p *CSharpParser) extractMethodBody(content string) string {
	braceCount := 0
	inBody := false
	bodyStart := -1
	
	for i, ch := range content {
		if ch == '{' {
			if !inBody {
				bodyStart = i
				inBody = true
			}
			braceCount++
		} else if ch == '}' {
			braceCount--
			if braceCount == 0 && inBody {
				return content[bodyStart : i+1]
			}
		}
	}
	
	return ""
}

// createMethodSignature creates a normalized method signature for comparison
func (p *CSharpParser) createMethodSignature(returnType, name, parameters string) string {
	// Normalize parameters (remove parameter names, keep only types)
	paramParts := strings.Split(parameters, ",")
	var paramTypes []string
	
	for _, part := range paramParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		// Extract type (everything before the last word/identifier)
		parts := strings.Fields(part)
		if len(parts) > 0 {
			paramTypes = append(paramTypes, parts[0])
		}
	}
	
	return returnType + " " + name + "(" + strings.Join(paramTypes, ",") + ")"
}

// FindProperty finds a property by name in a class
func (p *CSharpParser) FindProperty(class *CSharpClass, propertyName string) *CSharpProperty {
	for i := range class.Properties {
		if class.Properties[i].Name == propertyName {
			return &class.Properties[i]
		}
	}
	return nil
}

// FindMethod finds a method by signature in a class
func (p *CSharpParser) FindMethod(class *CSharpClass, signature string) *CSharpMethod {
	for i := range class.Methods {
		if class.Methods[i].Signature == signature {
			return &class.Methods[i]
		}
	}
	return nil
}

// FindMethodByName finds a method by name (ignoring signature)
func (p *CSharpParser) FindMethodByName(class *CSharpClass, name string) *CSharpMethod {
	for i := range class.Methods {
		if class.Methods[i].Name == name {
			return &class.Methods[i]
		}
	}
	return nil
}

// HasProperty checks if a class has a property with the given name
func (p *CSharpParser) HasProperty(class *CSharpClass, propertyName string) bool {
	return p.FindProperty(class, propertyName) != nil
}

// HasMethod checks if a class has a method with the given signature
func (p *CSharpParser) HasMethod(class *CSharpClass, signature string) bool {
	return p.FindMethod(class, signature) != nil
}

// ExtractNamespace extracts the namespace from C# code
func (p *CSharpParser) ExtractNamespace(content string) string {
	namespacePattern := regexp.MustCompile(`namespace\s+([\w.]+)`)
	matches := namespacePattern.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ExtractUsings extracts all using statements from C# code
func (p *CSharpParser) ExtractUsings(content string) []string {
	usingPattern := regexp.MustCompile(`(?m)^using\s+([^;]+);`)
	matches := usingPattern.FindAllStringSubmatch(content, -1)
	
	var usings []string
	for _, match := range matches {
		if len(match) > 1 {
			usings = append(usings, strings.TrimSpace(match[1]))
		}
	}
	
	return usings
}

// MergeUsings merges two sets of using statements, removing duplicates
func (p *CSharpParser) MergeUsings(existing []string, new []string) []string {
	usingSet := make(map[string]bool)
	var merged []string
	
	// Add existing
	for _, u := range existing {
		if !usingSet[u] {
			usingSet[u] = true
			merged = append(merged, u)
		}
	}
	
	// Add new
	for _, u := range new {
		if !usingSet[u] {
			usingSet[u] = true
			merged = append(merged, u)
		}
	}
	
	return merged
}

