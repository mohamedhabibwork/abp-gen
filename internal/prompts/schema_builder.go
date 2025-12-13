package prompts

import (
	"fmt"

	"github.com/mohamedhabibwork/abp-gen/internal/schema"
)

// BuildSchemaInteractively builds a complete schema through interactive prompts
func BuildSchemaInteractively() (*schema.Schema, error) {
	fmt.Println("\n=== ABP Code Generator - Interactive Mode ===")
	fmt.Println()

	// Build solution configuration
	solution, err := PromptSolutionConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get solution config: %w", err)
	}

	// Build entities
	entities, err := PromptEntities(solution.PrimaryKeyType)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	// Build generation options
	options, err := PromptGenerationOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to get generation options: %w", err)
	}

	sch := &schema.Schema{
		Solution: *solution,
		Entities: entities,
		Options:  *options,
	}

	// Validate schema
	if err := sch.Validate(); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	// Display summary
	DisplaySchemaSummary(sch)

	// Ask if user wants to save
	save, err := PromptConfirm("Save schema to file?", true)
	if err != nil {
		return nil, err
	}

	if save {
		path, err := PromptText("Schema file path", "schema.json")
		if err != nil {
			return nil, err
		}

		if err := sch.SaveToFile(path); err != nil {
			return nil, fmt.Errorf("failed to save schema: %w", err)
		}
		fmt.Printf("✓ Schema saved to %s\n", path)
	}

	return sch, nil
}

// PromptSolutionConfig prompts for solution configuration
func PromptSolutionConfig() (*schema.Solution, error) {
	fmt.Println("=== Solution Configuration ===")
	fmt.Println()

	name, err := PromptText("Solution name:", "")
	if err != nil {
		return nil, err
	}

	moduleName, err := PromptText("Module/Service name:", "")
	if err != nil {
		return nil, err
	}

	defaultNamespace := name
	namespaceRoot, err := PromptText("Namespace root:", defaultNamespace)
	if err != nil {
		return nil, err
	}

	abpVersion, err := PromptText("ABP version:", "9.0")
	if err != nil {
		return nil, err
	}

	primaryKeyType, err := PromptSelect(
		"Primary key type:",
		[]string{"Guid", "long", "configurable"},
		"Guid",
	)
	if err != nil {
		return nil, err
	}

	dbProvider, err := PromptSelect(
		"Database provider:",
		[]string{"efcore", "mongodb", "both"},
		"efcore",
	)
	if err != nil {
		return nil, err
	}

	generateControllers, err := PromptConfirm("Generate HTTP API Controllers?", true)
	if err != nil {
		return nil, err
	}

	return &schema.Solution{
		Name:                name,
		ModuleName:          moduleName,
		NamespaceRoot:       namespaceRoot,
		ABPVersion:          abpVersion,
		PrimaryKeyType:      primaryKeyType,
		DBProvider:          dbProvider,
		GenerateControllers: generateControllers,
	}, nil
}

// PromptEntities prompts for all entities
func PromptEntities(defaultPrimaryKeyType string) ([]schema.Entity, error) {
	var entities []schema.Entity
	entityCount := 1

	for {
		fmt.Printf("\n=== Entity %d ===\n", entityCount)

		entity, err := PromptEntity(defaultPrimaryKeyType)
		if err != nil {
			return nil, err
		}

		entities = append(entities, *entity)
		entityCount++

		addMore, err := PromptConfirm("Add another entity?", false)
		if err != nil {
			return nil, err
		}

		if !addMore {
			break
		}
	}

	return entities, nil
}

// PromptGenerationOptions prompts for generation options
func PromptGenerationOptions() (*schema.Options, error) {
	fmt.Println("\n=== Generation Options ===")
	fmt.Println()

	useAuditedAggregateRoot, err := PromptConfirm("Use audited aggregate root?", true)
	if err != nil {
		return nil, err
	}

	useSoftDelete, err := PromptConfirm("Use soft delete?", true)
	if err != nil {
		return nil, err
	}

	useConcurrencyStamp, err := PromptConfirm("Use concurrency stamp?", true)
	if err != nil {
		return nil, err
	}

	useExtraProperties, err := PromptConfirm("Use extra properties?", true)
	if err != nil {
		return nil, err
	}

	useLocalization, err := PromptConfirm("Use localization?", true)
	if err != nil {
		return nil, err
	}

	var localizationCultures []string
	if useLocalization {
		cultures, err := PromptMultiSelect(
			"Select localization cultures:",
			[]string{"en", "ar", "fr", "de", "es", "it", "zh", "ja"},
			[]string{"en", "ar"},
		)
		if err != nil {
			return nil, err
		}
		localizationCultures = cultures
	}

	validationType, err := PromptSelect(
		"Validation type:",
		[]string{"fluentvalidation", "native"},
		"fluentvalidation",
	)
	if err != nil {
		return nil, err
	}

	generateEventHandlers, err := PromptConfirm("Generate event handlers?", true)
	if err != nil {
		return nil, err
	}

	return &schema.Options{
		UseAuditedAggregateRoot: useAuditedAggregateRoot,
		UseSoftDelete:           useSoftDelete,
		UseConcurrencyStamp:     useConcurrencyStamp,
		UseExtraProperties:      useExtraProperties,
		UseLocalization:         useLocalization,
		LocalizationCultures:    localizationCultures,
		ValidationType:          validationType,
		GenerateEventHandlers:   generateEventHandlers,
	}, nil
}

// DisplaySchemaSummary displays a summary of the schema
func DisplaySchemaSummary(sch *schema.Schema) {
	fmt.Println("\n=== Schema Summary ===")
	fmt.Printf("Solution: %s\n", sch.Solution.Name)
	fmt.Printf("Module: %s\n", sch.Solution.ModuleName)
	fmt.Printf("Namespace: %s\n", sch.Solution.NamespaceRoot)
	fmt.Printf("ABP Version: %s\n", sch.Solution.ABPVersion)
	fmt.Printf("Primary Key Type: %s\n", sch.Solution.PrimaryKeyType)
	fmt.Printf("Database Provider: %s\n", sch.Solution.DBProvider)
	fmt.Printf("Generate Controllers: %v\n\n", sch.Solution.GenerateControllers)

	fmt.Printf("Entities (%d):\n", len(sch.Entities))
	for i, entity := range sch.Entities {
		fmt.Printf("  %d. %s (%s)\n", i+1, entity.Name, entity.EntityType)
		fmt.Printf("     Properties: %d\n", len(entity.Properties))
		if entity.HasRelations() {
			fmt.Printf("     Relations: One-to-Many(%d), Many-to-Many(%d)\n",
				len(entity.Relations.OneToMany),
				len(entity.Relations.ManyToMany))
		}
	}

	fmt.Println("\nOptions:")
	fmt.Printf("  Use Localization: %v", sch.Options.UseLocalization)
	if sch.Options.UseLocalization {
		fmt.Printf(" (%v)", sch.Options.LocalizationCultures)
	}
	fmt.Println()
	fmt.Printf("  Validation Type: %s\n", sch.Options.ValidationType)
	fmt.Printf("  Generate Event Handlers: %v\n", sch.Options.GenerateEventHandlers)
}
