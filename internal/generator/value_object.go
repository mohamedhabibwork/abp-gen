package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mohamedhabibwork/abp-gen/internal/detector"
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
	"github.com/mohamedhabibwork/abp-gen/internal/writer"
)

// ValueObjectGenerator generates enhanced value objects
type ValueObjectGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewValueObjectGenerator creates a new value object generator
func NewValueObjectGenerator(tmplLoader *templates.Loader, w *writer.Writer) *ValueObjectGenerator {
	return &ValueObjectGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates value object with enhanced features
func (g *ValueObjectGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType != "ValueObject" {
		return nil
	}

	// Load enhanced value object template
	tmpl, err := g.tmplLoader.Load("value_object_enhanced.tmpl")
	if err != nil {
		// If template not found, skip value object generation gracefully
		errStr := err.Error()
		if strings.Contains(errStr, "not found") || 
		   strings.Contains(errStr, "file does not exist") ||
		   strings.Contains(errStr, "no such file") {
			return nil // Skip value object generation if template doesn't exist
		}
		return fmt.Errorf("failed to load value_object_enhanced template: %w", err)
	}

	// Prepare config with defaults
	config := entity.ValueObjectConfig
	if config == nil {
		config = &schema.ValueObjectConfig{
			IsImmutable:        true,
			GenerateComparison: false,
		}
	}

	// Default equality members to all properties if not specified
	if len(config.EqualityMembers) == 0 {
		for _, prop := range entity.Properties {
			config.EqualityMembers = append(config.EqualityMembers, prop.Name)
		}
	}

	data := map[string]interface{}{
		"SolutionName":         sch.Solution.Name,
		"ModuleName":           sch.Solution.ModuleName,
		"ModuleNameWithSuffix": sch.Solution.GetModuleNameWithSuffix(),
		"NamespaceRoot":        sch.Solution.NamespaceRoot,
		"EntityName":           entity.Name,
		"Properties":           entity.Properties,
		"Config":               config,
		"TargetFramework":      sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute value_object_enhanced template: %w", err)
	}

	moduleFolder := sch.Solution.GetModuleFolderName()
	voPath := filepath.Join(paths.DomainEntities, moduleFolder, entity.Name+".cs")
	return g.writer.WriteFile(voPath, buf.String())
}

// GenerateFactory generates factory method for value object if configured
func (g *ValueObjectGenerator) GenerateFactory(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType != "ValueObject" || entity.ValueObjectConfig == nil || entity.ValueObjectConfig.FactoryMethod == "" {
		return nil
	}

	tmpl, err := g.tmplLoader.Load("value_object_factory.tmpl")
	if err != nil {
		// Factory template is optional
		return nil
	}

	data := map[string]interface{}{
		"SolutionName":         sch.Solution.Name,
		"ModuleName":           sch.Solution.ModuleName,
		"ModuleNameWithSuffix": sch.Solution.GetModuleNameWithSuffix(),
		"NamespaceRoot":        sch.Solution.NamespaceRoot,
		"EntityName":           entity.Name,
		"FactoryMethod":        entity.ValueObjectConfig.FactoryMethod,
		"Properties":           entity.Properties,
		"ValidationRules":      entity.ValueObjectConfig.ValidationRules,
		"TargetFramework":      sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute value_object_factory template: %w", err)
	}

	moduleFolder := sch.Solution.GetModuleFolderName()
	factoryPath := filepath.Join(paths.DomainEntities, moduleFolder, entity.Name+"Factory.cs")
	return g.writer.WriteFile(factoryPath, buf.String())
}
