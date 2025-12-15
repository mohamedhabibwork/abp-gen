# abp-gen

[![CI](https://github.com/mohamedhabibwork/abp-gen/workflows/CI/badge.svg)](https://github.com/mohamedhabibwork/abp-gen/actions)
[![Release](https://github.com/mohamedhabibwork/abp-gen/workflows/Release/badge.svg)](https://github.com/mohamedhabibwork/abp-gen/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)

Cross-platform CLI tool for generating ABP Framework C# code from JSON schemas.

## Features

### Core Generation
- ✅ **Full CRUD Generation**: Entities, DTOs, Services, Repositories, Controllers
- ✅ **Custom Repositories**: Define custom repository methods with query hints
- ✅ **Domain Events**: Domain and distributed events with handlers
- ✅ **Enum Generation**: Strongly-typed enums with localization
- ✅ **Value Objects**: Enhanced value object generation with equality
- ✅ **Rich Relationships**: One-to-One, One-to-Many, Many-to-One, Many-to-Many, Self-referencing
- ✅ **Integration Tests**: xUnit/MSTest test generation for ASP.NET Core and ABP

### Smart Detection
- 🔍 **Multi-Format Solutions**: Auto-detect .sln, .slnx, .abpsln, .abpslnx, .csproj
- 🔍 **Framework Detection**: Auto-detect ASP.NET Core 9/10 and ABP 8/9/10
- 🔍 **Multi-Tenancy Detection**: Infer tenancy from configs and module files
- 🔍 **Microservice Detection**: Identify microservice architecture patterns
- 🔍 **CLI Scaffolding**: Create solutions with `abp` or `dotnet` commands

### Advanced Features
- ✅ **Smart File Merging**: Intelligent merging with conflict resolution
- ✅ **Multi-Tenancy**: Support for host, tenant-per-db, tenant-per-schema
- ✅ **Localization Merging**: JSON localization with conflict strategies
- ✅ **Domain Managers**: Business logic encapsulation in domain managers
- ✅ **FluentValidation**: Automatic DTO validation
- ✅ **Multi-Database**: Entity Framework Core and MongoDB
- ✅ **Interactive Mode**: Build schemas through guided prompts
- ✅ **Customizable Templates**: Extract and modify embedded templates
- ✅ **Cross-Platform**: Works on Windows, Linux, and macOS
- ✅ **Dry-Run Mode**: Preview changes before applying

📖 **See [SMART_DETECTION.md](SMART_DETECTION.md) for detailed detection features**

## Installation

### Pre-built Binaries

Download pre-built binaries from the [releases page](https://github.com/mohamedhabibwork/abp-gen/releases).

**Quick Install:**

```bash
# Linux/macOS (replace VERSION with latest version, e.g., 1.0.0)
VERSION="1.0.0"
curl -L https://github.com/mohamedhabibwork/abp-gen/releases/download/v${VERSION}/abp-gen_${VERSION}_linux_amd64.tar.gz | tar xz
sudo mv abp-gen /usr/local/bin/

# macOS (Homebrew) - when tap is available
# brew install mohamedhabibwork/tap/abp-gen

# Windows
# Download abp-gen_${VERSION}_windows_amd64.zip from releases
# Extract and add to PATH
```

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

### GitHub Actions Build Failures

If you see errors like "directory cmd/abp-gen was not found" in GitHub Actions:

1. **Ensure all files are committed and pushed:**
   ```bash
   git add .
   git commit -m "Your commit message"
   git push origin main
   ```

2. **Verify the repository structure:**
   ```bash
   ls -la cmd/abp-gen/
   # Should show main.go
   ```

3. **Check that workflows are in the correct location:**
   ```bash
   ls -la .github/workflows/
   # Should show ci.yml, build.yml, release.yml, etc.
   ```

4. **If the repository is empty on GitHub:**
   - Make sure you've pushed all files, not just the workflows
   - The workflows will fail if they run before the code is pushed

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [ABP Framework](https://abp.io) - The amazing framework this tool generates code for
- [GoReleaser](https://goreleaser.com) - For cross-platform release automation
- All contributors who help improve this project

## Support

- GitHub Issues: https://github.com/mohamedhabibwork/abp-gen/issues
- GitHub Discussions: https://github.com/mohamedhabibwork/abp-gen/discussions
- ABP Forum: https://abp.io/support
- Documentation: https://docs.abp.io

## Security

For security vulnerabilities, please see [SECURITY.md](SECURITY.md).

## Roadmap

### ✅ Completed Features

- [x] Add support for custom repositories with methods
- [x] Generate integration tests (xUnit/MSTest for ASP.NET Core and ABP)
- [x] Support for Domain Events (domain and distributed events with handlers)
- [x] More relationship types (One-to-One, Many-to-One, Self-referencing)
- [x] Enum generation with localization support
- [x] Value Object generation improvements (immutability, equality, factory methods)
- [x] Multi-tenancy configuration (host, tenant-per-db, tenant-per-schema)
- [x] Localization file updates (JSON merging with conflict strategies)
- [x] Smart detection (auto-detect framework, tenancy, microservices)
- [x] CLI scaffolding (create solutions with `abp` or `dotnet` commands)
- [x] CLI flag overrides (override schema values from command line)

### 🚧 Planned Features

- [ ] Angular/Blazor frontend generation
- [ ] Azure DevOps/GitHub Actions workflow templates
- [ ] Docker support and Dockerfile generation
- [ ] GraphQL API generation
- [ ] gRPC service generation
- [ ] Background job generation (Hangfire/Quartz)
- [ ] SignalR hub generation
- [ ] API versioning support
- [ ] Swagger/OpenAPI documentation enhancements
- [ ] Database migration script generation
- [ ] Seed data generation from schema
- [ ] Multi-language template support
- [ ] Plugin system for custom generators
- [ ] Web UI for schema editing
- [ ] Schema validation and linting
- [ ] Import from existing C# code (reverse engineering)

## Version History

### v1.0.0 (Current)
- Initial release
- Full CRUD generation
- One-to-Many and Many-to-Many relationships
- Entity Framework Core and MongoDB support
- Interactive mode
- Template customization
- Cross-platform support

