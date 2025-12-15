package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mohamedhabibwork/abp-gen/internal/detector"
	"github.com/mohamedhabibwork/abp-gen/internal/merger"
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/writer"
)

// LocalizationGenerator generates and merges localization files
type LocalizationGenerator struct {
	writer *writer.Writer
}

// NewLocalizationGenerator creates a new localization generator
func NewLocalizationGenerator(w *writer.Writer) *LocalizationGenerator {
	return &LocalizationGenerator{
		writer: w,
	}
}

// MergeLocalizationFile merges localization content into existing file
func (g *LocalizationGenerator) MergeLocalizationFile(sch *schema.Schema, paths *detector.LayerPaths, culture string, newContent map[string]interface{}) error {
	if sch.Options.LocalizationMerge == nil || !sch.Options.LocalizationMerge.Enabled {
		return nil // Localization merge not enabled
	}

	// Determine target path
	targetPath := sch.Options.LocalizationMerge.TargetPath
	if targetPath == "" {
		targetPath = paths.DomainSharedLocalization
	}

	// Build full file path
	fileName := fmt.Sprintf("%s.json", culture)
	filePath := filepath.Join(targetPath, fileName)

	// Check if file exists
	var existingContent map[string]interface{}
	if fileData, err := os.ReadFile(filePath); err == nil {
		// File exists, parse it
		if err := json.Unmarshal(fileData, &existingContent); err != nil {
			return fmt.Errorf("failed to parse existing localization file: %w", err)
		}
	} else {
		// File doesn't exist, start fresh
		existingContent = make(map[string]interface{})
	}

	// Determine merge strategy
	strategy := sch.Options.LocalizationMerge.ConflictStrategy
	if strategy == "" {
		strategy = "append"
	}

	// Create JSON merger with strategy
	jsonMerger := merger.NewJSONMergerWithStrategy(strategy)

	// Convert to JSON strings for merger
	existingJSON, err := json.MarshalIndent(existingContent, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal existing content: %w", err)
	}

	newJSON, err := json.MarshalIndent(newContent, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal new content: %w", err)
	}

	// Perform merge
	mergedJSON, conflicts, err := jsonMerger.Merge(string(existingJSON), string(newJSON))
	if err != nil {
		return fmt.Errorf("failed to merge localization files: %w", err)
	}

	// Handle conflicts based on strategy
	if len(conflicts) > 0 {
		switch strategy {
		case "overwrite":
			// Conflicts are already resolved by overwriting
		case "skip":
			// Keep existing, no change needed
			return nil
		default:
			// For append, log conflicts but continue
			for _, conflict := range conflicts {
				fmt.Printf("Localization conflict: %s\n", conflict.Description)
			}
		}
	}

	// Write merged content
	return g.writer.WriteFile(filePath, mergedJSON)
}

// GenerateEntityLocalization generates localization entries for an entity
func (g *LocalizationGenerator) GenerateEntityLocalization(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if !sch.Options.UseLocalization {
		return nil
	}

	// Build localization content
	locContent := g.buildEntityLocalizationContent(sch, entity)

	// Merge for each culture
	for _, culture := range sch.Options.LocalizationCultures {
		if err := g.MergeLocalizationFile(sch, paths, culture, locContent); err != nil {
			return fmt.Errorf("failed to merge localization for culture %s: %w", culture, err)
		}
	}

	return nil
}

func (g *LocalizationGenerator) buildEntityLocalizationContent(sch *schema.Schema, entity *schema.Entity) map[string]interface{} {
	content := make(map[string]interface{})

	// Add entity name
	content[entity.Name] = entity.Name

	// Add properties
	for _, prop := range entity.Properties {
		key := fmt.Sprintf("%s.%s", entity.Name, prop.Name)
		content[key] = prop.Name
	}

	// Add permissions
	permissionBase := fmt.Sprintf("Permission:%s", entity.Name)
	content[permissionBase] = entity.Name
	content[permissionBase+".Create"] = fmt.Sprintf("Create %s", entity.Name)
	content[permissionBase+".Edit"] = fmt.Sprintf("Edit %s", entity.Name)
	content[permissionBase+".Delete"] = fmt.Sprintf("Delete %s", entity.Name)

	// Add enum localizations
	for _, enum := range entity.Enums {
		if enum.UseLocalization {
			for _, val := range enum.Values {
				key := val.LocalizationKey
				if key == "" {
					key = fmt.Sprintf("Enum:%s.%s", enum.Name, val.Name)
				}
				content[key] = val.Name
			}
		}
	}

	return content
}
