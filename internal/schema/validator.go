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
		s.Solution.NamespaceRoot = s.Solution.Name
	} else {
		// Normalize: if NamespaceRoot contains the module name, strip it
		// e.g., "EdeServices.User" -> "EdeServices"
		expectedWithModule := s.Solution.Name + "." + s.Solution.ModuleName
		if s.Solution.NamespaceRoot == expectedWithModule {
			s.Solution.NamespaceRoot = s.Solution.Name
		}
	}

	if s.Solution.ABPVersion == "" {
		s.Solution.ABPVersion = "9.0"
	}

	// Validate target framework
	if s.Solution.TargetFramework == "" {
		s.Solution.TargetFramework = TargetAuto
	}
	validTargets := map[TargetFramework]bool{
		TargetASPNETCore9:       true,
		TargetASPNETCore10:      true,
		TargetABP8Microservice:  true,
		TargetABP8Monolith:      true,
		TargetABP9Microservice:  true,
		TargetABP9Monolith:      true,
		TargetABP10Microservice: true,
		TargetABP10Monolith:     true,
		TargetAuto:              true,
	}
	if !validTargets[s.Solution.TargetFramework] {
		return fmt.Errorf("solution.targetFramework must be one of: aspnetcore9, aspnetcore10, abp8-monolith, abp8-microservice, abp9-monolith, abp9-microservice, abp10-monolith, abp10-microservice, or auto, got '%s'", s.Solution.TargetFramework)
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

	// Validate multi-tenancy configuration
	if s.Solution.MultiTenancy != nil {
		if err := s.validateMultiTenancy(s.Solution.MultiTenancy); err != nil {
			return fmt.Errorf("solution.multiTenancy: %w", err)
		}
	}

	// Set default validation type
	if s.Options.ValidationType == "" {
		s.Options.ValidationType = "fluentvalidation"
	}

	validValidationTypes := map[string]bool{"fluentvalidation": true, "native": true}
	if !validValidationTypes[s.Options.ValidationType] {
		return fmt.Errorf("options.validationType must be 'fluentvalidation' or 'native', got '%s'", s.Options.ValidationType)
	}

	// Validate localization merge configuration
	if s.Options.LocalizationMerge != nil {
		validStrategies := map[string]bool{"overwrite": true, "append": true, "skip": true}
		if s.Options.LocalizationMerge.ConflictStrategy != "" && !validStrategies[s.Options.LocalizationMerge.ConflictStrategy] {
			return fmt.Errorf("options.localizationMerge.conflictStrategy must be 'overwrite', 'append', or 'skip', got '%s'", s.Options.LocalizationMerge.ConflictStrategy)
		}
	}

	return nil
}

