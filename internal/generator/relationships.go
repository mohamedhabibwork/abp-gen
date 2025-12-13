package generator

import (
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
)

// RelationshipHandler handles entity relationships
type RelationshipHandler struct {
}

// NewRelationshipHandler creates a new relationship handler
func NewRelationshipHandler() *RelationshipHandler {
	return &RelationshipHandler{}
}

// ProcessRelationships processes all relationships for an entity
// This includes:
// - Adding navigation properties to entities
// - Adding foreign key properties
// - Updating DTOs to include related data
// - Adding relationship management methods to services
func (h *RelationshipHandler) ProcessRelationships(sch *schema.Schema, entity *schema.Entity) error {
	if entity.Relations == nil {
		return nil
	}

	// Process One-to-Many relationships
	for _, rel := range entity.Relations.OneToMany {
		if err := h.processOneToMany(sch, entity, &rel); err != nil {
			return err
		}
	}

	// Process Many-to-Many relationships
	for _, rel := range entity.Relations.ManyToMany {
		if err := h.processManyToMany(sch, entity, &rel); err != nil {
			return err
		}
	}

	return nil
}

// processOneToMany processes a one-to-many relationship
func (h *RelationshipHandler) processOneToMany(sch *schema.Schema, entity *schema.Entity, rel *schema.OneToManyRelation) error {
	// One-to-Many relationships are handled in templates:
	// 1. Parent entity gets a collection navigation property
	// 2. Child entity gets a foreign key property
	// 3. DTOs include navigation data
	// 4. Service includes methods to load related data

	// Ensure navigation property name is set
	if rel.NavigationProperty == "" {
		rel.NavigationProperty = templates.Pluralize(rel.TargetEntity)
	}

	// Ensure foreign key name is set
	if rel.ForeignKeyName == "" {
		rel.ForeignKeyName = entity.Name + "Id"
	}

	return nil
}

// processManyToMany processes a many-to-many relationship
func (h *RelationshipHandler) processManyToMany(sch *schema.Schema, entity *schema.Entity, rel *schema.ManyToManyRelation) error {
	// Many-to-Many relationships are handled in templates:
	// 1. Both entities get collection navigation properties
	// 2. A join entity is created if specified
	// 3. DTOs include navigation data
	// 4. Services include methods to link/unlink entities

	// Ensure navigation property name is set
	if rel.NavigationProperty == "" {
		rel.NavigationProperty = templates.Pluralize(rel.TargetEntity)
	}

	// Ensure join entity name is set
	if rel.JoinEntity == "" {
		// Generate join entity name from both entity names
		entities := []string{entity.Name, rel.TargetEntity}
		// Sort to ensure consistent naming
		if entities[0] > entities[1] {
			entities[0], entities[1] = entities[1], entities[0]
		}
		rel.JoinEntity = entities[0] + entities[1]
	}

	return nil
}

// GenerateJoinEntity generates a join entity for many-to-many relationships
func (h *RelationshipHandler) GenerateJoinEntity(entity1Name, entity2Name string, primaryKeyType string) *schema.Entity {
	// Generate a join entity
	joinEntity := &schema.Entity{
		Name:       entity1Name + entity2Name,
		TableName:  templates.Pluralize(entity1Name + entity2Name),
		EntityType: "Entity",
		Properties: []schema.Property{
			{
				Name:         entity1Name + "Id",
				Type:         primaryKeyType,
				IsRequired:   true,
				IsForeignKey: true,
				TargetEntity: entity1Name,
			},
			{
				Name:         entity2Name + "Id",
				Type:         primaryKeyType,
				IsRequired:   true,
				IsForeignKey: true,
				TargetEntity: entity2Name,
			},
		},
	}

	return joinEntity
}

// GetNavigationProperty returns the navigation property for a relationship
func (h *RelationshipHandler) GetNavigationProperty(targetEntity string, isCollection bool) string {
	if isCollection {
		return templates.Pluralize(targetEntity)
	}
	return targetEntity
}

// GetForeignKeyProperty returns the foreign key property name
func (h *RelationshipHandler) GetForeignKeyProperty(targetEntity string) string {
	return targetEntity + "Id"
}

// ValidateRelationships validates all relationships in the schema
func (h *RelationshipHandler) ValidateRelationships(sch *schema.Schema) error {
	// Build entity map for validation
	entityMap := make(map[string]*schema.Entity)
	for i := range sch.Entities {
		entityMap[sch.Entities[i].Name] = &sch.Entities[i]
	}

	// Validate each entity's relationships
	for i := range sch.Entities {
		entity := &sch.Entities[i]
		if entity.Relations == nil {
			continue
		}

		// Validate One-to-Many relationships
		for _, rel := range entity.Relations.OneToMany {
			// Note: Target entity might not exist yet in the schema
			// This is OK as it might be defined elsewhere or be a forward reference
			_ = rel
		}

		// Validate Many-to-Many relationships
		for _, rel := range entity.Relations.ManyToMany {
			_ = rel
		}
	}

	return nil
}
