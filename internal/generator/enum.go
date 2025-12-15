package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mohamedhabibwork/abp-gen/internal/detector"
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
	"github.com/mohamedhabibwork/abp-gen/internal/writer"
)

// EnumGenerator generates enum types and related files
type EnumGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewEnumGenerator creates a new enum generator
func NewEnumGenerator(tmplLoader *templates.Loader, w *writer.Writer) *EnumGenerator {
	return &EnumGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates enum definitions for an entity
func (g *EnumGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if len(entity.Enums) == 0 {
		return nil
	}

	for _, enum := range entity.Enums {
		// Generate enum class
		if err := g.generateEnumClass(sch, entity, &enum, paths); err != nil {
			return fmt.Errorf("failed to generate enum %s: %w", enum.Name, err)
		}

		// Generate lookup extensions if requested
		if enum.GenerateLookup {
			if err := g.generateEnumLookup(sch, entity, &enum, paths); err != nil {
				return fmt.Errorf("failed to generate lookup for enum %s: %w", enum.Name, err)
			}
		}

		// Generate localization entries if requested
		if enum.UseLocalization {
			if err := g.generateEnumLocalization(sch, entity, &enum, paths); err != nil {
				return fmt.Errorf("failed to generate localization for enum %s: %w", enum.Name, err)
			}
		}

		// Generate DTO representation
		if err := g.generateEnumDTO(sch, entity, &enum, paths); err != nil {
			return fmt.Errorf("failed to generate DTO for enum %s: %w", enum.Name, err)
		}
	}

	return nil
}

func (g *EnumGenerator) generateEnumClass(sch *schema.Schema, entity *schema.Entity, enum *schema.EnumDefinition, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("enum.tmpl")
	if err != nil {
		// If template not found, skip enum generation gracefully
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "file does not exist") {
			return nil // Skip enum generation if template doesn't exist
		}
		return fmt.Errorf("failed to load enum template: %w", err)
	}

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"EntityName":      entity.Name,
		"EnumName":        enum.Name,
		"UnderlyingType":  enum.UnderlyingType,
		"Values":          enum.Values,
		"Description":     enum.Description,
		"TargetFramework": sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute enum template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	enumPath := filepath.Join(paths.DomainSharedEnums, moduleFolder, enum.Name+".cs")
	return g.writer.WriteFile(enumPath, buf.String())
}

func (g *EnumGenerator) generateEnumLookup(sch *schema.Schema, entity *schema.Entity, enum *schema.EnumDefinition, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("enum_lookup.tmpl")
	if err != nil {
		// If template not found, skip lookup generation gracefully
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "file does not exist") {
			return nil // Skip lookup generation if template doesn't exist
		}
		return fmt.Errorf("failed to load enum_lookup template: %w", err)
	}

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"EntityName":      entity.Name,
		"EnumName":        enum.Name,
		"Values":          enum.Values,
		"TargetFramework": sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute enum_lookup template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	lookupPath := filepath.Join(paths.DomainSharedEnums, moduleFolder, enum.Name+"Extensions.cs")
	return g.writer.WriteFile(lookupPath, buf.String())
}

func (g *EnumGenerator) generateEnumLocalization(sch *schema.Schema, entity *schema.Entity, enum *schema.EnumDefinition, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("enum_localization.tmpl")
	if err != nil {
		// If template not found, skip localization generation gracefully
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "file does not exist") {
			return nil // Skip localization generation if template doesn't exist
		}
		return fmt.Errorf("failed to load enum_localization template: %w", err)
	}

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"EntityName":      entity.Name,
		"EnumName":        enum.Name,
		"Values":          enum.Values,
		"Cultures":        sch.Options.LocalizationCultures,
		"TargetFramework": sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute enum_localization template: %w", err)
	}

	// Generate JSON content for localization files
	moduleFolder := sch.Solution.ModuleName + "Module"
	locPath := filepath.Join(paths.DomainSharedLocalization, moduleFolder, enum.Name+"_enums.json")
	return g.writer.WriteFile(locPath, buf.String())
}

func (g *EnumGenerator) generateEnumDTO(sch *schema.Schema, entity *schema.Entity, enum *schema.EnumDefinition, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("enum_dto.tmpl")
	if err != nil {
		// If template not found, skip DTO generation gracefully
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "file does not exist") {
			return nil // Skip DTO generation if template doesn't exist
		}
		return fmt.Errorf("failed to load enum_dto template: %w", err)
	}

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"EntityName":      entity.Name,
		"EnumName":        enum.Name,
		"UnderlyingType":  enum.UnderlyingType,
		"Values":          enum.Values,
		"TargetFramework": sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute enum_dto template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	dtoPath := filepath.Join(paths.ContractsDTOs, moduleFolder, entity.Name, enum.Name+"Dto.cs")
	return g.writer.WriteFile(dtoPath, buf.String())
}
