package generator

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/mohamedhabibwork/abp-gen/internal/detector"
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
	"github.com/mohamedhabibwork/abp-gen/internal/writer"
)

// ManagerGenerator generates domain managers
type ManagerGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewManagerGenerator creates a new manager generator
func NewManagerGenerator(tmplLoader *templates.Loader, w *writer.Writer) *ManagerGenerator {
	return &ManagerGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates the domain manager
func (g *ManagerGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType == "ValueObject" {
		return nil // Value objects don't have managers
	}

	tmpl, err := g.tmplLoader.Load("manager.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load manager template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":            sch.Solution.Name,
		"ModuleName":              sch.Solution.ModuleName,
		"NamespaceRoot":           sch.Solution.NamespaceRoot,
		"EntityName":              entity.Name,
		"PrimaryKeyType":          primaryKeyType,
		"Properties":              entity.Properties,
		"NonForeignKeyProperties": entity.GetNonForeignKeyProperties(),
		"HasRelations":            entity.HasRelations(),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute manager template: %w", err)
	}

	// Managers are typically in Domain/Managers directory
	moduleFolder := sch.Solution.ModuleName + "Module"
	managerPath := filepath.Join(paths.Domain, "Managers", moduleFolder, entity.Name+"Manager.cs")
	return g.writer.WriteFile(managerPath, buf.String())
}
