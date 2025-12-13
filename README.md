# ABP Code Generator (abp-gen)

A cross-platform CLI tool written in Go that generates production-ready ABP Framework (.NET) C# code from JSON schemas or through interactive prompts.

## Features

- ✅ **Full CRUD Generation**: Entities, DTOs, Services, Repositories, Controllers
- ✅ **Smart File Merging**: Intelligent merging of existing files with conflict resolution
- ✅ **Domain Managers**: Business logic encapsulation in domain managers
- ✅ **FluentValidation**: Automatic DTO validation with FluentValidation
- ✅ **Distributed Cache**: Enhanced caching with expiration and list caching
- ✅ **Distributed Event Bus**: Event-driven architecture with distributed events
- ✅ **Relationship Support**: One-to-Many and Many-to-Many relationships
- ✅ **Multi-Database**: Entity Framework Core and MongoDB
- ✅ **Interactive Mode**: Build schemas through guided prompts
- ✅ **Customizable Templates**: Extract and modify embedded templates
- ✅ **Cross-Platform**: Works on Windows, Linux, and macOS
- ✅ **Dry-Run Mode**: Preview changes before applying
- ✅ **Idempotent**: Safe to run multiple times
- ✅ **ABP v9+ Compatible**: Follows latest ABP Framework conventions

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/mohamedhabibwork/abp-gen.git
cd abp-gen

# Build
go build -o abp-gen ./cmd/abp-gen

# Install to PATH (optional)
go install ./cmd/abp-gen
```

### Pre-built Binaries

Download pre-built binaries from the [releases page](https://github.com/mohamedhabibwork/abp-gen/releases).

## Quick Start

### 1. Interactive Mode (No JSON Required)

```bash
abp-gen generate
```

The tool will guide you through:
- Solution configuration
- Entity definitions
- Properties and relationships
- Generation options

### 2. Generate from Schema File

```bash
abp-gen generate --input schema.json
```

### 3. Extract Templates for Customization

```bash
abp-gen init
```

This extracts templates to `./abp-gen-templates/` for customization.

## Usage Examples

### Basic Generation

```bash
# Generate from schema file
abp-gen generate --input examples/schema.json

# Interactive mode
abp-gen generate

# Dry run to preview changes
abp-gen generate --input schema.json --dry-run

# Force overwrite existing files
abp-gen generate --input schema.json --force
```

### Advanced Options

```bash
# Specify solution file
abp-gen generate --input schema.json --solution ./MySolution.sln

# Use custom templates
abp-gen generate --input schema.json --templates ./my-templates

