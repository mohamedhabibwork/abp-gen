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

// MongoDBGenerator generates MongoDB files
type MongoDBGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewMongoDBGenerator creates a new MongoDB generator
func NewMongoDBGenerator(tmplLoader *templates.Loader, w *writer.Writer) *MongoDBGenerator {
	return &MongoDBGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates MongoDB repository
func (g *MongoDBGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType == "ValueObject" || paths.MongoDB == "" {
		return nil
	}

	return g.GenerateRepository(sch, entity, paths)
}

// GenerateRepository generates MongoDB repository implementation
func (g *MongoDBGenerator) GenerateRepository(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("mongodb_repository.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load MongoDB repository template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":   sch.Solution.Name,
		"ModuleName":     sch.Solution.ModuleName,
		"NamespaceRoot":  sch.Solution.NamespaceRoot,
		"EntityName":     entity.Name,
		"PrimaryKeyType": primaryKeyType,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute MongoDB repository template: %w", err)
	}

	filePath := filepath.Join(paths.MongoDBRepositories, "Mongo"+entity.Name+"Repository.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// GenerateConfiguration generates MongoDB entity configuration (collection mapping)
func (g *MongoDBGenerator) GenerateConfiguration(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("mongodb_config.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load MongoDB config template: %w", err)
	}

	data := map[string]interface{}{
		"SolutionName":   sch.Solution.Name,
		"ModuleName":     sch.Solution.ModuleName,
		"NamespaceRoot":  sch.Solution.NamespaceRoot,
		"EntityName":     entity.Name,
		"TableName":      entity.TableName,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute MongoDB config template: %w", err)
	}

	// MongoDB configurations are typically in a different location
	// Adjust path as needed based on ABP MongoDB structure
	configPath := filepath.Join(paths.MongoDB, "MongoDB", entity.Name+"MongoDbConfiguration.cs")
	return g.writer.WriteFile(configPath, buf.String())
}

