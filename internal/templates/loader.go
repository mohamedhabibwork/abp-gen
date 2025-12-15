package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed *.tmpl
var embeddedTemplates embed.FS

// Loader manages template loading from various sources
type Loader struct {
	customPath      string
	targetFramework string // Target framework: "aspnetcore9", "abp8-microservice", "abp8-monolith"
	templates       map[string]*template.Template
}

// NewLoader creates a new template loader
func NewLoader(customPath string) *Loader {
	return &Loader{
		customPath:      customPath,
		targetFramework: "abp8-monolith", // Default target
		templates:       make(map[string]*template.Template),
	}
}

// NewLoaderWithTarget creates a new template loader with specific target framework
func NewLoaderWithTarget(customPath string, targetFramework string) *Loader {
	if targetFramework == "" {
		targetFramework = "abp8-monolith"
	}
	return &Loader{
		customPath:      customPath,
		targetFramework: targetFramework,
		templates:       make(map[string]*template.Template),
	}
}

// SetTargetFramework sets the target framework for template loading
func (l *Loader) SetTargetFramework(target string) {
	l.targetFramework = target
	// Clear cached templates when target changes
	l.templates = make(map[string]*template.Template)
}

// Load loads a template by name with the following priority:
// 1. Target-specific custom path (if provided)
// 2. Target-specific extracted templates directory
// 3. Target-specific embedded templates
// 4. Common/shared templates (fallback)
func (l *Loader) Load(name string) (*template.Template, error) {
	// Check if already loaded
	cacheKey := l.targetFramework + ":" + name
	if tmpl, ok := l.templates[cacheKey]; ok {
		return tmpl, nil
	}

	var tmpl *template.Template
	var err error

	// Try target-specific custom path first
	if l.customPath != "" {
		targetPath := filepath.Join(l.customPath, l.targetFramework, name)
		tmpl, err = l.loadFromPath(targetPath)
		if err == nil {
			l.templates[cacheKey] = tmpl
			return tmpl, nil
		}

		// Try common custom path
		commonPath := filepath.Join(l.customPath, "common", name)
		tmpl, err = l.loadFromPath(commonPath)
		if err == nil {
			l.templates[cacheKey] = tmpl
			return tmpl, nil
		}
	}

	// Try target-specific extracted templates directory
	extractedPath := filepath.Join("./abp-gen-templates/", l.targetFramework, name)
	tmpl, err = l.loadFromPath(extractedPath)
	if err == nil {
		l.templates[cacheKey] = tmpl
		return tmpl, nil
	}

	// Try common extracted templates
	commonExtractedPath := filepath.Join("./abp-gen-templates/common", name)
	tmpl, err = l.loadFromPath(commonExtractedPath)
	if err == nil {
		l.templates[cacheKey] = tmpl
		return tmpl, nil
	}

	// Try target-specific embedded templates
	targetEmbedPath := filepath.Join(l.targetFramework, name)
	tmpl, err = l.loadFromEmbedded(targetEmbedPath)
	if err == nil {
		l.templates[cacheKey] = tmpl
		return tmpl, nil
	}

	// Fall back to common embedded templates
	commonEmbedPath := filepath.Join("common", name)
	tmpl, err = l.loadFromEmbedded(commonEmbedPath)
	if err == nil {
		l.templates[cacheKey] = tmpl
		return tmpl, nil
	}

	// Last resort: try loading from root (backward compatibility)
	tmpl, err = l.loadFromEmbedded(name)
	if err != nil {
		return nil, fmt.Errorf("template '%s' not found for target '%s' in any location: %w", name, l.targetFramework, err)
	}

	l.templates[cacheKey] = tmpl
	return tmpl, nil
}

// loadFromPath loads template from filesystem
func (l *Loader) loadFromPath(path string) (*template.Template, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(filepath.Base(path)).
		Funcs(GetTemplateFuncs()).
		Parse(string(content))
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// loadFromEmbedded loads template from embedded filesystem
func (l *Loader) loadFromEmbedded(name string) (*template.Template, error) {
	content, err := embeddedTemplates.ReadFile(name)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(name).
		Funcs(GetTemplateFuncs()).
		Parse(string(content))
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// ExtractTemplates extracts embedded templates to a directory
func ExtractTemplates(destPath string) error {
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return err
	}

	return fs.WalkDir(embeddedTemplates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		content, err := embeddedTemplates.ReadFile(path)
		if err != nil {
			return err
		}

		destFile := filepath.Join(destPath, path)
		if err := os.WriteFile(destFile, content, 0644); err != nil {
			return err
		}

		fmt.Printf("Extracted: %s\n", destFile)
		return nil
	})
}

// ListAvailableTemplates lists all available template names
func (l *Loader) ListAvailableTemplates() ([]string, error) {
	var templates []string

	err := fs.WalkDir(embeddedTemplates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".tmpl" {
			templates = append(templates, path)
		}

		return nil
	})

	return templates, err
}
