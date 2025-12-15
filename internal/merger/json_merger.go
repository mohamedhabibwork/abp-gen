package merger

import (
	"encoding/json"
	"fmt"
)

// JSONMerger handles JSON file merging
type JSONMerger struct {
	strategy string // Merge strategy: "overwrite", "append", "skip"
}

// NewJSONMerger creates a new JSON merger with default strategy
func NewJSONMerger() *JSONMerger {
	return &JSONMerger{
		strategy: "append", // Default strategy
	}
}

// NewJSONMergerWithStrategy creates a new JSON merger with specified strategy
func NewJSONMergerWithStrategy(strategy string) *JSONMerger {
	if strategy == "" {
		strategy = "append"
	}
	return &JSONMerger{
		strategy: strategy,
	}
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

	// Detect conflicts only for "append" strategy
	// For "overwrite" and "skip", conflicts are handled by the strategy itself
	var conflicts []Conflict
	if m.strategy == "append" {
		conflicts = m.detectConflicts(existingData, newData)
	}

	// Merge recursively (always merge, strategy determines how conflicts are handled)
	merged := m.mergeObjects(existingData, newData)

	// Convert back to JSON with indentation
	result, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal merged JSON: %w", err)
	}

	return string(result), conflicts, nil
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

	// Deep copy existing values
	for key, value := range existing {
		result[key] = m.deepCopyValue(value)
	}

	// Merge new values based on strategy
	for key, newValue := range new {
		existingValue, exists := result[key]

		if !exists {
			// Key doesn't exist, add it
			result[key] = m.deepCopyValue(newValue)
		} else {
			// Key exists, apply merge strategy
			switch m.strategy {
			case "overwrite":
				// Overwrite with new value
				result[key] = m.deepCopyValue(newValue)
			case "skip":
				// Keep existing value (already copied)
				// result[key] remains as existingValue
			case "append":
				fallthrough
			default:
				// Try to merge intelligently
				existingMap, existingIsMap := existingValue.(map[string]interface{})
				newMap, newIsMap := newValue.(map[string]interface{})

				if existingIsMap && newIsMap {
					// Both are maps, merge recursively
					result[key] = m.mergeObjects(existingMap, newMap)
				} else {
					// Check if both are arrays
					existingArray, existingIsArray := existingValue.([]interface{})
					newArray, newIsArray := newValue.([]interface{})
					if existingIsArray && newIsArray {
						// Merge arrays
						result[key] = m.MergeArrays(existingArray, newArray)
					} else {
						// Not maps or arrays, keep existing for append strategy
						// result[key] remains as existingValue
					}
				}
			}
		}
	}

	return result
}

// deepCopyValue creates a deep copy of a JSON value
func (m *JSONMerger) deepCopyValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		// Deep copy map
		copied := make(map[string]interface{})
		for k, val := range v {
			copied[k] = m.deepCopyValue(val)
		}
		return copied
	case []interface{}:
		// Deep copy array
		copied := make([]interface{}, len(v))
		for i, val := range v {
			copied[i] = m.deepCopyValue(val)
		}
		return copied
	default:
		// Primitive types can be copied directly
		return value
	}
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