func (s *Schema) validateMultiTenancy(mt *MultiTenancy) error {
	validStrategies := map[string]bool{"none": true, "host": true, "tenant-per-db": true, "tenant-per-schema": true}
	if mt.Strategy != "" && !validStrategies[mt.Strategy] {
		return fmt.Errorf("strategy must be 'none', 'host', 'tenant-per-db', or 'tenant-per-schema', got '%s'", mt.Strategy)
	}
	if mt.TenantIdProperty == "" {
		mt.TenantIdProperty = "TenantId"
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
		"Entity":                   true,
		"AggregateRoot":            true,
		"FullAuditedAggregateRoot": true,
		"AuditedAggregateRoot":     true,
		"ValueObject":              true,
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

	// Validate custom repository
	if entity.CustomRepository != nil {
		for i, method := range entity.CustomRepository.Methods {
			if err := s.validateRepositoryMethod(&method); err != nil {
				return fmt.Errorf("customRepository.methods[%d] '%s': %w", i, method.Name, err)
			}
		}
	}

	// Validate domain events
	for i, event := range entity.DomainEvents {
		if err := s.validateDomainEvent(&event); err != nil {
			return fmt.Errorf("domainEvents[%d] '%s': %w", i, event.Name, err)
		}
	}

	// Validate enums
	enumNames := make(map[string]bool)
	for i, enum := range entity.Enums {
		if err := s.validateEnum(&enum, enumNames); err != nil {
			return fmt.Errorf("enums[%d] '%s': %w", i, enum.Name, err)
		}
		enumNames[enum.Name] = true
	}

	// Validate value object configuration
	if entity.ValueObjectConfig != nil && entity.EntityType == "ValueObject" {
		if err := s.validateValueObjectConfig(entity.ValueObjectConfig, entity.Properties); err != nil {
			return fmt.Errorf("valueObjectConfig: %w", err)
		}
	}

	return nil
}

func (s *Schema) validateRepositoryMethod(method *RepositoryMethod) error {
	if method.Name == "" {
		return fmt.Errorf("method name is required")
	}
	if method.ReturnType == "" {
		return fmt.Errorf("return type is required")
	}
	return nil
}

func (s *Schema) validateDomainEvent(event *DomainEvent) error {
	if event.Name == "" {
		return fmt.Errorf("event name is required")
	}
	if event.Type == "" {
		event.Type = "domain"
	}
	validTypes := map[string]bool{"domain": true, "distributed": true}
	if !validTypes[event.Type] {
		return fmt.Errorf("event type must be 'domain' or 'distributed', got '%s'", event.Type)
	}
	return nil
}

func (s *Schema) validateEnum(enum *EnumDefinition, existingNames map[string]bool) error {
	if enum.Name == "" {
		return fmt.Errorf("enum name is required")
	}
	if existingNames[enum.Name] {
		return fmt.Errorf("duplicate enum name '%s'", enum.Name)
	}
	if enum.UnderlyingType == "" {
		enum.UnderlyingType = "int"
	}
	if len(enum.Values) == 0 {
		return fmt.Errorf("enum must have at least one value")
	}
	valueNames := make(map[string]bool)
	for i, val := range enum.Values {
		if val.Name == "" {
			return fmt.Errorf("enum value[%d] name is required", i)
		}
		if valueNames[val.Name] {
			return fmt.Errorf("duplicate enum value name '%s'", val.Name)
		}
		valueNames[val.Name] = true
	}
	return nil
}

func (s *Schema) validateValueObjectConfig(config *ValueObjectConfig, properties []Property) error {
	// Validate equality members exist
	propNames := make(map[string]bool)
	for _, prop := range properties {
		propNames[prop.Name] = true
	}
	for _, member := range config.EqualityMembers {
		if !propNames[member] {
			return fmt.Errorf("equality member '%s' does not exist in properties", member)
		}
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

	// Allow custom types (might be enums or other entities)
	// Types are validated at compile time, so we accept any type string here

	if prop.IsForeignKey && prop.TargetEntity == "" {
		return fmt.Errorf("foreign key property must specify targetEntity")
	}

	return nil
}

func (s *Schema) validateRelations(entity *Entity, entityNames map[string]bool) error {
	if entity.Relations == nil {
		return nil
	}

	for i, rel := range entity.Relations.OneToOne {
		if rel.TargetEntity == "" {
			return fmt.Errorf("oneToOne[%d]: targetEntity is required", i)
		}
		if rel.NavigationProperty == "" {
			rel.NavigationProperty = rel.TargetEntity
		}
	}

	for i, rel := range entity.Relations.OneToMany {
		if rel.TargetEntity == "" {
			return fmt.Errorf("oneToMany[%d]: targetEntity is required", i)
		}
		// Note: Target entity might not exist yet (forward reference) - this is OK for generation
	}

	for i, rel := range entity.Relations.ManyToOne {
		if rel.TargetEntity == "" {
			return fmt.Errorf("manyToOne[%d]: targetEntity is required", i)
		}
		if rel.NavigationProperty == "" {
			rel.NavigationProperty = rel.TargetEntity
		}
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
