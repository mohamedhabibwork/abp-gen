package schema

import (
	"encoding/json"
	"os"
)

// Schema represents the complete ABP code generation schema
type Schema struct {
	Solution Solution `json:"solution"`
	Entities []Entity `json:"entities"`
	Options  Options  `json:"options"`
}

// TargetFramework represents the target framework type
type TargetFramework string

const (
	TargetASPNETCore9       TargetFramework = "aspnetcore9"
	TargetASPNETCore10      TargetFramework = "aspnetcore10"
	TargetABP8Microservice  TargetFramework = "abp8-microservice"
	TargetABP8Monolith      TargetFramework = "abp8-monolith"
	TargetABP9Microservice  TargetFramework = "abp9-microservice"
	TargetABP9Monolith      TargetFramework = "abp9-monolith"
	TargetABP10Microservice TargetFramework = "abp10-microservice"
	TargetABP10Monolith     TargetFramework = "abp10-monolith"
	TargetAuto              TargetFramework = "auto" // Auto-detect
)

// Solution represents solution-level configuration
type Solution struct {
	Name                string          `json:"name"`
	ModuleName          string          `json:"moduleName"`
	NamespaceRoot       string          `json:"namespaceRoot"`
	ABPVersion          string          `json:"abpVersion"`
	TargetFramework     TargetFramework `json:"targetFramework"` // Target framework type
	PrimaryKeyType      string          `json:"primaryKeyType"`  // "Guid" or "long" or "configurable"
	DBProvider          string          `json:"dbProvider"`      // "efcore" or "mongodb" or "both"
	GenerateControllers bool            `json:"generateControllers"`
	MultiTenancy        *MultiTenancy   `json:"multiTenancy,omitempty"` // Multi-tenancy configuration
}

// Entity represents a domain entity
type Entity struct {
	Name                     string             `json:"name"`
	TableName                string             `json:"tableName"`
	EntityType               string             `json:"entityType"` // "Entity", "AggregateRoot", "FullAuditedAggregateRoot", "ValueObject"
	PrimaryKeyType           string             `json:"primaryKeyType,omitempty"`
	Properties               []Property         `json:"properties"`
	Relations                *Relations         `json:"relations,omitempty"`
	CustomRepository         *CustomRepository  `json:"customRepository,omitempty"`  // Custom repository methods
	DomainEvents             []DomainEvent      `json:"domainEvents,omitempty"`      // Domain events
	Enums                    []EnumDefinition   `json:"enums,omitempty"`             // Associated enums
	ValueObjectConfig        *ValueObjectConfig `json:"valueObjectConfig,omitempty"` // Value object configuration
	GenerateIntegrationTests bool               `json:"generateIntegrationTests"`    // Generate integration tests
}

// Property represents an entity property
type Property struct {
	Name            string           `json:"name"`
	Type            string           `json:"type"`
	IsRequired      bool             `json:"isRequired"`
	MaxLength       int              `json:"maxLength,omitempty"`
	MinLength       int              `json:"minLength,omitempty"`
	Nullable        bool             `json:"nullable"`
	DefaultValue    string           `json:"defaultValue,omitempty"`
	IsForeignKey    bool             `json:"isForeignKey,omitempty"`
	TargetEntity    string           `json:"targetEntity,omitempty"`    // For foreign keys
	IsEnum          bool             `json:"isEnum,omitempty"`          // Whether this is an enum type
	EnumName        string           `json:"enumName,omitempty"`        // Name of the enum type
	IsValueObject   bool             `json:"isValueObject,omitempty"`   // Whether this is a value object
	ValidationRules []ValidationRule `json:"validationRules,omitempty"` // Custom validation rules
}

// Relations represents entity relationships
type Relations struct {
	OneToOne   []OneToOneRelation   `json:"oneToOne,omitempty"`
	OneToMany  []OneToManyRelation  `json:"oneToMany,omitempty"`
	ManyToOne  []ManyToOneRelation  `json:"manyToOne,omitempty"`
	ManyToMany []ManyToManyRelation `json:"manyToMany,omitempty"`
}

// OneToOneRelation represents a one-to-one relationship
type OneToOneRelation struct {
	TargetEntity       string `json:"targetEntity"`
	ForeignKeyName     string `json:"foreignKeyName"`
	NavigationProperty string `json:"navigationProperty"`
	IsRequired         bool   `json:"isRequired"`    // Whether the relationship is required
	IsOwned            bool   `json:"isOwned"`       // Whether the related entity is owned (EF Core owned type)
	CascadeDelete      bool   `json:"cascadeDelete"` // Whether to cascade delete
}

// ManyToOneRelation represents a many-to-one relationship
type ManyToOneRelation struct {
	TargetEntity       string `json:"targetEntity"`
	ForeignKeyName     string `json:"foreignKeyName"`
	NavigationProperty string `json:"navigationProperty"`
	IsRequired         bool   `json:"isRequired"`
	CascadeDelete      bool   `json:"cascadeDelete"`
}

// OneToManyRelation represents a one-to-many relationship
type OneToManyRelation struct {
	TargetEntity       string `json:"targetEntity"`
	ForeignKeyName     string `json:"foreignKeyName"`
	NavigationProperty string `json:"navigationProperty"`
	IsCollection       bool   `json:"isCollection"`
	CascadeDelete      bool   `json:"cascadeDelete"`
	IsSelfReference    bool   `json:"isSelfReference"` // Self-referencing relationship
}

