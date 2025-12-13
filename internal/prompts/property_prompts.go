package prompts

import (
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
)

// PromptProperty prompts for a single property
func PromptProperty() (*schema.Property, error) {
	name, err := PromptText("Property name:", "")
	if err != nil {
		return nil, err
	}

	propertyType, err := PromptSelect(
		"Property type:",
		[]string{
			"string",
			"int",
			"long",
			"decimal",
			"DateTime",
			"bool",
			"Guid",
			"byte",
			"short",
			"float",
			"double",
			"Custom",
		},
		"string",
	)
	if err != nil {
		return nil, err
	}

	// If custom type, prompt for type name
	if propertyType == "Custom" {
		propertyType, err = PromptText("Custom type name:", "")
		if err != nil {
			return nil, err
		}
	}

	isRequired, err := PromptConfirm("Is required?", true)
	if err != nil {
		return nil, err
	}

	nullable, err := PromptConfirm("Is nullable?", false)
	if err != nil {
		return nil, err
	}

	var maxLength int
	if propertyType == "string" {
		maxLength, err = PromptInt("Max length (0 for no limit):", 0)
		if err != nil {
			return nil, err
		}
	}

	var defaultValue string
	hasDefault, err := PromptConfirm("Has default value?", false)
	if err != nil {
		return nil, err
	}
	if hasDefault {
		defaultValue, err = PromptText("Default value:", "")
		if err != nil {
			return nil, err
		}
	}

	// Check if it's a foreign key
	isForeignKey, err := PromptConfirm("Is this a foreign key?", false)
	if err != nil {
		return nil, err
	}

	var targetEntity string
	if isForeignKey {
		targetEntity, err = PromptText("Target entity name:", "")
		if err != nil {
			return nil, err
		}
	}

	property := &schema.Property{
		Name:         name,
		Type:         propertyType,
		IsRequired:   isRequired,
		MaxLength:    maxLength,
		Nullable:     nullable,
		DefaultValue: defaultValue,
		IsForeignKey: isForeignKey,
		TargetEntity: targetEntity,
	}

	return property, nil
}

