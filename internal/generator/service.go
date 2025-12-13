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

// ServiceGenerator generates application services
type ServiceGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewServiceGenerator creates a new service generator
func NewServiceGenerator(tmplLoader *templates.Loader, w *writer.Writer) *ServiceGenerator {
	return &ServiceGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates the application service
func (g *ServiceGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType == "ValueObject" {
		return nil // Value objects don't have application services
	}

	tmpl, err := g.tmplLoader.Load("app_service.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load app service template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":            sch.Solution.Name,
		"ModuleName":              sch.Solution.ModuleName,
		"NamespaceRoot":           sch.Solution.NamespaceRoot,
		"EntityName":              entity.Name,
		"PrimaryKeyType":          primaryKeyType,
		"EntityType":              entity.EntityType,
		"Properties":              entity.Properties,
		"NonForeignKeyProperties": entity.GetNonForeignKeyProperties(),
		"ForeignKeyProperties":    entity.GetForeignKeyProperties(),
		"HasRelations":            entity.HasRelations(),
		"OneToManyRelations":      getOneToManyRelations(entity),
		"ManyToManyRelations":     getManyToManyRelations(entity),
		"IsAggregateRoot":         entity.EntityType == "AggregateRoot" || entity.EntityType == "FullAuditedAggregateRoot",
		"HasEvents":               entity.EntityType != "ValueObject" && entity.EntityType != "Entity",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute app service template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	filePath := filepath.Join(paths.ApplicationServices, moduleFolder, entity.Name+"AppService.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// GenerateAutoMapperProfile generates AutoMapper profile
func (g *ServiceGenerator) GenerateAutoMapperProfile(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("mapper_profile.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load mapper profile template: %w", err)
	}

	data := map[string]interface{}{
		"SolutionName":  sch.Solution.Name,
		"ModuleName":    sch.Solution.ModuleName,
		"NamespaceRoot": sch.Solution.NamespaceRoot,
		"EntityName":    entity.Name,
		"IsValueObject": entity.EntityType == "ValueObject",
		"HasEvents":     entity.EntityType != "ValueObject" && entity.EntityType != "Entity",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute mapper profile template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	filePath := filepath.Join(paths.ApplicationAutoMapper, moduleFolder, entity.Name+"Profile.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// GenerateController generates HTTP API controller
func (g *ServiceGenerator) GenerateController(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType == "ValueObject" || !sch.Solution.GenerateControllers {
		return nil
	}

	tmpl, err := g.tmplLoader.Load("controller.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load controller template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":     sch.Solution.Name,
		"ModuleName":       sch.Solution.ModuleName,
		"NamespaceRoot":    sch.Solution.NamespaceRoot,
		"EntityName":       entity.Name,
		"EntityNamePlural": templates.Pluralize(entity.Name),
		"PrimaryKeyType":   primaryKeyType,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute controller template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	filePath := filepath.Join(paths.HttpApiControllers, moduleFolder, entity.Name+"Controller.cs")
	return g.writer.WriteFile(filePath, buf.String())
}