# Verbose output
abp-gen generate --input schema.json --verbose
```

## Schema Format

The schema is a JSON file that describes your entities, relationships, and generation options.

### Basic Structure

```json
{
  "solution": {
    "name": "MyCompany",
    "moduleName": "ProductService",
    "namespaceRoot": "MyCompany.ProductService",
    "abpVersion": "9.0",
    "primaryKeyType": "Guid",
    "dbProvider": "efcore",
    "generateControllers": true
  },
  "entities": [
    {
      "name": "Product",
      "tableName": "Products",
      "entityType": "FullAuditedAggregateRoot",
      "properties": [
        {
          "name": "Name",
          "type": "string",
          "isRequired": true,
          "maxLength": 200,
          "nullable": false
        },
        {
          "name": "Price",
          "type": "decimal",
          "isRequired": true,
          "nullable": false
        }
      ]
    }
  ],
  "options": {
    "useAuditedAggregateRoot": true,
    "useSoftDelete": true,
    "useConcurrencyStamp": true,
    "useExtraProperties": true,
    "useLocalization": true,
    "localizationCultures": ["en", "ar"]
  }
}
```

### Solution Configuration

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `name` | string | Solution name | Required |
| `moduleName` | string | Module/Service name | Required |
| `namespaceRoot` | string | Root namespace | `{name}.{moduleName}` |
| `abpVersion` | string | ABP Framework version | `"9.0"` |
| `primaryKeyType` | string | Primary key type: `Guid`, `long`, or `configurable` | `"Guid"` |
| `dbProvider` | string | Database provider: `efcore`, `mongodb`, or `both` | `"efcore"` |
| `generateControllers` | boolean | Generate HTTP API controllers | `true` |

### Entity Configuration

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Entity name (PascalCase) |
| `tableName` | string | Database table name (auto-pluralized if not provided) |
| `entityType` | string | Entity type: `Entity`, `AggregateRoot`, `FullAuditedAggregateRoot`, `ValueObject` |
| `primaryKeyType` | string | Override solution default (optional) |
| `properties` | array | Entity properties |
| `relations` | object | Entity relationships (optional) |

### Property Configuration

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Property name (PascalCase) |
| `type` | string | C# type: `string`, `int`, `long`, `decimal`, `DateTime`, `bool`, `Guid`, custom |
| `isRequired` | boolean | Is required field |
| `nullable` | boolean | Is nullable |
| `maxLength` | integer | Max length for strings (optional) |
| `defaultValue` | string | Default value (optional) |
| `isForeignKey` | boolean | Is foreign key |
| `targetEntity` | string | Target entity for foreign keys |

### Relationships

#### One-to-Many

```json
{
  "relations": {
    "oneToMany": [
      {
        "targetEntity": "OrderItem",
        "foreignKeyName": "ProductId",
        "navigationProperty": "OrderItems",
        "isCollection": true
      }
    ]
  }
}
```

#### Many-to-Many

```json
{
  "relations": {
    "manyToMany": [
      {
        "targetEntity": "Category",
        "joinEntity": "ProductCategory",
        "navigationProperty": "Categories"
      }
    ]
  }
}
```

### Generation Options

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `useAuditedAggregateRoot` | boolean | Use audited aggregate roots | `true` |
| `useSoftDelete` | boolean | Enable soft delete | `true` |
| `useConcurrencyStamp` | boolean | Enable concurrency stamps | `true` |
| `useExtraProperties` | boolean | Enable extra properties | `true` |
| `useLocalization` | boolean | Enable localization | `true` |
| `localizationCultures` | array | Localization cultures | `["en"]` |
| `validationType` | string | Validation type: `fluentvalidation` or `native` | `"fluentvalidation"` |
| `generateEventHandlers` | boolean | Generate distributed event handlers | `true` |

## Generated Files

For each entity, the generator creates:

### Domain Layer
- `Entities/{EntityName}.cs` - Domain entity
- `Repositories/I{EntityName}Repository.cs` - Repository interface
- `Managers/{EntityName}Manager.cs` - Domain manager for business logic
- `Data/{EntityName}DataSeeder.cs` - Data seeder

### Domain.Shared Layer
- `Constants/{EntityName}Constants.cs` - Entity constants
- `Events/{EntityName}EtoTypes.cs` - Event type constants
- `Events/{EntityName}Eto.cs` - Event Transfer Object

### Application.Contracts Layer
- `{EntityName}/{EntityName}Dto.cs` - Read DTO
- `{EntityName}/Create{EntityName}Dto.cs` - Create DTO
- `{EntityName}/Update{EntityName}Dto.cs` - Update DTO
- `Services/I{EntityName}AppService.cs` - Service interface
- `Permissions/{ModuleName}Permissions.cs` - Permission constants (updated)
- `Permissions/{ModuleName}PermissionDefinitionProvider.cs` - Permission provider (updated)

### Application Layer
- `Services/{EntityName}AppService.cs` - CRUD application service (with distributed cache & event bus)
- `Validators/Create{EntityName}DtoValidator.cs` - FluentValidation validator for Create DTO (if FluentValidation enabled)
- `Validators/Update{EntityName}DtoValidator.cs` - FluentValidation validator for Update DTO (if FluentValidation enabled)
- `EventHandlers/{EntityName}CreatedEventHandler.cs` - Created event handler (if event handlers enabled)
- `EventHandlers/{EntityName}UpdatedEventHandler.cs` - Updated event handler (if event handlers enabled)
- `EventHandlers/{EntityName}DeletedEventHandler.cs` - Deleted event handler (if event handlers enabled)
- `AutoMapper/{EntityName}Profile.cs` - AutoMapper profile

### HttpApi Layer
- `Controllers/{EntityName}Controller.cs` - API controller (if enabled)

### EntityFrameworkCore Layer (if EF Core)
- `EntityFrameworkCore/Configurations/{EntityName}Configuration.cs` - EF Core configuration
- `EntityFrameworkCore/Repositories/EfCore{EntityName}Repository.cs` - Repository implementation
- `EntityFrameworkCore/{ModuleName}DbContext.cs` - DbContext (updated with DbSet)
- `EntityFrameworkCore/I{ModuleName}DbContext.cs` - IDbContext (updated with DbSet)

### MongoDB Layer (if MongoDB)
- `MongoDB/Repositories/Mongo{EntityName}Repository.cs` - MongoDB repository
- `MongoDB/{EntityName}MongoDbConfiguration.cs` - MongoDB configuration

## Key Features Explained

### Domain Managers

Domain managers encapsulate business logic and validation rules. They are used by application services to ensure business rules are enforced consistently.

**Example Usage:**
```csharp
// In Application Service
var entity = await _manager.CreateAsync(id, name, price);
```

### Validation

The generator supports two validation approaches:

**1. FluentValidation (Default)**
- Generates FluentValidation validators for Create and Update DTOs
- Validators are automatically registered by ABP Framework
- Features:
  - Required field validation
  - String length validation
  - Numeric range validation
  - Custom validation rules can be added

**2. Native (Data Annotations)**
- Uses Data Annotations directly on DTOs
  - `[Required]` for required fields
  - `[MaxLength]` for string length validation
- No separate validator classes generated
- Simpler but less flexible than FluentValidation

### Distributed Cache

Enhanced caching with:
- **Single Entity Caching**: Individual entities cached with expiration
- **List Caching**: Full list caching for performance
- **Automatic Cache Invalidation**: Cache cleared on create/update/delete
- **Configurable Expiration**: Cache entries expire after 1 hour (configurable)

### Distributed Event Bus

Events are published via `IDistributedEventBus` for:
- **Microservices Communication**: Events can be consumed by other services
- **Event-Driven Architecture**: Decouple services through events
- **Event Sourcing**: Track all entity changes through events

**Event Types:**
- `Created` - Published when entity is created
- `Updated` - Published when entity is updated
- `Deleted` - Published when entity is deleted

**Event Handlers:**
- `{EntityName}CreatedEventHandler` - Handles Created events
- `{EntityName}UpdatedEventHandler` - Handles Updated events
- `{EntityName}DeletedEventHandler` - Handles Deleted events

Event handlers are automatically registered by ABP Framework and can be used for:
- Cache invalidation
- Sending notifications
- Updating related entities
- Integration with external systems

### Smart File Merging

The generator includes an intelligent file merging system that detects existing files and offers merge options:

**Merge Modes:**
- **Interactive (default)**: Prompts for each existing file
- **Auto-merge**: Automatically merges all files without prompting (`--merge-all`)
- **Force**: Overwrites all files without merging (`--force`)
- **No-merge**: Skips all existing files (`--no-merge`)

**Merge Strategies:**

1. **Pattern-Based** (for simple files):
   - Permission files
   - DbContext files
   - Permission providers
   - Localization JSON

2. **AST-Based** (for complex C# files):
   - Entities (merge properties and methods)
   - DTOs (merge properties)
   - Services (merge methods)
   - Managers (merge methods)
   - Controllers (merge actions)
   - Validators (merge rules)

3. **JSON Merging**:
   - Localization files
   - Configuration files
   - Preserves existing keys and adds new ones

**Conflict Resolution:**

When conflicts are detected, you can:
- Keep existing code
- Use new code
- Keep both (renames new code)
- Skip the conflict
- Apply resolution to all similar conflicts

**Example Workflow:**

```bash
# First generation
abp-gen generate --input schema.json

