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

// CustomRepositoryGenerator generates custom repository interfaces and implementations
type CustomRepositoryGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewCustomRepositoryGenerator creates a new custom repository generator
func NewCustomRepositoryGenerator(tmplLoader *templates.Loader, w *writer.Writer) *CustomRepositoryGenerator {
	return &CustomRepositoryGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates custom repository interface and implementation
func (g *CustomRepositoryGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.CustomRepository == nil || len(entity.CustomRepository.Methods) == 0 {
		return nil // No custom repository defined
	}

	// Generate interface
	if err := g.generateInterface(sch, entity, paths); err != nil {
		return err
	}

	// Generate EF Core implementation if applicable
	if sch.Solution.DBProvider == "efcore" || sch.Solution.DBProvider == "both" {
		if err := g.generateEFCoreImplementation(sch, entity, paths); err != nil {
			return err
		}
	}

	// Generate MongoDB implementation if applicable
	if sch.Solution.DBProvider == "mongodb" || sch.Solution.DBProvider == "both" {
		if err := g.generateMongoDBImplementation(sch, entity, paths); err != nil {
			return err
		}
	}

	return nil
}

func (g *CustomRepositoryGenerator) generateInterface(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("repository_custom.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load repository_custom template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"EntityName":      entity.Name,
		"PrimaryKeyType":  primaryKeyType,
		"Methods":         entity.CustomRepository.Methods,
		"TargetFramework": sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute repository_custom template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	repoPath := filepath.Join(paths.DomainRepositories, moduleFolder, "I"+entity.Name+"CustomRepository.cs")
	return g.writer.WriteFile(repoPath, buf.String())
}

func (g *CustomRepositoryGenerator) generateEFCoreImplementation(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("repository_impl_ef.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load repository_impl_ef template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"EntityName":      entity.Name,
		"PrimaryKeyType":  primaryKeyType,
		"Methods":         entity.CustomRepository.Methods,
		"TargetFramework": sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute repository_impl_ef template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	repoPath := filepath.Join(paths.EFCoreRepositories, moduleFolder, entity.Name+"CustomRepository.cs")
	return g.writer.WriteFile(repoPath, buf.String())
}

func (g *CustomRepositoryGenerator) generateMongoDBImplementation(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("repository_impl_mongo.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load repository_impl_mongo template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"EntityName":      entity.Name,
		"PrimaryKeyType":  primaryKeyType,
		"Methods":         entity.CustomRepository.Methods,
		"TargetFramework": sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute repository_impl_mongo template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	repoPath := filepath.Join(paths.MongoDBRepositories, moduleFolder, entity.Name+"CustomRepository.cs")
	return g.writer.WriteFile(repoPath, buf.String())
}
