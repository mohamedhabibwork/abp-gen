package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mohamedhabibwork/abp-gen/internal/detector"
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
	"github.com/mohamedhabibwork/abp-gen/internal/writer"
)

// DTOGenerator generates Data Transfer Objects
type DTOGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewDTOGenerator creates a new DTO generator
func NewDTOGenerator(tmplLoader *templates.Loader, w *writer.Writer) *DTOGenerator {
	return &DTOGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates all DTO files for an entity
func (g *DTOGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType == "ValueObject" {
		// Value objects only get a read DTO
		return g.GenerateEntityDto(sch, entity, paths)
	}

	// Ensure DTO directory exists
	moduleFolder := sch.Solution.GetModuleFolderName()
	dtoPath := paths.GetEntityDTOPath(moduleFolder, entity.Name)
	if err := os.MkdirAll(dtoPath, 0755); err != nil {
		return fmt.Errorf("failed to create DTO directory: %w", err)
	}

	// Generate Create DTO
	if err := g.GenerateCreateDto(sch, entity, paths); err != nil {
		return err
	}

	// Generate Update DTO
	if err := g.GenerateUpdateDto(sch, entity, paths); err != nil {
		return err
	}

	// Generate Entity DTO (read)
	if err := g.GenerateEntityDto(sch, entity, paths); err != nil {
		return err
	}

	return nil
}

// GenerateCreateDto generates the Create DTO
func (g *DTOGenerator) GenerateCreateDto(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("create_dto.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load create DTO template: %w", err)
	}

	data := g.prepareDtoData(sch, entity)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute create DTO template: %w", err)
	}

	moduleFolder := sch.Solution.GetModuleFolderName()
	dtoPath := paths.GetEntityDTOPath(moduleFolder, entity.Name)
	filePath := filepath.Join(dtoPath, "Create"+entity.Name+"Dto.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// GenerateUpdateDto generates the Update DTO
func (g *DTOGenerator) GenerateUpdateDto(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("update_dto.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load update DTO template: %w", err)
	}

	data := g.prepareDtoData(sch, entity)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute update DTO template: %w", err)
	}

	moduleFolder := sch.Solution.GetModuleFolderName()
	dtoPath := paths.GetEntityDTOPath(moduleFolder, entity.Name)
	filePath := filepath.Join(dtoPath, "Update"+entity.Name+"Dto.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// GenerateEntityDto generates the Entity DTO (read)
func (g *DTOGenerator) GenerateEntityDto(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("entity_dto.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load entity DTO template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":            sch.Solution.Name,
		"ModuleName":              sch.Solution.ModuleName,
		"ModuleNameWithSuffix":    sch.Solution.GetModuleNameWithSuffix(),
		"NamespaceRoot":           sch.Solution.NamespaceRoot,
		"EntityName":              entity.Name,
		"PrimaryKeyType":          primaryKeyType,
		"Properties":              entity.Properties,
		"NonForeignKeyProperties": entity.GetNonForeignKeyProperties(),
		"ForeignKeyProperties":    entity.GetForeignKeyProperties(),
		"HasRelations":            entity.HasRelations(),
		"OneToManyRelations":      getOneToManyRelations(entity),
		"ManyToManyRelations":     getManyToManyRelations(entity),
		"HasEnumProperties":       entity.HasEnumProperties(),
		"EnumNames":               entity.GetEnumNames(),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute entity DTO template: %w", err)
	}

	moduleFolder := sch.Solution.GetModuleFolderName()
	dtoPath := paths.GetEntityDTOPath(moduleFolder, entity.Name)
	filePath := filepath.Join(dtoPath, entity.Name+"Dto.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// prepareDtoData prepares common data for DTO templates
func (g *DTOGenerator) prepareDtoData(sch *schema.Schema, entity *schema.Entity) map[string]interface{} {
	return map[string]interface{}{
		"SolutionName":            sch.Solution.Name,
		"ModuleName":              sch.Solution.ModuleName,
		"ModuleNameWithSuffix":    sch.Solution.GetModuleNameWithSuffix(),
		"NamespaceRoot":           sch.Solution.NamespaceRoot,
		"EntityName":              entity.Name,
		"Properties":              entity.Properties,
		"NonForeignKeyProperties": entity.GetNonForeignKeyProperties(),
		"ForeignKeyProperties":    entity.GetForeignKeyProperties(),
		"HasEnumProperties":       entity.HasEnumProperties(),
		"EnumNames":               entity.GetEnumNames(),
		"NeedsDataAnnotations":    entity.NeedsDataAnnotations(),
	}
}

// GenerateAppServiceInterface generates the application service interface
func (g *DTOGenerator) GenerateAppServiceInterface(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if entity.EntityType == "ValueObject" {
		return nil
	}

	tmpl, err := g.tmplLoader.Load("app_service_interface.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load app service interface template: %w", err)
	}

	primaryKeyType := entity.GetEffectivePrimaryKeyType(sch.Solution.PrimaryKeyType)

	data := map[string]interface{}{
		"SolutionName":         sch.Solution.Name,
		"ModuleName":           sch.Solution.ModuleName,
		"ModuleNameWithSuffix": sch.Solution.GetModuleNameWithSuffix(),
		"NamespaceRoot":        sch.Solution.NamespaceRoot,
		"EntityName":           entity.Name,
		"PrimaryKeyType":       primaryKeyType,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute app service interface template: %w", err)
	}

	moduleFolder := sch.Solution.GetModuleFolderName()
	filePath := filepath.Join(paths.ContractsServices, moduleFolder, "I"+entity.Name+"AppService.cs")
	return g.writer.WriteFile(filePath, buf.String())
}
