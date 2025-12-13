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

// EntityGenerator generates domain entities
type EntityGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewEntityGenerator creates a new entity generator
func NewEntityGenerator(tmplLoader *templates.Loader, w *writer.Writer) *EntityGenerator {
	return &EntityGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates the entity file
func (g *EntityGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	// Load template
	tmpl, err := g.tmplLoader.Load("entity.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load entity template: %w", err)
	}

	// Prepare template data
	data := g.prepareEntityData(sch, entity)

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute entity template: %w", err)
	}

	// Write file
	entityPath := filepath.Join(paths.DomainEntities, entity.Name+".cs")
	return g.writer.WriteFile(entityPath, buf.String())
}

// prepareEntityData prepares data for the entity template
func (g *EntityGenerator) prepareEntityData(sch *schema.Schema, entity *schema.Entity) map[string]interface{} {
	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)
	
	return map[string]interface{}{
		"SolutionName":             sch.Solution.Name,
		"ModuleName":               sch.Solution.ModuleName,
		"NamespaceRoot":            sch.Solution.NamespaceRoot,
		"EntityName":               entity.Name,
		"TableName":                entity.TableName,
		"EntityType":               entity.EntityType,
		"PrimaryKeyType":           primaryKeyType,
		"Properties":               entity.Properties,
		"NonForeignKeyProperties":  entity.GetNonForeignKeyProperties(),
		"ForeignKeyProperties":     entity.GetForeignKeyProperties(),
		"HasRelations":             entity.HasRelations(),
		"Relations":                entity.Relations,
		"OneToManyRelations":       getOneToManyRelations(entity),
		"ManyToManyRelations":      getManyToManyRelations(entity),
		"HasEvents":                entity.EntityType != "ValueObject" && entity.EntityType != "Entity",
		"IsValueObject":            entity.EntityType == "ValueObject",
		"IsAggregateRoot":          entity.EntityType == "AggregateRoot" || entity.EntityType == "FullAuditedAggregateRoot",
	}
}

// GenerateRepository generates repository interface
func (g *EntityGenerator) GenerateRepository(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType == "ValueObject" {
		return nil // Value objects don't have repositories
	}

	// Load template
	tmpl, err := g.tmplLoader.Load("repository.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load repository template: %w", err)
	}

	// Prepare template data
	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)
	
	data := map[string]interface{}{
		"SolutionName":     sch.Solution.Name,
		"ModuleName":       sch.Solution.ModuleName,
		"NamespaceRoot":    sch.Solution.NamespaceRoot,
		"EntityName":       entity.Name,
		"PrimaryKeyType":   primaryKeyType,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute repository template: %w", err)
	}

	// Write file
	repoPath := filepath.Join(paths.DomainRepositories, "I"+entity.Name+"Repository.cs")
	return g.writer.WriteFile(repoPath, buf.String())
}

// GenerateConstants generates entity constants
func (g *EntityGenerator) GenerateConstants(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	// Load template
	tmpl, err := g.tmplLoader.Load("constants.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load constants template: %w", err)
	}

	// Prepare validation constants
	validationConstants := make(map[string]int)
	for _, prop := range entity.Properties {
		if prop.MaxLength > 0 {
			validationConstants[prop.Name+"MaxLength"] = prop.MaxLength
		}
	}

	data := map[string]interface{}{
		"SolutionName":         sch.Solution.Name,
		"ModuleName":           sch.Solution.ModuleName,
		"NamespaceRoot":        sch.Solution.NamespaceRoot,
		"EntityName":           entity.Name,
		"ValidationConstants":  validationConstants,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute constants template: %w", err)
	}

	// Write file
	constantsPath := filepath.Join(paths.DomainSharedConstants, entity.Name+"Constants.cs")
	return g.writer.WriteFile(constantsPath, buf.String())
}

// GenerateEvents generates event types and ETOs
func (g *EntityGenerator) GenerateEvents(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType == "ValueObject" || entity.EntityType == "Entity" {
		return nil // These types don't generate distributed events
	}

	// Generate event types
	if err := g.generateEventTypes(sch, entity, paths); err != nil {
		return err
	}

	// Generate ETO
	return g.generateETO(sch, entity, paths)
}

func (g *EntityGenerator) generateEventTypes(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("event_types.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load event types template: %w", err)
	}

	data := map[string]interface{}{
		"SolutionName":  sch.Solution.Name,
		"ModuleName":    sch.Solution.ModuleName,
		"NamespaceRoot": sch.Solution.NamespaceRoot,
		"EntityName":    entity.Name,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute event types template: %w", err)
	}

	eventTypesPath := filepath.Join(paths.DomainSharedEvents, entity.Name+"EtoTypes.cs")
	return g.writer.WriteFile(eventTypesPath, buf.String())
}

func (g *EntityGenerator) generateETO(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("eto.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load ETO template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"EntityName":      entity.Name,
		"PrimaryKeyType":  primaryKeyType,
		"Properties":      entity.Properties,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute ETO template: %w", err)
	}

	etoPath := filepath.Join(paths.DomainSharedEvents, entity.Name+"Eto.cs")
	return g.writer.WriteFile(etoPath, buf.String())
}

// GenerateDataSeeder generates data seeder
func (g *EntityGenerator) GenerateDataSeeder(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType == "ValueObject" {
		return nil
	}

	tmpl, err := g.tmplLoader.Load("seeder.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load seeder template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":             sch.Solution.Name,
		"ModuleName":               sch.Solution.ModuleName,
		"NamespaceRoot":            sch.Solution.NamespaceRoot,
		"EntityName":               entity.Name,
		"PrimaryKeyType":           primaryKeyType,
		"NonForeignKeyProperties":  entity.GetNonForeignKeyProperties(),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute seeder template: %w", err)
	}

	seederPath := filepath.Join(paths.DomainData, entity.Name+"DataSeeder.cs")
	return g.writer.WriteFile(seederPath, buf.String())
}

func getOneToManyRelations(entity *schema.Entity) []schema.OneToManyRelation {
	if entity.Relations == nil {
		return nil
	}
	return entity.Relations.OneToMany
}

func getManyToManyRelations(entity *schema.Entity) []schema.ManyToManyRelation {
	if entity.Relations == nil {
		return nil
	}
	return entity.Relations.ManyToMany
}

