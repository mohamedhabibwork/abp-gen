package schema

import (
	"encoding/json"
	"os"
)

// Schema represents the complete ABP code generation schema
type Schema struct {
	Solution Solution  `json:"solution"`
	Entities []Entity  `json:"entities"`
	Options  Options   `json:"options"`
}

// Solution represents solution-level configuration
type Solution struct {
	Name                string `json:"name"`
	ModuleName          string `json:"moduleName"`
	NamespaceRoot       string `json:"namespaceRoot"`
	ABPVersion          string `json:"abpVersion"`
	PrimaryKeyType      string `json:"primaryKeyType"`      // "Guid" or "long" or "configurable"
	DBProvider          string `json:"dbProvider"`          // "efcore" or "mongodb" or "both"
	GenerateControllers bool   `json:"generateControllers"`
}

// Entity represents a domain entity
type Entity struct {
	Name           string      `json:"name"`
	TableName      string      `json:"tableName"`
	EntityType     string      `json:"entityType"` // "Entity", "AggregateRoot", "FullAuditedAggregateRoot", "ValueObject"
	PrimaryKeyType string      `json:"primaryKeyType,omitempty"`
	Properties     []Property  `json:"properties"`
	Relations      *Relations  `json:"relations,omitempty"`
}

// Property represents an entity property
type Property struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	IsRequired   bool   `json:"isRequired"`
	MaxLength    int    `json:"maxLength,omitempty"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"defaultValue,omitempty"`
	IsForeignKey bool   `json:"isForeignKey,omitempty"`
	TargetEntity string `json:"targetEntity,omitempty"` // For foreign keys
}

// Relations represents entity relationships
type Relations struct {
	OneToMany  []OneToManyRelation  `json:"oneToMany,omitempty"`
	ManyToMany []ManyToManyRelation `json:"manyToMany,omitempty"`
}

// OneToManyRelation represents a one-to-many relationship
type OneToManyRelation struct {
	TargetEntity       string `json:"targetEntity"`
	ForeignKeyName     string `json:"foreignKeyName"`
	NavigationProperty string `json:"navigationProperty"`
	IsCollection       bool   `json:"isCollection"`
}

// ManyToManyRelation represents a many-to-many relationship
type ManyToManyRelation struct {
	TargetEntity       string `json:"targetEntity"`
	JoinEntity         string `json:"joinEntity"`
	NavigationProperty string `json:"navigationProperty"`
}

// Options represents generation options
type Options struct {
	UseAuditedAggregateRoot bool     `json:"useAuditedAggregateRoot"`
	UseSoftDelete           bool     `json:"useSoftDelete"`
	UseConcurrencyStamp     bool     `json:"useConcurrencyStamp"`
	UseExtraProperties      bool     `json:"useExtraProperties"`
	UseLocalization         bool     `json:"useLocalization"`
	LocalizationCultures    []string `json:"localizationCultures"`
	ValidationType          string   `json:"validationType"` // "fluentvalidation" or "native"
	GenerateEventHandlers   bool     `json:"generateEventHandlers"`
}

// LoadFromFile loads schema from a JSON file
func LoadFromFile(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

// SaveToFile saves schema to a JSON file
func (s *Schema) SaveToFile(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetEffectivePrimaryKeyType returns the effective primary key type for an entity
func (e *Entity) GetEffectivePrimaryKeyType(solutionDefault string) string {
	if e.PrimaryKeyType != "" {
		return e.PrimaryKeyType
	}
	return solutionDefault
}

// GetNonForeignKeyProperties returns properties that are not foreign keys
func (e *Entity) GetNonForeignKeyProperties() []Property {
	var props []Property
	for _, p := range e.Properties {
		if !p.IsForeignKey {
			props = append(props, p)
		}
	}
	return props
}

// GetForeignKeyProperties returns properties that are foreign keys
func (e *Entity) GetForeignKeyProperties() []Property {
	var props []Property
	for _, p := range e.Properties {
		if p.IsForeignKey {
			props = append(props, p)
		}
	}
	return props
}

// HasRelations checks if entity has any relations defined
func (e *Entity) HasRelations() bool {
	return e.Relations != nil && (len(e.Relations.OneToMany) > 0 || len(e.Relations.ManyToMany) > 0)
}

