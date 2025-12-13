package schema

import (
	"fmt"
	"strings"
)

// Validate validates the schema and returns any errors
func (s *Schema) Validate() error {
	if err := s.validateSolution(); err != nil {
		return err
	}

	if len(s.Entities) == 0 {
		return fmt.Errorf("schema must contain at least one entity")
	}

	entityNames := make(map[string]bool)
	for i, entity := range s.Entities {
		if err := s.validateEntity(&entity, entityNames); err != nil {
			return fmt.Errorf("entity[%d] '%s': %w", i, entity.Name, err)
		}
		entityNames[entity.Name] = true
	}

	// Validate relations reference existing entities
	for _, entity := range s.Entities {
		if err := s.validateRelations(&entity, entityNames); err != nil {
			return fmt.Errorf("entity '%s' relations: %w", entity.Name, err)
		}
	}

	return nil
}

func (s *Schema) validateSolution() error {
	if s.Solution.Name == "" {
		return fmt.Errorf("solution.name is required")
	}

	if s.Solution.ModuleName == "" {
		return fmt.Errorf("solution.moduleName is required")
	}

	if s.Solution.NamespaceRoot == "" {
		s.Solution.NamespaceRoot = fmt.Sprintf("%s.%s", s.Solution.Name, s.Solution.ModuleName)
	}

	if s.Solution.ABPVersion == "" {
		s.Solution.ABPVersion = "9.0"
	}

	if s.Solution.PrimaryKeyType == "" {
		s.Solution.PrimaryKeyType = "Guid"
	}

	validPKTypes := map[string]bool{"Guid": true, "long": true, "configurable": true}
	if !validPKTypes[s.Solution.PrimaryKeyType] {
		return fmt.Errorf("solution.primaryKeyType must be 'Guid', 'long', or 'configurable', got '%s'", s.Solution.PrimaryKeyType)
	}

	if s.Solution.DBProvider == "" {
		s.Solution.DBProvider = "efcore"
	}

	validProviders := map[string]bool{"efcore": true, "mongodb": true, "both": true}
	if !validProviders[s.Solution.DBProvider] {
		return fmt.Errorf("solution.dbProvider must be 'efcore', 'mongodb', or 'both', got '%s'", s.Solution.DBProvider)
	}

	// Set default validation type
	if s.Options.ValidationType == "" {
		s.Options.ValidationType = "fluentvalidation"
	}

	validValidationTypes := map[string]bool{"fluentvalidation": true, "native": true}
	if !validValidationTypes[s.Options.ValidationType] {
		return fmt.Errorf("options.validationType must be 'fluentvalidation' or 'native', got '%s'", s.Options.ValidationType)
	}

	return nil
}

func (s *Schema) validateEntity(entity *Entity, existingNames map[string]bool) error {
	if entity.Name == "" {
		return fmt.Errorf("entity name is required")
	}

	if existingNames[entity.Name] {
		return fmt.Errorf("duplicate entity name '%s'", entity.Name)
	}

	if entity.TableName == "" {
		entity.TableName = Pluralize(entity.Name)
	}

	if entity.EntityType == "" {
		entity.EntityType = "FullAuditedAggregateRoot"
	}

	validTypes := map[string]bool{
		"Entity":                      true,
		"AggregateRoot":               true,
		"FullAuditedAggregateRoot":    true,
		"AuditedAggregateRoot":        true,
		"ValueObject":                 true,
	}
	if !validTypes[entity.EntityType] {
		return fmt.Errorf("invalid entityType '%s'", entity.EntityType)
	}

	if len(entity.Properties) == 0 && entity.EntityType != "ValueObject" {
		return fmt.Errorf("entity must have at least one property")
	}

	propertyNames := make(map[string]bool)
	for i, prop := range entity.Properties {
		if err := s.validateProperty(&prop, propertyNames); err != nil {
			return fmt.Errorf("property[%d] '%s': %w", i, prop.Name, err)
		}
		propertyNames[prop.Name] = true
	}

	return nil
}

func (s *Schema) validateProperty(prop *Property, existingNames map[string]bool) error {
	if prop.Name == "" {
		return fmt.Errorf("property name is required")
	}

	if existingNames[prop.Name] {
		return fmt.Errorf("duplicate property name '%s'", prop.Name)
	}

	if prop.Type == "" {
		return fmt.Errorf("property type is required")
	}

	validTypes := map[string]bool{
		"string": true, "int": true, "long": true, "decimal": true,
		"DateTime": true, "bool": true, "Guid": true, "byte": true,
		"short": true, "float": true, "double": true,
	}

	// Allow custom types (might be enums or other entities)
	if !validTypes[prop.Type] && !strings.Contains(prop.Type, ".") {
		// Could be a custom type, we'll allow it but add a warning in verbose mode
	}

	if prop.IsForeignKey && prop.TargetEntity == "" {
		return fmt.Errorf("foreign key property must specify targetEntity")
	}

	return nil
}

func (s *Schema) validateRelations(entity *Entity, entityNames map[string]bool) error {
	if entity.Relations == nil {
		return nil
	}

	for i, rel := range entity.Relations.OneToMany {
		if rel.TargetEntity == "" {
			return fmt.Errorf("oneToMany[%d]: targetEntity is required", i)
		}
		// Note: Target entity might not exist yet (forward reference) - this is OK for generation
	}

	for i, rel := range entity.Relations.ManyToMany {
		if rel.TargetEntity == "" {
			return fmt.Errorf("manyToMany[%d]: targetEntity is required", i)
		}
		if rel.JoinEntity == "" {
			// Auto-generate join entity name
			entities := []string{entity.Name, rel.TargetEntity}
			// Sort to ensure consistent naming
			if entities[0] > entities[1] {
				entities[0], entities[1] = entities[1], entities[0]
			}
			rel.JoinEntity = entities[0] + entities[1]
		}
	}

	return nil
}

// Pluralize converts a singular word to plural (simple implementation)
func Pluralize(word string) string {
	// This is a simple implementation - the actual pluralization will be handled by go-pluralize library
	if strings.HasSuffix(word, "y") && len(word) > 1 {
		return word[:len(word)-1] + "ies"
	}
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "x") ||
		strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "sh") {
		return word + "es"
	}
	return word + "s"
}

