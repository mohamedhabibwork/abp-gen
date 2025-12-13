package merger

import (
	"encoding/json"
	"fmt"
)

// JSONMerger handles JSON file merging
type JSONMerger struct{}

// NewJSONMerger creates a new JSON merger
func NewJSONMerger() *JSONMerger {
	return &JSONMerger{}
}

// Merge merges two JSON files
func (m *JSONMerger) Merge(existing string, newContent string) (string, []Conflict, error) {
	// Parse existing JSON
	var existingData map[string]interface{}
	if err := json.Unmarshal([]byte(existing), &existingData); err != nil {
		return "", nil, fmt.Errorf("failed to parse existing JSON: %w", err)
	}

	// Parse new JSON
	var newData map[string]interface{}
	if err := json.Unmarshal([]byte(newContent), &newData); err != nil {
		return "", nil, fmt.Errorf("failed to parse new JSON: %w", err)
	}

	// Detect conflicts
	conflicts := m.detectConflicts(existingData, newData)

	if len(conflicts) > 0 {
		return existing, conflicts, nil
	}

	// Merge recursively
	merged := m.mergeObjects(existingData, newData)

	// Convert back to JSON with indentation
	result, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal merged JSON: %w", err)
	}

	return string(result), nil, nil
}

// detectConflicts detects conflicts in JSON merging
func (m *JSONMerger) detectConflicts(existing map[string]interface{}, new map[string]interface{}) []Conflict {
	var conflicts []Conflict

	for key, newValue := range new {
		existingValue, exists := existing[key]
		if exists {
			// Check if values are different
			if !m.valuesEqual(existingValue, newValue) {
				conflicts = append(conflicts, Conflict{
					Type:         ConflictTypeDifferentValue,
					Description:  fmt.Sprintf("Key '%s' has different values", key),
					ExistingCode: fmt.Sprintf("%v", existingValue),
					NewCode:      fmt.Sprintf("%v", newValue),
				})
			}
		}
	}

	return conflicts
}

// mergeObjects recursively merges two JSON objects
func (m *JSONMerger) mergeObjects(existing map[string]interface{}, new map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy existing values
	for key, value := range existing {
		result[key] = value
	}

	// Merge new values
	for key, newValue := range new {
		existingValue, exists := result[key]

		if !exists {
			// Key doesn't exist, add it
			result[key] = newValue
		} else {
			// Key exists, try to merge
			existingMap, existingIsMap := existingValue.(map[string]interface{})
			newMap, newIsMap := newValue.(map[string]interface{})

			if existingIsMap && newIsMap {
				// Both are maps, merge recursively
				result[key] = m.mergeObjects(existingMap, newMap)
			} else {
				// Not maps or different types, keep existing
				result[key] = existingValue
			}
		}
	}

	return result
}

// valuesEqual checks if two JSON values are equal
func (m *JSONMerger) valuesEqual(v1 interface{}, v2 interface{}) bool {
	// Simple equality check
	v1JSON, _ := json.Marshal(v1)
	v2JSON, _ := json.Marshal(v2)
	return string(v1JSON) == string(v2JSON)
}

// MergeArrays merges two JSON arrays (used for arrays in JSON)
func (m *JSONMerger) MergeArrays(existing []interface{}, new []interface{}) []interface{} {
	result := make([]interface{}, 0, len(existing)+len(new))

	// Add existing items
	result = append(result, existing...)

	// Add new items that don't exist
	for _, newItem := range new {
		found := false
		for _, existingItem := range existing {
			if m.valuesEqual(existingItem, newItem) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, newItem)
		}
	}

	return result
}
