package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
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
	// Skip if template is not available (test templates are optional)
	if err := g.generateTestBase(sch, paths); err != nil {
		// If template not found, skip test generation gracefully
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "file does not exist") {
			return nil // Skip test generation if templates don't exist
		}
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
	templateName := ""

	// Choose template based on target framework
	switch sch.Solution.TargetFramework {
	case schema.TargetASPNETCore9, schema.TargetASPNETCore10:
		templateName = "test_base_aspnetcore.tmpl"
	case schema.TargetABP8Microservice, schema.TargetABP9Microservice, schema.TargetABP10Microservice:
		templateName = "test_base_abp_microservice.tmpl"
	case schema.TargetABP8Monolith, schema.TargetABP9Monolith, schema.TargetABP10Monolith:
		// For monolith, try microservice template as fallback, then generic ABP template
		templateName = "test_base_abp_microservice.tmpl"
		tmpl, err = g.tmplLoader.Load(templateName)
		if err != nil {
			// Fallback to generic ABP template
			templateName = "test_base_abp.tmpl"
			tmpl, err = g.tmplLoader.Load(templateName)
		}
	default:
		// For auto or unknown, try microservice template first
		templateName = "test_base_abp_microservice.tmpl"
		tmpl, err = g.tmplLoader.Load(templateName)
		if err != nil {
			// Fallback to generic ABP template
			templateName = "test_base_abp.tmpl"
			tmpl, err = g.tmplLoader.Load(templateName)
		}
	}

	// Load template if not already loaded
	if tmpl == nil {
		tmpl, err = g.tmplLoader.Load(templateName)
	}

	if err != nil {
		return fmt.Errorf("failed to load test base template '%s': %w", templateName, err)
	}

	data := map[string]interface{}{
		"SolutionName":         sch.Solution.Name,
		"ModuleName":           sch.Solution.ModuleName,
		"ModuleNameWithSuffix": sch.Solution.GetModuleNameWithSuffix(),
		"NamespaceRoot":        sch.Solution.NamespaceRoot,
		"TargetFramework":      sch.Solution.TargetFramework,
		"DBProvider":           sch.Solution.DBProvider,
		"MultiTenancy":         sch.Solution.MultiTenancy,
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
		"SolutionName":         sch.Solution.Name,
		"ModuleName":           sch.Solution.ModuleName,
		"ModuleNameWithSuffix": sch.Solution.GetModuleNameWithSuffix(),
		"NamespaceRoot":        sch.Solution.NamespaceRoot,
		"EntityName":           entity.Name,
		"PrimaryKeyType":       primaryKeyType,
		"Properties":           entity.Properties,
		"TargetFramework":      sch.Solution.TargetFramework,
		"CustomRepository":     entity.CustomRepository,
		"Relations":            entity.Relations,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute repository test template: %w", err)
	}

	testPath := g.getTestProjectPath(paths, sch)
	moduleFolder := sch.Solution.GetModuleFolderName()
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
		"SolutionName":         sch.Solution.Name,
		"ModuleName":           sch.Solution.ModuleName,
		"ModuleNameWithSuffix": sch.Solution.GetModuleNameWithSuffix(),
		"NamespaceRoot":        sch.Solution.NamespaceRoot,
		"EntityName":           entity.Name,
		"PrimaryKeyType":       primaryKeyType,
		"Properties":           entity.Properties,
		"TargetFramework":      sch.Solution.TargetFramework,
		"Relations":            entity.Relations,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute service test template: %w", err)
	}

	testPath := g.getTestProjectPath(paths, sch)
	moduleFolder := sch.Solution.GetModuleFolderName()
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
		"SolutionName":         sch.Solution.Name,
		"ModuleName":           sch.Solution.ModuleName,
		"ModuleNameWithSuffix": sch.Solution.GetModuleNameWithSuffix(),
		"NamespaceRoot":        sch.Solution.NamespaceRoot,
		"EntityName":           entity.Name,
		"PrimaryKeyType":       primaryKeyType,
		"Properties":           entity.Properties,
		"TargetFramework":      sch.Solution.TargetFramework,
		"DomainEvents":         entity.DomainEvents,
		"Manager":              entity.EntityType == "AggregateRoot" || entity.EntityType == "FullAuditedAggregateRoot",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute domain test template: %w", err)
	}

	testPath := g.getTestProjectPath(paths, sch)
	moduleFolder := sch.Solution.GetModuleFolderName()
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
