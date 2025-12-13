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
	customPath string
	templates  map[string]*template.Template
}

// NewLoader creates a new template loader
func NewLoader(customPath string) *Loader {
	return &Loader{
		customPath: customPath,
		templates:  make(map[string]*template.Template),
	}
}

// Load loads a template by name with the following priority:
// 1. Custom path (if provided)
// 2. Extracted templates directory (./abp-gen-templates/)
// 3. Embedded templates
func (l *Loader) Load(name string) (*template.Template, error) {
	// Check if already loaded
	if tmpl, ok := l.templates[name]; ok {
		return tmpl, nil
	}

	var tmpl *template.Template
	var err error

	// Try custom path first
	if l.customPath != "" {
		tmpl, err = l.loadFromPath(filepath.Join(l.customPath, name))
		if err == nil {
			l.templates[name] = tmpl
			return tmpl, nil
		}
	}

	// Try extracted templates directory
	extractedPath := "./abp-gen-templates/" + name
	tmpl, err = l.loadFromPath(extractedPath)
	if err == nil {
		l.templates[name] = tmpl
		return tmpl, nil
	}

	// Fall back to embedded templates
	tmpl, err = l.loadFromEmbedded(name)
	if err != nil {
		return nil, fmt.Errorf("template '%s' not found in any location: %w", name, err)
	}

	l.templates[name] = tmpl
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

