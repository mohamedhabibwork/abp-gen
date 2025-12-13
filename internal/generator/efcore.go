package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/mohamedhabibwork/abp-gen/internal/detector"
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
	"github.com/mohamedhabibwork/abp-gen/internal/writer"
)

// EFCoreGenerator generates Entity Framework Core files
type EFCoreGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewEFCoreGenerator creates a new EF Core generator
func NewEFCoreGenerator(tmplLoader *templates.Loader, w *writer.Writer) *EFCoreGenerator {
	return &EFCoreGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates EF Core configuration and repository
func (g *EFCoreGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType == "ValueObject" {
		return nil
	}

	// Generate entity configuration
	if err := g.GenerateConfiguration(sch, entity, paths); err != nil {
		return err
	}

	// Generate repository implementation
	if err := g.GenerateRepository(sch, entity, paths); err != nil {
		return err
	}

	// Update DbContext
	if err := g.UpdateDbContext(sch, entity, paths); err != nil {
		return err
	}

	// Update IDbContext
	return g.UpdateIDbContext(sch, entity, paths)
}

// GenerateConfiguration generates EF Core entity configuration
func (g *EFCoreGenerator) GenerateConfiguration(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("efcore_config.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load EF Core config template: %w", err)
	}

	data := map[string]interface{}{
		"SolutionName":        sch.Solution.Name,
		"ModuleName":          sch.Solution.ModuleName,
		"NamespaceRoot":       sch.Solution.NamespaceRoot,
		"EntityName":          entity.Name,
		"TableName":           entity.TableName,
		"Properties":          entity.Properties,
		"HasRelations":        entity.HasRelations(),
		"OneToManyRelations":  getOneToManyRelations(entity),
		"ManyToManyRelations": getManyToManyRelations(entity),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute EF Core config template: %w", err)
	}

	filePath := filepath.Join(paths.EFCoreConfigurations, entity.Name+"Configuration.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// GenerateRepository generates EF Core repository implementation
func (g *EFCoreGenerator) GenerateRepository(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("efcore_repository.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load EF Core repository template: %w", err)
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
		return fmt.Errorf("failed to execute EF Core repository template: %w", err)
	}

	filePath := filepath.Join(paths.EFCoreRepositories, "EfCore"+entity.Name+"Repository.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// UpdateDbContext updates the DbContext to add DbSet
func (g *EFCoreGenerator) UpdateDbContext(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	dbContextPath := paths.GetDbContextPath(sch.Solution.ModuleName)

	// Check if DbSet already exists
	searchPattern := fmt.Sprintf("DbSet<%s>", entity.Name)

	entityPlural := templates.Pluralize(entity.Name)
	dbSetProperty := fmt.Sprintf("\n    public virtual DbSet<%s> %s { get; set; }\n", entity.Name, entityPlural)

	return g.writer.UpdateFileIdempotent(dbContextPath, searchPattern, func(content string) (string, error) {
		// Find the DbContext constructor and insert DbSet before it
		pattern := regexp.MustCompile(`(\s+)(public\s+\w+DbContext\()`)
		if !pattern.MatchString(content) {
			return "", fmt.Errorf("DbContext constructor not found")
		}

		updated := pattern.ReplaceAllString(content, dbSetProperty+"$1$2")
		return updated, nil
	})
}

// UpdateIDbContext updates the IDbContext interface to add DbSet
func (g *EFCoreGenerator) UpdateIDbContext(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	idbContextPath := paths.GetIDbContextPath(sch.Solution.ModuleName)

	// Check if DbSet already exists
	searchPattern := fmt.Sprintf("DbSet<%s>", entity.Name)

	entityPlural := templates.Pluralize(entity.Name)
	dbSetProperty := fmt.Sprintf("\n    DbSet<%s> %s { get; }\n", entity.Name, entityPlural)

	return g.writer.UpdateFileIdempotent(idbContextPath, searchPattern, func(content string) (string, error) {
		// Find the closing brace of the interface and insert before it
		pattern := regexp.MustCompile(`(\s+)(}\s*$)`)
		if !pattern.MatchString(content) {
			return "", fmt.Errorf("interface closing brace not found")
		}

		updated := pattern.ReplaceAllString(content, dbSetProperty+"$1$2")
		return updated, nil
	})
}

// UpdateModelCreating adds entity configuration to OnModelCreating
func (g *EFCoreGenerator) UpdateModelCreating(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	dbContextPath := paths.GetDbContextPath(sch.Solution.ModuleName)

	searchPattern := fmt.Sprintf("ApplyConfiguration(new %sConfiguration())", entity.Name)

	configLine := fmt.Sprintf("\n            builder.ApplyConfiguration(new %sConfiguration());\n", entity.Name)

	return g.writer.UpdateFileIdempotent(dbContextPath, searchPattern, func(content string) (string, error) {
		// Find OnModelCreating method and insert configuration
		pattern := regexp.MustCompile(`(protected override void OnModelCreating\(ModelBuilder builder\)\s*{)`)
		if !pattern.MatchString(content) {
			return "", fmt.Errorf("OnModelCreating method not found")
		}

		updated := pattern.ReplaceAllString(content, "$1"+configLine)
		return updated, nil
	})
}
