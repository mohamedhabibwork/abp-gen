package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/mohamedhabibwork/abp-gen/internal/detector"
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
	"github.com/mohamedhabibwork/abp-gen/internal/writer"
)

// IntegrationTestGenerator generates integration tests
type IntegrationTestGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewIntegrationTestGenerator creates a new integration test generator
func NewIntegrationTestGenerator(tmplLoader *templates.Loader, w *writer.Writer) *IntegrationTestGenerator {
	return &IntegrationTestGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates integration tests for an entity
func (g *IntegrationTestGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if !entity.GenerateIntegrationTests && !sch.Options.GenerateIntegrationTests {
		return nil // Integration tests not enabled
	}

	// Skip value objects
	if entity.EntityType == "ValueObject" {
		return nil
	}

	// Generate test base class (once)
	if err := g.generateTestBase(sch, paths); err != nil {
		return fmt.Errorf("failed to generate test base: %w", err)
	}

	// Generate repository tests
	if err := g.generateRepositoryTests(sch, entity, paths); err != nil {
		return fmt.Errorf("failed to generate repository tests: %w", err)
	}

	// Generate service tests
	if err := g.generateServiceTests(sch, entity, paths); err != nil {
		return fmt.Errorf("failed to generate service tests: %w", err)
	}

	// Generate domain tests (for aggregate roots with domain logic)
	if entity.EntityType == "AggregateRoot" || entity.EntityType == "FullAuditedAggregateRoot" {
		if err := g.generateDomainTests(sch, entity, paths); err != nil {
			return fmt.Errorf("failed to generate domain tests: %w", err)
		}
	}

	return nil
}

func (g *IntegrationTestGenerator) generateTestBase(sch *schema.Schema, paths *detector.LayerPaths) error {
	var tmpl *template.Template
	var err error

	// Choose template based on target framework
	switch sch.Solution.TargetFramework {
	case schema.TargetASPNETCore9:
		tmpl, err = g.tmplLoader.Load("test_base_aspnetcore.tmpl")
	case schema.TargetABP8Microservice:
		tmpl, err = g.tmplLoader.Load("test_base_abp_microservice.tmpl")
	default:
		tmpl, err = g.tmplLoader.Load("test_base_abp.tmpl")
	}

	if err != nil {
		return fmt.Errorf("failed to load test base template: %w", err)
	}

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"TargetFramework": sch.Solution.TargetFramework,
		"DBProvider":      sch.Solution.DBProvider,
		"MultiTenancy":    sch.Solution.MultiTenancy,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute test base template: %w", err)
	}

	// Determine test project path
	testPath := g.getTestProjectPath(paths, sch)
	baseTestPath := filepath.Join(testPath, sch.Solution.ModuleName+"TestBase.cs")
	return g.writer.WriteFile(baseTestPath, buf.String())
}

func (g *IntegrationTestGenerator) generateRepositoryTests(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("integration_test_repository.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load repository test template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":     sch.Solution.Name,
		"ModuleName":       sch.Solution.ModuleName,
		"NamespaceRoot":    sch.Solution.NamespaceRoot,
		"EntityName":       entity.Name,
		"PrimaryKeyType":   primaryKeyType,
		"Properties":       entity.Properties,
		"TargetFramework":  sch.Solution.TargetFramework,
		"CustomRepository": entity.CustomRepository,
		"Relations":        entity.Relations,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute repository test template: %w", err)
	}

	testPath := g.getTestProjectPath(paths, sch)
	moduleFolder := sch.Solution.ModuleName + "Module"
	repoTestPath := filepath.Join(testPath, "Repositories", moduleFolder, entity.Name+"RepositoryTests.cs")
	return g.writer.WriteFile(repoTestPath, buf.String())
}

func (g *IntegrationTestGenerator) generateServiceTests(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("integration_test_service.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load service test template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"EntityName":      entity.Name,
		"PrimaryKeyType":  primaryKeyType,
		"Properties":      entity.Properties,
		"TargetFramework": sch.Solution.TargetFramework,
		"Relations":       entity.Relations,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute service test template: %w", err)
	}

	testPath := g.getTestProjectPath(paths, sch)
	moduleFolder := sch.Solution.ModuleName + "Module"
	serviceTestPath := filepath.Join(testPath, "Services", moduleFolder, entity.Name+"ServiceTests.cs")
	return g.writer.WriteFile(serviceTestPath, buf.String())
}

func (g *IntegrationTestGenerator) generateDomainTests(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("integration_test_domain.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load domain test template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"EntityName":      entity.Name,
		"PrimaryKeyType":  primaryKeyType,
		"Properties":      entity.Properties,
		"TargetFramework": sch.Solution.TargetFramework,
		"DomainEvents":    entity.DomainEvents,
		"Manager":         entity.EntityType == "AggregateRoot" || entity.EntityType == "FullAuditedAggregateRoot",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute domain test template: %w", err)
	}

	testPath := g.getTestProjectPath(paths, sch)
	moduleFolder := sch.Solution.ModuleName + "Module"
	domainTestPath := filepath.Join(testPath, "Domain", moduleFolder, entity.Name+"Tests.cs")
	return g.writer.WriteFile(domainTestPath, buf.String())
}

func (g *IntegrationTestGenerator) getTestProjectPath(paths *detector.LayerPaths, sch *schema.Schema) string {
	// For test projects, we typically have:
	// - For monolith: test/{Solution}.{Module}.Tests/
	// - For microservice: test/{Solution}.{Module}.Service.Tests/

	basePath := filepath.Dir(paths.Domain)
	testProjectName := fmt.Sprintf("%s.%s.Tests", sch.Solution.Name, sch.Solution.ModuleName)

	if sch.Solution.TargetFramework == schema.TargetABP8Microservice {
		testProjectName = fmt.Sprintf("%s.%s.Service.Tests", sch.Solution.Name, sch.Solution.ModuleName)
	}

	return filepath.Join(basePath, "..", "test", testProjectName)
}

// GenerateTestProject generates the test project file if it doesn't exist
func (g *IntegrationTestGenerator) GenerateTestProject(sch *schema.Schema, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("test_project.tmpl")
	if err != nil {
		// Test project template is optional
		return nil
	}

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"TargetFramework": sch.Solution.TargetFramework,
		"ABPVersion":      sch.Solution.ABPVersion,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute test project template: %w", err)
	}

	testPath := g.getTestProjectPath(paths, sch)
	projectFile := filepath.Join(testPath, fmt.Sprintf("%s.%s.Tests.csproj", sch.Solution.Name, sch.Solution.ModuleName))
	return g.writer.WriteFile(projectFile, buf.String())
}
