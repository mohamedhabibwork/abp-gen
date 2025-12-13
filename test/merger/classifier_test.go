package merger_test

import (
	"testing"

	"github.com/mohamedhabibwork/abp-gen/internal/merger"
)

func TestClassifier_ClassifyFile(t *testing.T) {
	classifier := merger.NewClassifier()

	tests := []struct {
		name         string
		path         string
		expectedType merger.FileType
	}{
		{
			name:         "Permissions file",
			path:         "/path/to/ProductManagementPermissions.cs",
			expectedType: merger.FileTypePermissions,
		},
		{
			name:         "Permission provider",
			path:         "/path/to/MyModulePermissionDefinitionProvider.cs",
			expectedType: merger.FileTypePermissionProvider,
		},
		{
			name:         "DbContext",
			path:         "/path/to/MyModuleDbContext.cs",
			expectedType: merger.FileTypeDbContext,
		},
		{
			name:         "IDbContext",
			path:         "/path/to/IMyModuleDbContext.cs",
			expectedType: merger.FileTypeIDbContext,
		},
		{
			name:         "Entity",
			path:         "/path/to/Entities/Product.cs",
			expectedType: merger.FileTypeEntity,
		},
		{
			name:         "DTO",
			path:         "/path/to/ProductDto.cs",
			expectedType: merger.FileTypeDTO,
		},
		{
			name:         "Service",
			path:         "/path/to/ProductAppService.cs",
			expectedType: merger.FileTypeService,
		},
		{
			name:         "Manager",
			path:         "/path/to/ProductManager.cs",
			expectedType: merger.FileTypeManager,
		},
		{
			name:         "Controller",
			path:         "/path/to/ProductController.cs",
			expectedType: merger.FileTypeController,
		},
		{
			name:         "Validator",
			path:         "/path/to/CreateProductDtoValidator.cs",
			expectedType: merger.FileTypeValidator,
		},
		{
			name:         "Localization JSON",
			path:         "/path/to/en.json",
			expectedType: merger.FileTypeLocalizationJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.ClassifyFile(tt.path)
			if result != tt.expectedType {
				t.Errorf("ClassifyFile(%s) = %v, want %v", tt.path, result, tt.expectedType)
			}
		})
	}
}

func TestClassifier_GetMergeStrategy(t *testing.T) {
	classifier := merger.NewClassifier()

	tests := []struct {
		name             string
		fileType         merger.FileType
		expectedStrategy merger.MergeStrategy
	}{
		{
			name:             "Permissions uses pattern strategy",
			fileType:         merger.FileTypePermissions,
			expectedStrategy: merger.MergeStrategyPattern,
		},
		{
			name:             "Entity uses AST strategy",
			fileType:         merger.FileTypeEntity,
			expectedStrategy: merger.MergeStrategyAST,
		},
		{
			name:             "JSON uses JSON strategy",
			fileType:         merger.FileTypeLocalizationJSON,
			expectedStrategy: merger.MergeStrategyJSON,
		},
		{
			name:             "Unknown type has no strategy",
			fileType:         merger.FileTypeUnknown,
			expectedStrategy: merger.MergeStrategyNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.GetMergeStrategy(tt.fileType)
			if result != tt.expectedStrategy {
				t.Errorf("GetMergeStrategy(%v) = %v, want %v", tt.fileType, result, tt.expectedStrategy)
			}
		})
	}
}

