package generator

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/mohamedhabibwork/abp-gen/internal/detector"
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
	"github.com/mohamedhabibwork/abp-gen/internal/writer"
)

// PermissionsGenerator generates permission constants and providers
type PermissionsGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewPermissionsGenerator creates a new permissions generator
func NewPermissionsGenerator(tmplLoader *templates.Loader, w *writer.Writer) *PermissionsGenerator {
	return &PermissionsGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates or updates permission files
func (g *PermissionsGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType == "ValueObject" {
		return nil // Value objects don't have permissions
	}

	// Update permissions file
	if err := g.updatePermissionsFile(sch, entity, paths); err != nil {
		return err
	}

	// Update permission provider
	return g.updatePermissionProvider(sch, entity, paths)
}

// updatePermissionsFile updates the permissions constants file
func (g *PermissionsGenerator) updatePermissionsFile(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	moduleFolder := sch.Solution.GetModuleFolderName()
	permissionsPath := paths.GetPermissionsFilePath(moduleFolder, sch.Solution.ModuleName)

	// Check if entity permissions already exist
	searchPattern := fmt.Sprintf("public static class %sManagement", entity.Name)

	// Template for new permissions
	tmpl, err := g.tmplLoader.Load("permissions.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load permissions template: %w", err)
	}

	data := map[string]interface{}{
		"EntityName": entity.Name,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute permissions template: %w", err)
	}

	newPermissions := buf.String()

	// Create initial file content if file doesn't exist
	createInitialContent := func() (string, error) {
		namespaceRoot := sch.Solution.NamespaceRoot
		moduleName := sch.Solution.ModuleName

		// Build initial permissions file structure
		moduleNamespace := sch.Solution.GetModuleNameWithSuffix()
		content := fmt.Sprintf(`using Volo.Abp.Authorization.Permissions;

namespace %s.Application.Contracts.Permissions.%s
{
    public static class %sPermissions
    {
        public const string GroupName = "%s";

%s
        public static string[] GetAll()
        {
            return new[]
            {
                %sManagement.Default,
                %sManagement.Create,
                %sManagement.Update,
                %sManagement.Delete
            };
        }
    }
}
`, namespaceRoot, moduleNamespace, moduleName, moduleName, newPermissions, entity.Name, entity.Name, entity.Name, entity.Name)
		return content, nil
	}

	// Update file idempotently
	return g.writer.UpdateFileIdempotent(permissionsPath, searchPattern, func(content string) (string, error) {
		// Find GetAll() method and insert before it
		getAllPattern := regexp.MustCompile(`(\s+)(public static string\[\] GetAll\(\))`)
		if !getAllPattern.MatchString(content) {
			return "", fmt.Errorf("GetAll() method not found in permissions file")
		}

		updated := getAllPattern.ReplaceAllString(content, newPermissions+"$1$2")
		return updated, nil
	}, createInitialContent)
}

// updatePermissionProvider updates the permission definition provider
func (g *PermissionsGenerator) updatePermissionProvider(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	moduleFolder := sch.Solution.GetModuleFolderName()
	providerPath := paths.GetPermissionProviderPath(moduleFolder, sch.Solution.ModuleName)

	// Check if entity permissions already exist
	searchPattern := fmt.Sprintf("%sManagement.Default", entity.Name)

	// Template for new permission definitions
	tmpl, err := g.tmplLoader.Load("permission_provider.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load permission provider template: %w", err)
	}

	entityNameLower := strings.ToLower(entity.Name[:1]) + entity.Name[1:]
	moduleNameLower := strings.ToLower(sch.Solution.ModuleName[:1]) + sch.Solution.ModuleName[1:]

	data := map[string]interface{}{
		"ModuleName":      sch.Solution.ModuleName,
		"ModuleNameLower": moduleNameLower,
		"EntityName":      entity.Name,
		"EntityNameLower": entityNameLower,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute permission provider template: %w", err)
	}

	newDefinitions := buf.String()

	// Create initial file content if file doesn't exist
	createInitialContent := func() (string, error) {
		namespaceRoot := sch.Solution.NamespaceRoot
		moduleName := sch.Solution.ModuleName

		moduleNameLower := strings.ToLower(moduleName[:1]) + moduleName[1:]

		// Build initial permission provider file structure
		moduleNamespace := sch.Solution.GetModuleNameWithSuffix()
		content := fmt.Sprintf(`using Volo.Abp.Authorization.Permissions;
using Volo.Abp.Localization;
using %s.Localization.%sModule;

namespace %s.Application.Contracts.Permissions.%s
{
    public class %sPermissionDefinitionProvider : PermissionDefinitionProvider
    {
        public override void Define(IPermissionDefinitionContext context)
        {
            var %sGroup = context.GetGroupOrNull(%sPermissions.GroupName)
                ?? context.AddGroup(%sPermissions.GroupName, L("Permission:%s"));

%s        }
    }
}
`, namespaceRoot, moduleName, namespaceRoot, moduleNamespace, moduleName, moduleNameLower, moduleName, moduleName, moduleName, newDefinitions)
		return content, nil
	}

	// Update file idempotently
	return g.writer.UpdateFileIdempotent(providerPath, searchPattern, func(content string) (string, error) {
		// Find the closing braces of the Define method and insert before them
		// Pattern: } (end of method) } (end of class) } (end of namespace)
		pattern := regexp.MustCompile(`(\s+)(}\s+}\s+}\s*$)`)
		if !pattern.MatchString(content) {
			return "", fmt.Errorf("could not find insertion point in permission provider")
		}

		updated := pattern.ReplaceAllString(content, newDefinitions+"$1$2")
		return updated, nil
	}, createInitialContent)
}

// GenerateLocalization generates or updates localization files
func (g *PermissionsGenerator) GenerateLocalization(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if !sch.Options.UseLocalization {
		return nil
	}

	for _, culture := range sch.Options.LocalizationCultures {
		if err := g.updateLocalizationFile(sch, entity, paths, culture); err != nil {
			return fmt.Errorf("failed to update %s localization: %w", culture, err)
		}
	}

	return nil
}

// updateLocalizationFile updates a localization JSON file
func (g *PermissionsGenerator) updateLocalizationFile(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths, culture string) error {
	// This is a simplified version - in production, you'd want to properly parse and merge JSON
	// For now, we'll just note that the file should be updated manually or with a JSON library

	// The localization file updates would require proper JSON parsing
	// which we'll handle in the template or with a JSON library

	// Placeholder for now - in full implementation, would:
	// 1. Read existing JSON
	// 2. Add new keys if not present
	// 3. Sort keys
	// 4. Write back

	return nil
}
