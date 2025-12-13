# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.1] - 2025-12-14

### Changed
Refactor changelog generation script for improved entry insertion


[0.0.1]: https://github.com/mohamedhabibwork/abp-gen/releases/tag/v0.0.1

### Added
- Smart file merging system with conflict resolution
- Domain managers for business logic encapsulation
- FluentValidation support for DTO validation
- Distributed cache and event bus integration
- Event handlers for distributed events
- Interactive merge prompts
- Pattern-based and AST-based merge strategies
- JSON file merging support
- Automated changelog generation
- Automated documentation version updates

### Changed
- Enhanced application services with managers and validators
- Improved distributed cache with expiration
- Updated event publishing to use distributed event bus

## [1.0.0] - 2025-12-13

### Added
- Initial release of abp-gen
- Cross-platform CLI tool for ABP Framework code generation
- Support for Entity Framework Core and MongoDB
- Full CRUD generation (Entities, DTOs, Services, Repositories, Controllers)
- One-to-Many and Many-to-Many relationship support
- Interactive schema builder
- Template system with customization support
- Permission generation with idempotent updates
- AutoMapper profile generation
- Localization support
- Solution and project structure detection
- Dry-run and force modes
- Comprehensive documentation

### Features
- Generate domain entities with properties and relationships
- Generate DTOs (Create, Update, Entity)
- Generate application services with CRUD operations
- Generate repositories (interface and implementation)
- Generate API controllers
- Generate permission constants and providers
- Generate EF Core configurations and DbContext updates
- Generate MongoDB repositories and configurations
- Generate data seeders
- Generate event types and ETOs
- Generate constants classes

[Unreleased]: https://github.com/mohamedhabibwork/abp-gen/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/mohamedhabibwork/abp-gen/compare/v0.0.0...v1.0.0

