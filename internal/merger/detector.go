package merger

import (
	"fmt"
	"os"
	"path/filepath"
)


// FileExistence represents file existence status
type FileExistence struct {
	Path     string
	Exists   bool
	FileType FileType
	Size     int64
}

// Detector handles file existence detection and merge decisions
type Detector struct {
	classifier *Classifier
}

// NewDetector creates a new merge detector
func NewDetector() *Detector {
	return &Detector{
		classifier: NewClassifier(),
	}
}

// CheckFile checks if a file exists and returns its information
func (d *Detector) CheckFile(path string) (*FileExistence, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return &FileExistence{
			Path:   path,
			Exists: false,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", path, err)
	}

	fileType := d.classifier.ClassifyFile(path)

	return &FileExistence{
		Path:     path,
		Exists:   true,
		FileType: fileType,
		Size:     info.Size(),
	}, nil
}

// NeedsUserDecision determines if a file requires user decision for merging
func (d *Detector) NeedsUserDecision(path string, force bool) (bool, error) {
	if force {
		return false, nil
	}

	fileExists, err := d.CheckFile(path)
	if err != nil {
		return false, err
	}

	return fileExists.Exists, nil
}

// GetRelativePath returns a relative path for display purposes
func (d *Detector) GetRelativePath(path string, basePath string) string {
	rel, err := filepath.Rel(basePath, path)
	if err != nil {
		return path
	}
	return rel
}

// CanMerge determines if a file can be merged based on its type
func (d *Detector) CanMerge(fileType FileType) bool {
	switch fileType {
	case FileTypePermissions,
		FileTypePermissionProvider,
		FileTypeDbContext,
		FileTypeIDbContext,
		FileTypeLocalizationJSON,
		FileTypeEntity,
		FileTypeDTO,
		FileTypeService,
		FileTypeManager,
		FileTypeController,
		FileTypeValidator:
		return true
	default:
		return false
	}
}

// ShouldPromptUser determines if we should prompt the user for a merge decision
func (d *Detector) ShouldPromptUser(path string, force bool, mergeAll bool) (bool, error) {
	if force || mergeAll {
		return false, nil
	}

	fileExists, err := d.CheckFile(path)
	if err != nil {
		return false, err
	}

	if !fileExists.Exists {
		return false, nil
	}

	return d.CanMerge(fileExists.FileType), nil
}

