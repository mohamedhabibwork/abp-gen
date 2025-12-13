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
	permissionsPath := paths.GetPermissionsFilePath(sch.Solution.ModuleName)

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

	// Update file idempotently
	return g.writer.UpdateFileIdempotent(permissionsPath, searchPattern, func(content string) (string, error) {
		// Find GetAll() method and insert before it
		getAllPattern := regexp.MustCompile(`(\s+)(public static string\[\] GetAll\(\))`)
		if !getAllPattern.MatchString(content) {
			return "", fmt.Errorf("GetAll() method not found in permissions file")
		}

		updated := getAllPattern.ReplaceAllString(content, newPermissions+"$1$2")
		return updated, nil
	})
}

// updatePermissionProvider updates the permission definition provider
func (g *PermissionsGenerator) updatePermissionProvider(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	providerPath := paths.GetPermissionProviderPath(sch.Solution.ModuleName)

	// Check if entity permissions already exist
	searchPattern := fmt.Sprintf("%sManagement.Default", entity.Name)

	// Template for new permission definitions
	tmpl, err := g.tmplLoader.Load("permission_provider.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load permission provider template: %w", err)
	}

	entityNameLower := strings.ToLower(entity.Name[:1]) + entity.Name[1:]

	data := map[string]interface{}{
		"ModuleName":      sch.Solution.ModuleName,
		"EntityName":      entity.Name,
		"EntityNameLower": entityNameLower,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute permission provider template: %w", err)
	}

	newDefinitions := buf.String()

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
	})
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
