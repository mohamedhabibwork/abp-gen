package merger

import (
	"path/filepath"
	"strings"
)

// FileType represents the type of file being processed
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypePermissions
	FileTypePermissionProvider
	FileTypeDbContext
	FileTypeIDbContext
	FileTypeEntity
	FileTypeDTO
	FileTypeService
	FileTypeManager
	FileTypeController
	FileTypeValidator
	FileTypeRepository
	FileTypeConstants
	FileTypeLocalizationJSON
	FileTypeAutoMapperProfile
	FileTypeEventHandler
)

// MergeStrategy represents the strategy to use for merging
type MergeStrategy int

const (
	MergeStrategyPattern MergeStrategy = iota
	MergeStrategyAST
	MergeStrategyJSON
	MergeStrategyNone
)

// Classifier classifies files and determines merge strategies
type Classifier struct {
	fileTypeMap map[string]FileType
}

// NewClassifier creates a new file classifier
func NewClassifier() *Classifier {
	return &Classifier{
		fileTypeMap: make(map[string]FileType),
	}
}

// ClassifyFile determines the file type based on path and filename
func (c *Classifier) ClassifyFile(path string) FileType {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)

	// JSON files
	if ext == ".json" {
		return FileTypeLocalizationJSON
	}

	// C# files
	if ext != ".cs" {
		return FileTypeUnknown
	}

	// Permission files
	if strings.HasSuffix(filename, "Permissions.cs") && !strings.Contains(filename, "Provider") {
		return FileTypePermissions
	}

	// Permission provider
	if strings.Contains(filename, "PermissionDefinitionProvider.cs") {
		return FileTypePermissionProvider
	}

	// DbContext files
	if strings.HasSuffix(filename, "DbContext.cs") && !strings.HasPrefix(filename, "I") {
		return FileTypeDbContext
	}

	// IDbContext files
	if strings.HasPrefix(filename, "I") && strings.HasSuffix(filename, "DbContext.cs") {
		return FileTypeIDbContext
	}

	// DTO files
	if strings.Contains(filename, "Dto.cs") {
		return FileTypeDTO
	}

	// Service files
	if strings.HasSuffix(filename, "AppService.cs") {
		return FileTypeService
	}

	// Manager files
	if strings.HasSuffix(filename, "Manager.cs") {
		return FileTypeManager
	}

	// Controller files
	if strings.HasSuffix(filename, "Controller.cs") {
		return FileTypeController
	}

	// Validator files
	if strings.HasSuffix(filename, "Validator.cs") {
		return FileTypeValidator
	}

	// Repository files
	if strings.Contains(filename, "Repository.cs") {
		return FileTypeRepository
	}

	// Constants files
	if strings.HasSuffix(filename, "Constants.cs") {
		return FileTypeConstants
	}

	// AutoMapper profile
	if strings.HasSuffix(filename, "Profile.cs") {
		return FileTypeAutoMapperProfile
	}

	// Event handler files
	if strings.HasSuffix(filename, "EventHandler.cs") {
		return FileTypeEventHandler
	}

	// Entity files (in Entities folder or inheriting from Entity/AggregateRoot)
	if strings.Contains(path, "/Entities/") || strings.Contains(path, "\\Entities\\") {
		return FileTypeEntity
	}

	return FileTypeUnknown
}

// GetMergeStrategy determines the appropriate merge strategy for a file type
func (c *Classifier) GetMergeStrategy(fileType FileType) MergeStrategy {
	switch fileType {
	case FileTypePermissions,
		FileTypePermissionProvider,
		FileTypeDbContext,
		FileTypeIDbContext:
		return MergeStrategyPattern

	case FileTypeEntity,
		FileTypeDTO,
		FileTypeService,
		FileTypeManager,
		FileTypeController,
		FileTypeValidator,
		FileTypeRepository:
		return MergeStrategyAST

	case FileTypeLocalizationJSON:
		return MergeStrategyJSON

	default:
		return MergeStrategyNone
	}
}

// GetFileTypeName returns a human-readable name for a file type
func (c *Classifier) GetFileTypeName(fileType FileType) string {
	switch fileType {
	case FileTypePermissions:
		return "Permissions"
	case FileTypePermissionProvider:
		return "Permission Provider"
	case FileTypeDbContext:
		return "DbContext"
	case FileTypeIDbContext:
		return "IDbContext"
	case FileTypeEntity:
		return "Entity"
	case FileTypeDTO:
		return "DTO"
	case FileTypeService:
		return "Service"
	case FileTypeManager:
		return "Manager"
	case FileTypeController:
		return "Controller"
	case FileTypeValidator:
		return "Validator"
	case FileTypeRepository:
		return "Repository"
	case FileTypeConstants:
		return "Constants"
	case FileTypeLocalizationJSON:
		return "Localization JSON"
	case FileTypeAutoMapperProfile:
		return "AutoMapper Profile"
	case FileTypeEventHandler:
		return "Event Handler"
	default:
		return "Unknown"
	}
}

// IsMergeable determines if a file type can be merged
func (c *Classifier) IsMergeable(fileType FileType) bool {
	strategy := c.GetMergeStrategy(fileType)
	return strategy != MergeStrategyNone
}