# Modify schema (add new property)
# Run again with smart merge
abp-gen generate --input schema.json --merge

# Output:
# File exists: Domain/Entities/Product.cs (Entity). What would you like to do?
#   [x] Merge intelligently (recommended)
#   [ ] Overwrite with new content
#   [ ] Skip this file
#   [ ] Show diff first
#
# ✓ Merged Domain/Entities/Product.cs
# Added property: Stock
```

## Template Customization

1. Extract templates:
```bash
abp-gen init
```

2. Customize templates in `./abp-gen-templates/`

3. Use custom templates:
```bash
abp-gen generate --input schema.json --templates ./abp-gen-templates
```

### Available Templates

- `entity.tmpl` - Domain entity
- `repository.tmpl` - Repository interface
- `manager.tmpl` - Domain manager
- `constants.tmpl` - Entity constants
- `event_types.tmpl` - Event types
- `eto.tmpl` - Event Transfer Object
- `seeder.tmpl` - Data seeder
- `create_dto.tmpl` - Create DTO
- `update_dto.tmpl` - Update DTO
- `entity_dto.tmpl` - Read DTO
- `create_validator.tmpl` - FluentValidation validator for Create DTO
- `update_validator.tmpl` - FluentValidation validator for Update DTO
- `event_handler_created.tmpl` - Created event handler
- `event_handler_updated.tmpl` - Updated event handler
- `event_handler_deleted.tmpl` - Deleted event handler
- `app_service_interface.tmpl` - Service interface
- `app_service.tmpl` - Service implementation (with managers, validators, distributed cache & event bus)
- `mapper_profile.tmpl` - AutoMapper profile
- `controller.tmpl` - API controller
- `permissions.tmpl` - Permission constants
- `permission_provider.tmpl` - Permission provider
- `efcore_config.tmpl` - EF Core configuration
- `efcore_repository.tmpl` - EF Core repository
- `mongodb_repository.tmpl` - MongoDB repository
- `mongodb_config.tmpl` - MongoDB configuration

## Build Instructions

### Prerequisites

- Go 1.21 or higher
- (For running generated code) .NET 8.0+ SDK
- ABP Framework 9.0+

### Build from Source

```bash
# Clone repository
git clone https://github.com/mohamedhabibwork/abp-gen.git
cd abp-gen

