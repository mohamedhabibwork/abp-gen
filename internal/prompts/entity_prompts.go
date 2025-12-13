package prompts

import (
	"fmt"

	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
)

// PromptEntity prompts for entity information
func PromptEntity(defaultPrimaryKeyType string) (*schema.Entity, error) {
	name, err := PromptText("Entity name:", "")
	if err != nil {
		return nil, err
	}

	defaultTableName := templates.Pluralize(name)
	tableName, err := PromptText("Table name:", defaultTableName)
	if err != nil {
		return nil, err
	}

	entityType, err := PromptSelect(
		"Entity type:",
		[]string{
			"FullAuditedAggregateRoot",
			"AggregateRoot",
			"AuditedAggregateRoot",
			"Entity",
			"ValueObject",
		},
		"FullAuditedAggregateRoot",
	)
	if err != nil {
		return nil, err
	}

	var primaryKeyType string
	if defaultPrimaryKeyType == "configurable" {
		primaryKeyType, err = PromptSelect(
			"Primary key type:",
			[]string{"Guid", "long"},
			"Guid",
		)
		if err != nil {
			return nil, err
		}
	}

	// Prompt for properties
	fmt.Printf("\n=== Properties for %s ===\n", name)
	properties, err := PromptProperties()
	if err != nil {
		return nil, err
	}

	// Prompt for relations
	fmt.Printf("\n=== Relations for %s ===\n", name)
	relations, err := PromptRelations()
	if err != nil {
		return nil, err
	}

	entity := &schema.Entity{
		Name:           name,
		TableName:      tableName,
		EntityType:     entityType,
		PrimaryKeyType: primaryKeyType,
		Properties:     properties,
		Relations:      relations,
	}

	return entity, nil
}

// PromptProperties prompts for all properties
func PromptProperties() ([]schema.Property, error) {
	var properties []schema.Property

	for {
		addProperty, err := PromptConfirm("Add property?", len(properties) == 0)
		if err != nil {
			return nil, err
		}

		if !addProperty {
			break
		}

		property, err := PromptProperty()
		if err != nil {
			return nil, err
		}

		properties = append(properties, *property)
	}

	return properties, nil
}

