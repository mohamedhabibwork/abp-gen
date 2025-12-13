package templates

import (
	"strings"
	"text/template"

	"github.com/gertd/go-pluralize"
)

var pluralizeClient *pluralize.Client

func init() {
	pluralizeClient = pluralize.NewClient()
}

// GetTemplateFuncs returns all custom template functions
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"pluralize":   Pluralize,
		"camelCase":   CamelCase,
		"pascalCase":  PascalCase,
		"lowerFirst":  LowerFirst,
		"upperFirst":  UpperFirst,
		"csType":      CSType,
		"nullable":    Nullable,
		"attribute":   Attribute,
		"contains":    strings.Contains,
		"hasPrefix":   strings.HasPrefix,
		"hasSuffix":   strings.HasSuffix,
		"trimSuffix":  strings.TrimSuffix,
		"trimPrefix":  strings.TrimPrefix,
		"toLower":     strings.ToLower,
		"toUpper":     strings.ToUpper,
		"join":        strings.Join,
	}
}

// Pluralize converts singular to plural
func Pluralize(word string) string {
	return pluralizeClient.Plural(word)
}

// CamelCase converts string to camelCase
func CamelCase(s string) string {
	if s == "" {
		return ""
	}
	return LowerFirst(PascalCase(s))
}

// PascalCase converts string to PascalCase
func PascalCase(s string) string {
	if s == "" {
		return ""
	}

	// Handle common separators
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	words := strings.Fields(s)
	for i, word := range words {
		words[i] = UpperFirst(strings.ToLower(word))
	}

	return strings.Join(words, "")
}

// LowerFirst lowercases the first letter
func LowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// UpperFirst uppercases the first letter
func UpperFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// CSType maps common type names to C# types
func CSType(typeName string) string {
	typeMap := map[string]string{
		"string":   "string",
		"int":      "int",
		"long":     "long",
		"decimal":  "decimal",
		"DateTime": "DateTime",
		"bool":     "bool",
		"Guid":     "Guid",
		"byte":     "byte",
		"short":    "short",
		"float":    "float",
		"double":   "double",
	}

	if csType, ok := typeMap[typeName]; ok {
		return csType
	}

	// Return as-is for custom types
	return typeName
}

// Nullable adds nullable marker for nullable types
func Nullable(typeName string, isNullable bool) string {
	if !isNullable {
		return typeName
	}

	// Value types that need nullable marker
	valueTypes := map[string]bool{
		"int": true, "long": true, "decimal": true,
		"DateTime": true, "bool": true, "Guid": true,
		"byte": true, "short": true, "float": true,
		"double": true,
	}

	if valueTypes[typeName] {
		return typeName + "?"
	}

	// Reference types (string, custom types) don't need ?
	return typeName
}

// Attribute generates C# data annotation attributes
func Attribute(attrType string, value interface{}) string {
	switch attrType {
	case "Required":
		return "[Required]"
	case "MaxLength":
		return "[MaxLength(" + value.(string) + ")]"
	case "MinLength":
		return "[MinLength(" + value.(string) + ")]"
	case "Range":
		return "[Range(" + value.(string) + ")]"
	case "StringLength":
		return "[StringLength(" + value.(string) + ")]"
	case "ForeignKey":
		return "[ForeignKey(\"" + value.(string) + "\")]"
	default:
		return "[" + attrType + "]"
	}
}