# Install dependencies
go mod download

# Build
go build -o abp-gen ./cmd/abp-gen

# Run tests
go test ./...
```

### Cross-Platform Builds

Use the provided `.goreleaser.yml`:

```bash
# Install goreleaser
go install github.com/goreleaser/goreleaser@latest

# Build for all platforms
goreleaser build --snapshot --clean

# Create release
goreleaser release --clean
```

This creates binaries for:
- Windows (amd64, arm64)
- Linux (amd64, arm64)
- macOS (amd64, arm64)

## After Generation

After generating code:

1. **Add Database Migration**:
```bash
dotnet ef migrations add Add{EntityName} --project src/YourProject.EntityFrameworkCore
```

2. **Update Database**:
```bash
dotnet ef database update --project src/YourProject.EntityFrameworkCore
```

3. **Build Solution**:
```bash
dotnet build
```

4. **Run Application**:
```bash
dotnet run --project src/YourProject.HttpApi.Host
```

## Examples

See the `examples/` directory for sample schemas:

- `examples/schema.json` - Complete example with relationships

## Troubleshooting

### Solution Not Found

If the tool can't find your solution:
```bash
abp-gen generate --input schema.json --solution ./path/to/YourSolution.sln
```

### Template Errors

Re-extract templates:
```bash
rm -rf ./abp-gen-templates
abp-gen init
```

### Permission Already Exists

The generator is idempotent - it will skip existing permissions. Use `--force` to overwrite all files.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see LICENSE file for details

## Credits

Created by [Your Name]

Inspired by ABP Framework code generation patterns and community needs.

## Support

- GitHub Issues: https://github.com/mohamedhabibwork/abp-gen/issues
- ABP Forum: https://abp.io/support
- Documentation: https://docs.abp.io

## Roadmap

- [ ] Add support for custom repositories with methods
- [ ] Generate integration tests
- [ ] Support for Domain Events
- [ ] Angular/Blazor frontend generation
- [ ] Azure DevOps/GitHub Actions workflows
- [ ] Docker support
- [ ] More relationship types (One-to-One, etc.)
- [ ] Enum generation
- [ ] Value Object generation improvements
- [ ] Multi-tenancy configuration
- [ ] Localization file updates (JSON merging)

## Version History

### v1.0.0 (Current)
- Initial release
- Full CRUD generation
- One-to-Many and Many-to-Many relationships
- Entity Framework Core and MongoDB support
- Interactive mode
- Template customization
- Cross-platform support

