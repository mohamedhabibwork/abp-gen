package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mohamedhabibwork/abp-gen/internal/detector"
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
	"github.com/mohamedhabibwork/abp-gen/internal/writer"
)

// ValidatorGenerator generates FluentValidation validators
type ValidatorGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewValidatorGenerator creates a new validator generator
func NewValidatorGenerator(tmplLoader *templates.Loader, w *writer.Writer) *ValidatorGenerator {
	return &ValidatorGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates all validator files for an entity
func (g *ValidatorGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType == "ValueObject" {
		return nil // Value objects don't have DTOs, so no validators
	}

	// Skip validators if using native validation (Data Annotations)
	if sch.Options.ValidationType == "native" {
		return nil
	}

	// Ensure validators directory exists
	validatorsPath := filepath.Join(paths.Application, "Validators")
	if err := os.MkdirAll(validatorsPath, 0755); err != nil {
		return fmt.Errorf("failed to create validators directory: %w", err)
	}

	// Generate Create DTO validator
	if err := g.GenerateCreateValidator(sch, entity, paths); err != nil {
		return err
	}

	// Generate Update DTO validator
	return g.GenerateUpdateValidator(sch, entity, paths)
}

// GenerateCreateValidator generates the Create DTO validator
func (g *ValidatorGenerator) GenerateCreateValidator(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("create_validator.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load create validator template: %w", err)
	}

	data := g.prepareValidatorData(sch, entity)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute create validator template: %w", err)
	}

	validatorsPath := filepath.Join(paths.Application, "Validators")
	filePath := filepath.Join(validatorsPath, "Create"+entity.Name+"DtoValidator.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// GenerateUpdateValidator generates the Update DTO validator
func (g *ValidatorGenerator) GenerateUpdateValidator(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("update_validator.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load update validator template: %w", err)
	}

	data := g.prepareValidatorData(sch, entity)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute update validator template: %w", err)
	}

	validatorsPath := filepath.Join(paths.Application, "Validators")
	filePath := filepath.Join(validatorsPath, "Update"+entity.Name+"DtoValidator.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// prepareValidatorData prepares common data for validator templates
func (g *ValidatorGenerator) prepareValidatorData(sch *schema.Schema, entity *schema.Entity) map[string]interface{} {
	return map[string]interface{}{
		"SolutionName":            sch.Solution.Name,
		"ModuleName":              sch.Solution.ModuleName,
		"NamespaceRoot":           sch.Solution.NamespaceRoot,
		"EntityName":              entity.Name,
		"Properties":              entity.Properties,
		"NonForeignKeyProperties": entity.GetNonForeignKeyProperties(),
		"ForeignKeyProperties":    entity.GetForeignKeyProperties(),
	}
}