// ManyToManyRelation represents a many-to-many relationship
type ManyToManyRelation struct {
	TargetEntity       string `json:"targetEntity"`
	JoinEntity         string `json:"joinEntity"`
	NavigationProperty string `json:"navigationProperty"`
	InverseProperty    string `json:"inverseProperty,omitempty"` // Inverse navigation property name
}

// Options represents generation options
type Options struct {
	UseAuditedAggregateRoot  bool               `json:"useAuditedAggregateRoot"`
	UseSoftDelete            bool               `json:"useSoftDelete"`
	UseConcurrencyStamp      bool               `json:"useConcurrencyStamp"`
	UseExtraProperties       bool               `json:"useExtraProperties"`
	UseLocalization          bool               `json:"useLocalization"`
	LocalizationCultures     []string           `json:"localizationCultures"`
	ValidationType           string             `json:"validationType"` // "fluentvalidation" or "native"
	GenerateEventHandlers    bool               `json:"generateEventHandlers"`
	GenerateIntegrationTests bool               `json:"generateIntegrationTests"`
	LocalizationMerge        *LocalizationMerge `json:"localizationMerge,omitempty"` // Localization merging options
}

// LocalizationMerge represents localization file merge configuration
type LocalizationMerge struct {
	Enabled          bool   `json:"enabled"`
	TargetPath       string `json:"targetPath"`       // Target JSON path for merge
	ConflictStrategy string `json:"conflictStrategy"` // "overwrite", "append", "skip"
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
	return e.Relations != nil && (len(e.Relations.OneToOne) > 0 || len(e.Relations.OneToMany) > 0 ||
		len(e.Relations.ManyToOne) > 0 || len(e.Relations.ManyToMany) > 0)
}

// MultiTenancy represents multi-tenancy configuration
type MultiTenancy struct {
	Enabled             bool   `json:"enabled"`
	Strategy            string `json:"strategy"`            // "none", "host", "tenant-per-db", "tenant-per-schema"
	EnableCrossTenant   bool   `json:"enableCrossTenant"`   // Allow cross-tenant queries
	TenantIdProperty    string `json:"tenantIdProperty"`    // Property name for tenant ID (default: "TenantId")
	EnableDataIsolation bool   `json:"enableDataIsolation"` // Enable automatic data isolation
}

// CustomRepository represents custom repository configuration
type CustomRepository struct {
	Methods []RepositoryMethod `json:"methods"` // Custom repository methods
}

// RepositoryMethod represents a custom repository method
type RepositoryMethod struct {
	Name        string            `json:"name"`
	ReturnType  string            `json:"returnType"` // e.g., "Task<List<Product>>", "Task<Product>"
	Parameters  []MethodParameter `json:"parameters"`
	IsAsync     bool              `json:"isAsync"`
	QueryHint   string            `json:"queryHint,omitempty"` // SQL hint or LINQ pattern
	Description string            `json:"description,omitempty"`
}

// MethodParameter represents a method parameter
type MethodParameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// DomainEvent represents a domain event definition
type DomainEvent struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`               // "domain" or "distributed"
	Payload     []EventProperty `json:"payload"`            // Event payload properties
	Handlers    []EventHandler  `json:"handlers,omitempty"` // Event handlers
	Description string          `json:"description,omitempty"`
}

// EventProperty represents a property in an event payload
type EventProperty struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// EventHandler represents an event handler configuration
type EventHandler struct {
	Name        string `json:"name"`
	HandlerType string `json:"handlerType"` // "local", "distributed", "integration"
	Action      string `json:"action"`      // Description of what the handler does
}

// EnumDefinition represents an enum type
type EnumDefinition struct {
	Name            string      `json:"name"`
	UnderlyingType  string      `json:"underlyingType"` // "int", "string", etc.
	Values          []EnumValue `json:"values"`
	UseLocalization bool        `json:"useLocalization"` // Generate localization entries
	GenerateLookup  bool        `json:"generateLookup"`  // Generate lookup/extension methods
	Description     string      `json:"description,omitempty"`
}

// EnumValue represents a single enum value
type EnumValue struct {
	Name            string `json:"name"`
	Value           string `json:"value"` // Can be int or string
	LocalizationKey string `json:"localizationKey,omitempty"`
	Description     string `json:"description,omitempty"`
}

// ValueObjectConfig represents value object configuration
type ValueObjectConfig struct {
	IsImmutable        bool     `json:"isImmutable"`               // Whether the value object is immutable
	FactoryMethod      string   `json:"factoryMethod,omitempty"`   // Factory method name (e.g., "Create")
	ValidationRules    []string `json:"validationRules,omitempty"` // Validation rules to apply
	EqualityMembers    []string `json:"equalityMembers,omitempty"` // Properties to use for equality
	GenerateComparison bool     `json:"generateComparison"`        // Generate comparison operators
}

// ValidationRule represents a custom validation rule
type ValidationRule struct {
	Type         string `json:"type"`  // "Range", "RegularExpression", "Custom", etc.
	Value        string `json:"value"` // The validation value/pattern
	ErrorMessage string `json:"errorMessage,omitempty"`
}
